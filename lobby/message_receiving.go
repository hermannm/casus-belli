package lobby

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"hermannm.dev/bfh-server/game/gametypes"
)

func (player *Player) readMessagesUntilSocketCloses(lobby *Lobby) {
	for {
		socketIsClosed, err := player.readMessage(lobby)

		if socketIsClosed {
			log.Println(fmt.Errorf("socket closed for player '%s': %w", player.String(), err))
			return
		} else if err != nil {
			log.Println(fmt.Errorf("message error for player '%s': %w", player.String(), err))
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
			return false, fmt.Errorf("failed to read message from WebSocket connection: %w", err)
		}
	}

	var messageWithType map[MessageType]json.RawMessage
	if err := json.Unmarshal(messageBytes, &messageWithType); err != nil {
		return false, fmt.Errorf("failed to parse received message: %w", err)
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
		return false, fmt.Errorf("failed to handle message of type '%s': %w", messageType, err)
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
				"received game message of type '%s' before player's game ID was set", messageType,
			)
		}
	}

	return false, nil
}

func (player *Player) handleLobbyMessage(
	messageType MessageType, rawMessage json.RawMessage, lobby *Lobby,
) (isLobbyMessage bool, err error) {
	switch messageType {
	case MessageTypeSelectGameID:
		var message SelectGameIDMessage
		if err := json.Unmarshal(rawMessage, &message); err != nil {
			return true, fmt.Errorf("failed to parse message: %w", err)
		}

		if err := player.selectGameID(message.GameID, lobby); err != nil {
			return true, fmt.Errorf("failed to select game ID: %w", err)
		}

		if err := lobby.SendPlayerStatusMessage(player); err != nil {
			return true, fmt.Errorf(
				"failed to update other players about game ID selection: %w", err,
			)
		}

		return true, nil
	case MessageTypeReady:
		var message ReadyToStartGameMessage
		if err := json.Unmarshal(rawMessage, &message); err != nil {
			return true, fmt.Errorf("failed to parse message: %w", err)
		}

		if err := player.setReadyToStartGame(message.Ready); err != nil {
			return true, fmt.Errorf("failed to set ready status: %w", err)
		}

		if err := lobby.SendPlayerStatusMessage(player); err != nil {
			return true, fmt.Errorf("failed to update other players about ready status: %w", err)
		}

		return true, nil
	case MessageTypeStartGame:
		if err := lobby.startGame(); err != nil {
			return true, fmt.Errorf("failed to start game: %w", err)
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
	supports   SupportMessageQueue
}

func newGameMessageReceiver() GameMessageReceiver {
	return GameMessageReceiver{
		orders:     make(chan SubmitOrdersMessage),
		winterVote: make(chan WinterVoteMessage),
		sword:      make(chan SwordMessage),
		raven:      make(chan RavenMessage),
		supports:   NewSupportMessageQueue(),
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
			player.gameMessageReceiver.supports.AddMessage(message)
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
		err = fmt.Errorf("failed to parse message of type '%s': %w", messageType, err)
		log.Println(fmt.Errorf("message error for player '%s': %w", player.String(), err))
		player.SendError(err)
	}
}

func (lobby *Lobby) AwaitOrders(fromPlayer string) ([]gametypes.Order, error) {
	player, ok := lobby.getPlayer(fromPlayer)
	if !ok {
		return nil, fmt.Errorf(
			"failed to get order message from player '%s': player not found", fromPlayer,
		)
	}

	orders := <-player.gameMessageReceiver.orders
	return orders.Orders, nil
}

func (lobby *Lobby) AwaitSupport(
	fromPlayer string, supportingRegion string, embattledRegion string,
) (supportedPlayer string, err error) {
	player, ok := lobby.getPlayer(fromPlayer)
	if !ok {
		return "", fmt.Errorf(
			"failed to get support message from player '%s' in region '%s': player not found",
			fromPlayer, supportingRegion,
		)
	}

	supportedPlayer = player.gameMessageReceiver.supports.AwaitSupportMatchingRegions(
		supportingRegion, embattledRegion,
	)

	return supportedPlayer, nil
}

type SupportMessageQueue struct {
	messages           []GiveSupportMessage // Must hold newMessageNotifier.L to access safely.
	newMessageNotifier sync.Cond
}

func NewSupportMessageQueue() SupportMessageQueue {
	return SupportMessageQueue{messages: nil, newMessageNotifier: sync.Cond{L: &sync.Mutex{}}}
}

func (queue *SupportMessageQueue) AddMessage(message GiveSupportMessage) {
	queue.newMessageNotifier.L.Lock()
	queue.messages = append(queue.messages, message)
	queue.newMessageNotifier.L.Unlock()
	queue.newMessageNotifier.Broadcast()
}

// Reading a received support message involves these steps:
//  1. Acquire the lock on newMessageNotifier condition variable
//  2. Go through received support messages, and check if any match the requested supporting and
//     embattled regions
//  3. If a match was found: take message from queue, release lock and return supported player
//  4. If not: call newMessageNotifier.Wait(), which releases the lock and waits
//  5. Once a new support message is received, SupportMessageQueue.AddMessage will call
//     newMessageNotifier.Broadcast(), which wakes all waiting goroutines
//  6. Once Wait() returns in this goroutine, the lock is re-acquired, and we repeat from step 2
//
// For more info, see the docs on sync.Cond: https://pkg.go.dev/sync#Cond
func (queue *SupportMessageQueue) AwaitSupportMatchingRegions(
	supportingRegion string, embattledRegion string,
) (supportedPlayer string) {
	queue.newMessageNotifier.L.Lock()
	for {
		foundMatchingSupport := false
		remainingMessages := make([]GiveSupportMessage, 0, cap(queue.messages))

		for _, message := range queue.messages {
			foundMatchingSupport = message.SupportingRegion == supportingRegion &&
				message.EmbattledRegion == embattledRegion

			if foundMatchingSupport {
				if message.SupportedPlayer != nil {
					supportedPlayer = *message.SupportedPlayer
				}
			} else {
				remainingMessages = append(remainingMessages, message)
			}
		}

		if foundMatchingSupport {
			queue.messages = remainingMessages
			queue.newMessageNotifier.L.Unlock()
			return supportedPlayer
		}

		queue.newMessageNotifier.Wait()
	}
}
