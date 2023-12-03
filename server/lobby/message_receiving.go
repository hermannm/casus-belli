package lobby

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
	"hermannm.dev/casus-belli/server/game"
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

	lobby.lock.RLock()
	gameStarted := lobby.gameStarted
	lobby.lock.RUnlock()

	if gameStarted {
		err = player.handleGameMessage(message.Tag, message.Data, lobby)
	} else {
		err = player.handleLobbyMessage(message.Tag, message.Data, lobby)
	}
	if err != nil {
		return false, wrap.Errorf(err, "failed to handle message of type '%s'", message.Tag)
	}

	return false, nil
}

func (player *Player) handleLobbyMessage(
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

		lobby.SendPlayerStatusMessage(player)
	case MessageTagStartGame:
		if err := lobby.startGame(); err != nil {
			return wrap.Error(err, "failed to start game")
		}
	default:
		return fmt.Errorf("invalid lobby message tag '%s'", messageTag)
	}

	return nil
}

func (player *Player) handleGameMessage(
	messageTag MessageTag,
	rawMessage json.RawMessage,
	lobby *Lobby,
) error {
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
		return fmt.Errorf("invalid game message tag '%s'", messageTag)
	}

	lobby.gameMessageQueue.Add(ReceivedMessage{
		Tag:          messageTag,
		Data:         messageData,
		ReceivedFrom: player.gameFaction,
	})
	return nil
}

func (lobby *Lobby) AwaitOrders(
	ctx context.Context,
	from game.PlayerFaction,
) ([]game.Order, error) {
	message, err := lobby.gameMessageQueue.AwaitMatchingItem(
		ctx,
		func(message ReceivedMessage) bool {
			return message.ReceivedFrom == from && message.Tag == MessageTagSubmitOrders
		},
	)
	if err != nil {
		return nil, err
	}

	messageData, ok := message.Data.(SubmitOrdersMessage)
	if !ok {
		return nil, errors.New("failed to cast received message to SubmitOrdersMessage")
	}

	return messageData.Orders, nil
}

func (lobby *Lobby) AwaitSupport(
	ctx context.Context,
	from game.PlayerFaction,
	embattled game.RegionName,
) (supported game.PlayerFaction, err error) {
	ctx, cancel := context.WithCancelCause(ctx)

	message, err := lobby.gameMessageQueue.AwaitMatchingItem(
		ctx,
		func(message ReceivedMessage) bool {
			if message.ReceivedFrom != from || message.Tag != MessageTagGiveSupport {
				return false
			}

			messageData, ok := message.Data.(GiveSupportMessage)
			if !ok {
				cancel(errors.New("failed to cast received message to GiveSupportMessage"))
				return false
			}

			return messageData.EmbattledRegion == embattled
		},
	)
	if err != nil {
		return "", err
	}

	messageData := message.Data.(GiveSupportMessage) // Already checked inside AwaitMatchingItem
	return messageData.SupportedFaction, nil
}

func (lobby *Lobby) AwaitDiceRoll(ctx context.Context, from game.PlayerFaction) error {
	_, err := lobby.gameMessageQueue.AwaitMatchingItem(ctx, func(message ReceivedMessage) bool {
		return message.ReceivedFrom == from && message.Tag == MessageTagDiceRoll
	})
	return err
}
