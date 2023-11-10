package lobby

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
	"hermannm.dev/bfh-server/game"
	"hermannm.dev/condqueue"
	"hermannm.dev/devlog/log"
	"hermannm.dev/wrap"
)

func (player *Player) readMessagesUntilSocketCloses(lobby *Lobby) {
	for {
		socketIsClosed, err := player.readMessage(lobby)

		if socketIsClosed {
			log.Errorf(
				err,
				"socket closed for player %s, removing them from lobby",
				player.String(),
			)
			lobby.RemovePlayer(player.username)
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

	var message struct {
		Tag  MessageTag
		Data json.RawMessage
	}
	if err := json.Unmarshal(messageBytes, &message); err != nil {
		return false, wrap.Error(err, "failed to parse received message")
	}

	isLobbyMessage, err := player.handleLobbyMessage(message.Tag, message.Data, lobby)
	if err != nil {
		return false, wrap.Errorf(err, "failed to handle message of type '%s'", message.Tag)
	}

	if !isLobbyMessage {
		player.lock.RLock()
		hasFaction := player.gameFaction != ""
		player.lock.RUnlock()

		if hasFaction {
			// Launch in new goroutine, so it can send on channels while this goroutine keeps
			// reading messages
			go player.handleGameMessage(message.Tag, message.Data)
		} else {
			return false, fmt.Errorf(
				"received game message of type '%s' before player's game ID was set",
				message.Tag,
			)
		}
	}

	return false, nil
}

func (player *Player) handleLobbyMessage(
	messageTag MessageTag,
	rawMessage json.RawMessage,
	lobby *Lobby,
) (isLobbyMessage bool, err error) {
	switch messageTag {
	case MessageTagSelectFaction:
		var message SelectFactionMessage
		if err := json.Unmarshal(rawMessage, &message); err != nil {
			return true, wrap.Error(err, "failed to parse message")
		}

		if err := player.selectFaction(message.Faction, lobby); err != nil {
			return true, wrap.Error(err, "failed to select game ID")
		}

		if err := lobby.SendPlayerStatusMessage(player); err != nil {
			return true, wrap.Error(err, "failed to update other players about game ID selection")
		}

		return true, nil
	case MessageTagReady:
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
	case MessageTagStartGame:
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

func (player *Player) handleGameMessage(tag MessageTag, rawMessage json.RawMessage) {
	var err error // Error declared here in order to handle it after the switch

	switch tag {
	case MessageTagSubmitOrders:
		var message SubmitOrdersMessage
		if err = json.Unmarshal(rawMessage, &message); err == nil {
			player.gameMessageReceiver.orders <- message
		}
	case MessageTagGiveSupport:
		var message GiveSupportMessage
		if err = json.Unmarshal(rawMessage, &message); err == nil {
			player.gameMessageReceiver.supports.Add(message)
		}
	case MessageTagWinterVote:
		var message WinterVoteMessage
		if err = json.Unmarshal(rawMessage, &message); err == nil {
			player.gameMessageReceiver.winterVote <- message
		}
	case MessageTagSword:
		var message SwordMessage
		if err = json.Unmarshal(rawMessage, &message); err == nil {
			player.gameMessageReceiver.sword <- message
		}
	case MessageTagRaven:
		var message RavenMessage
		if err = json.Unmarshal(rawMessage, &message); err == nil {
			player.gameMessageReceiver.raven <- message
		}
	default:
		err = errors.New("unrecognized message type")
	}

	if err != nil {
		err = wrap.Errorf(err, "failed to parse message of type '%s'", tag)
		log.Errorf(err, "message error for player %s", player.String())
		player.SendError(err)
	}
}

func (lobby *Lobby) AwaitOrders(from game.PlayerFaction) ([]game.Order, error) {
	player, ok := lobby.getPlayer(from)
	if !ok {
		return nil, fmt.Errorf(
			"failed to get order message from player '%s': player not found",
			from,
		)
	}

	orders := <-player.gameMessageReceiver.orders
	return orders.Orders, nil
}

func (lobby *Lobby) AwaitSupport(
	from game.PlayerFaction,
	supportingRegion string,
	embattledRegion string,
) (supported game.PlayerFaction, err error) {
	player, ok := lobby.getPlayer(from)
	if !ok {
		return "", fmt.Errorf(
			"failed to get support message from player '%s' in region '%s': player not found",
			from,
			supportingRegion,
		)
	}

	supportMessage, err := player.gameMessageReceiver.supports.AwaitMatchingItem(
		context.Background(), // TODO: implement timeout/cancellation
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

	if supportMessage.SupportedFaction != nil {
		supported = *supportMessage.SupportedFaction
	}

	return supported, nil
}
