package lobby

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
	"hermannm.dev/bfh-server/game"
	"hermannm.dev/wrap"
)

func (player *Player) readMessagesUntilSocketCloses(lobby *Lobby) {
	for {
		socketClosed, err := player.readMessage(lobby)
		if socketClosed {
			player.log.ErrorCause(err, "socket closed, removing from lobby")
			lobby.RemovePlayer(player.username)
			return
		} else if err != nil {
			player.log.Error(err)
			player.SendError(err)
		}
	}
}

func (player *Player) readMessage(lobby *Lobby) (socketClosed bool, err error) {
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

	if err := player.handleMessage(message.Tag, message.Data, lobby); err != nil {
		return false, wrap.Errorf(err, "failed to handle message of type '%s'", message.Tag)
	}

	return false, nil
}

func (player *Player) handleMessage(
	messageTag MessageTag,
	rawMessage json.RawMessage,
	lobby *Lobby,
) error {
	switch messageTag {
	case MessageTagSelectFaction:
		var message SelectFactionMessage
		if err := json.Unmarshal(rawMessage, &message); err != nil {
			return wrap.Error(err, "failed to parse message")
		}

		if err := player.selectFaction(message.Faction, lobby); err != nil {
			return wrap.Error(err, "failed to select faction")
		}

		if err := lobby.SendPlayerStatusMessage(player); err != nil {
			return wrap.Error(err, "failed to update other players about faction selection")
		}
	case MessageTagReady:
		var message ReadyToStartGameMessage
		if err := json.Unmarshal(rawMessage, &message); err != nil {
			return wrap.Error(err, "failed to parse message")
		}

		if err := player.setReadyToStartGame(message.Ready); err != nil {
			return wrap.Error(err, "failed to set ready status")
		}

		if err := lobby.SendPlayerStatusMessage(player); err != nil {
			return wrap.Error(err, "failed to update other players about ready status")
		}
	case MessageTagStartGame:
		if err := lobby.startGame(); err != nil {
			return wrap.Error(err, "failed to start game")
		}
	default: // The message should now be a game message
		if err := player.handleGameMessage(messageTag, rawMessage, lobby); err != nil {
			return err
		}
	}

	return nil
}

func (player *Player) handleGameMessage(
	messageTag MessageTag,
	rawMessage json.RawMessage,
	lobby *Lobby,
) error {
	player.lock.RLock()
	playerFaction := player.gameFaction
	player.lock.RUnlock()

	if playerFaction == "" {
		return errors.New("received game message before player selected faction")
	}

	var messageData any
	switch messageTag {
	case MessageTagSubmitOrders:
		var message SubmitOrdersMessage
		if err := json.Unmarshal(rawMessage, &message); err != nil {
			return wrap.Error(err, "failed to parse message")
		}
		messageData = message
	case MessageTagGiveSupport:
		var message GiveSupportMessage
		if err := json.Unmarshal(rawMessage, &message); err != nil {
			return wrap.Error(err, "failed to parse message")
		}
		messageData = message
	default:
		return fmt.Errorf("unrecognized message tag '%d'", messageTag)
	}

	lobby.receivedMessages.Add(
		ReceivedMessage{Tag: messageTag, Data: messageData, ReceivedFrom: playerFaction},
	)
	return nil
}

func (lobby *Lobby) AwaitOrders(from game.PlayerFaction) ([]game.Order, error) {
	message, err := lobby.receivedMessages.AwaitMatchingItem(
		context.Background(),
		func(message ReceivedMessage) bool {
			return message.ReceivedFrom == from && message.Tag == MessageTagSubmitOrders
		},
	)
	if err != nil {
		return nil, err
	}

	messageData, ok := message.Data.(SubmitOrdersMessage)
	if !ok {
		return nil, errors.New("failed to cast message to SubmitOrdersMessage")
	}

	return messageData.Orders, nil
}

func (lobby *Lobby) AwaitSupport(
	from game.PlayerFaction,
	supporting game.RegionName,
	embattled game.RegionName,
) (supported game.PlayerFaction, err error) {
	ctx, cancel := context.WithCancelCause(context.Background())

	message, err := lobby.receivedMessages.AwaitMatchingItem(
		ctx,
		func(message ReceivedMessage) bool {
			if message.ReceivedFrom != from || message.Tag != MessageTagGiveSupport {
				return false
			}

			messageData, ok := message.Data.(GiveSupportMessage)
			if !ok {
				cancel(errors.New("failed to cast message to GiveSupportMessage"))
				return false
			}

			return messageData.SupportingRegion == supporting &&
				messageData.EmbattledRegion == embattled
		},
	)
	if err != nil {
		return "", wrap.Errorf(
			err,
			"received no support message from region '%s' to region '%s'",
			supporting,
			embattled,
		)
	}

	messageData := message.Data.(GiveSupportMessage) // Already checked inside AwaitMatchingItem
	if messageData.SupportedFaction != nil {
		supported = *messageData.SupportedFaction
	}

	return supported, nil
}
