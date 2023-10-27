package lobby

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
	"hermannm.dev/bfh-server/game/gametypes"
	"hermannm.dev/condqueue"
	"hermannm.dev/devlog/log"
	"hermannm.dev/wrap"
)

func (player *Player) readMessagesUntilSocketCloses(lobby *Lobby) {
	for {
		socketIsClosed, err := player.readMessage(lobby)

		if socketIsClosed {
			log.Errorf(err, "socket closed for player %s", player.String())
			return
		} else if err != nil {
			log.Errorf(err, "message error for player %s", player.String())
			player.SendError(err)
		}
	}
}

func (player *Player) readMessage(lobby *Lobby) (socketIsClosed bool, err error) {
	// Reads from socket without holding lock, as this should be the only goroutine calling this.
	// Websocket supports 1 concurrent reader and 1 concurrent writer, so this should be safe.
	// See https://pkg.go.dev/github.com/gorilla/websocket#hdr-Concurrency
	_, messageBytes, err := player.socket.ReadMessage()
	if err != nil {
		switch err.(type) {
		case *websocket.CloseError:
			return true, err
		default:
			return false, wrap.Error(err, "failed to read message from WebSocket connection")
		}
	}

	var messageWithType map[MessageType]json.RawMessage
	if err := json.Unmarshal(messageBytes, &messageWithType); err != nil {
		return false, wrap.Error(err, "failed to parse received message")
	}

	if len(messageWithType) != 1 {
		return false, errors.New("failed to parse received message: invalid message format")
	}

	var messageType MessageType
	var rawMessage json.RawMessage
	for messageType, rawMessage = range messageWithType {
		break
	}

	isLobbyMessage, err := player.handleLobbyMessage(messageType, rawMessage, lobby)
	if err != nil {
		return false, wrap.Errorf(err, "failed to handle message of type '%s'", messageType)
	}

	if !isLobbyMessage {
		player.lock.RLock()
		hasGameID := player.gameID != ""
		player.lock.RUnlock()

		if hasGameID {
			// Launch in new goroutine, so it can send on channels while this goroutine keeps
			// reading messages
			go player.handleGameMessage(messageType, rawMessage)
		} else {
			return false, fmt.Errorf(
				"received game message of type '%s' before player's game ID was set",
				messageType,
			)
		}
	}

	return false, nil
}

func (player *Player) handleLobbyMessage(
	messageType MessageType,
	rawMessage json.RawMessage,
	lobby *Lobby,
) (isLobbyMessage bool, err error) {
	switch messageType {
	case MessageTypeSelectGameID:
		var message SelectGameIDMessage
		if err := json.Unmarshal(rawMessage, &message); err != nil {
			return true, wrap.Error(err, "failed to parse message")
		}

		if err := player.selectGameID(message.GameID, lobby); err != nil {
			return true, wrap.Error(err, "failed to select game ID")
		}

		if err := lobby.SendPlayerStatusMessage(player); err != nil {
			return true, wrap.Error(err, "failed to update other players about game ID selection")
		}

		return true, nil
	case MessageTypeReady:
		var message ReadyToStartGameMessage
		if err := json.Unmarshal(rawMessage, &message); err != nil {
			return true, wrap.Error(err, "failed to parse message")
		}

		if err := player.setReadyToStartGame(message.Ready); err != nil {
			return true, wrap.Error(err, "failed to set ready status")
		}

		if err := lobby.SendPlayerStatusMessage(player); err != nil {
			return true, wrap.Error(err, "failed to update other players about ready status")
		}

		return true, nil
	case MessageTypeStartGame:
		if err := lobby.startGame(); err != nil {
			return true, wrap.Error(err, "failed to start game")
		}

		return true, nil
	default:
		return false, nil
	}
}

type GameMessageReceiver struct {
	orders     chan SubmitOrdersMessage
	winterVote chan WinterVoteMessage
	sword      chan SwordMessage
	raven      chan RavenMessage
	supports   *condqueue.CondQueue[GiveSupportMessage]
}

func newGameMessageReceiver() GameMessageReceiver {
	return GameMessageReceiver{
		orders:     make(chan SubmitOrdersMessage),
		winterVote: make(chan WinterVoteMessage),
		sword:      make(chan SwordMessage),
		raven:      make(chan RavenMessage),
		supports:   condqueue.New[GiveSupportMessage](),
	}
}

func (player *Player) handleGameMessage(messageType MessageType, rawMessage json.RawMessage) {
	var err error // Error declared here in order to handle it after the switch

	switch messageType {
	case MessageTypeSubmitOrders:
		var message SubmitOrdersMessage
		if err = json.Unmarshal(rawMessage, &message); err == nil {
			player.gameMessageReceiver.orders <- message
		}
	case MessageTypeGiveSupport:
		var message GiveSupportMessage
		if err = json.Unmarshal(rawMessage, &message); err == nil {
			player.gameMessageReceiver.supports.AddItem(context.Background(), message)
		}
	case MessageTypeWinterVote:
		var message WinterVoteMessage
		if err = json.Unmarshal(rawMessage, &message); err == nil {
			player.gameMessageReceiver.winterVote <- message
		}
	case MessageTypeSword:
		var message SwordMessage
		if err = json.Unmarshal(rawMessage, &message); err == nil {
			player.gameMessageReceiver.sword <- message
		}
	case MessageTypeRaven:
		var message RavenMessage
		if err = json.Unmarshal(rawMessage, &message); err == nil {
			player.gameMessageReceiver.raven <- message
		}
	default:
		err = errors.New("unrecognized message type")
	}

	if err != nil {
		err = wrap.Errorf(err, "failed to parse message of type '%s'", messageType)
		log.Errorf(err, "message error for player %s", player.String())
		player.SendError(err)
	}
}

func (lobby *Lobby) AwaitOrders(fromPlayer string) ([]gametypes.Order, error) {
	player, ok := lobby.getPlayer(fromPlayer)
	if !ok {
		return nil, fmt.Errorf(
			"failed to get order message from player '%s': player not found",
			fromPlayer,
		)
	}

	orders := <-player.gameMessageReceiver.orders
	return orders.Orders, nil
}

func (lobby *Lobby) AwaitSupport(
	fromPlayer string,
	supportingRegion string,
	embattledRegion string,
) (supportedPlayer string, err error) {
	player, ok := lobby.getPlayer(fromPlayer)
	if !ok {
		return "", fmt.Errorf(
			"failed to get support message from player '%s' in region '%s': player not found",
			fromPlayer,
			supportingRegion,
		)
	}

	supportMessage, err := player.gameMessageReceiver.supports.AwaitMatchingItem(
		context.Background(),
		func(candidate GiveSupportMessage) bool {
			return candidate.SupportingRegion == supportingRegion &&
				candidate.EmbattledRegion == embattledRegion
		},
	)
	if err != nil {
		return "", wrap.Errorf(
			err,
			"received no support message from region '%s' to region '%s'",
			supportingRegion,
			embattledRegion,
		)
	}

	if supportMessage.SupportedPlayer != nil {
		supportedPlayer = *supportMessage.SupportedPlayer
	}

	return supportedPlayer, nil
}
