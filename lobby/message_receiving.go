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
	orders          chan SubmitOrdersMessage
	winterVote      chan WinterVoteMessage
	sword           chan SwordMessage
	raven           chan RavenMessage
	supports        []GiveSupportMessage // Must hold supportNotifier.L to access safely.
	supportNotifier sync.Cond
}

func newGameMessageReceiver() GameMessageReceiver {
	return GameMessageReceiver{
		orders:          make(chan SubmitOrdersMessage),
		winterVote:      make(chan WinterVoteMessage),
		sword:           make(chan SwordMessage),
		raven:           make(chan RavenMessage),
		supports:        nil,
		supportNotifier: sync.Cond{L: &sync.Mutex{}},
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
			receiver := &player.gameMessageReceiver
			receiver.supportNotifier.L.Lock()
			receiver.supports = append(receiver.supports, message)
			receiver.supportNotifier.L.Unlock()
			receiver.supportNotifier.Broadcast()
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

	receiver := &player.gameMessageReceiver

	// Reading a received support message involves these steps:
	// 1. Acquire the lock on supportNotifier
	// 2. Go through received support messages, and check if any match the requested supporting and
	//    embattled regions
	// 3. If a match was found: take message from queue, release lock and return supported player
	// 4. If not: call supportNotifier.Wait(), which releases the lock and waits
	// 5. Once a new support message is received, Player.handleGameMessage will call
	//    supportNotifier.Broadcast(), which wakes all waiting goroutines
	// 6. Once Wait() returns in this goroutine, the lock is re-acquired, and we repeat from step 2
	// For more info, see the docs on sync.Cond: https://pkg.go.dev/sync#Cond
	receiver.supportNotifier.L.Lock()
	for {
		var supportedPlayer string
		foundMatchingSupport := false
		remainingSupports := make([]GiveSupportMessage, 0, cap(receiver.supports))

		for _, support := range receiver.supports {
			if support.SupportingRegion == supportingRegion &&
				support.EmbattledRegion == embattledRegion {

				foundMatchingSupport = true
				if support.SupportedPlayer != nil {
					supportedPlayer = *support.SupportedPlayer
				}
			} else {
				remainingSupports = append(remainingSupports, support)
			}
		}

		if foundMatchingSupport {
			receiver.supports = remainingSupports
			receiver.supportNotifier.L.Unlock()
			return supportedPlayer, nil
		}

		receiver.supportNotifier.Wait()
	}
}
