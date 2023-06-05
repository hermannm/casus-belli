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

// Continuously listens for messages from the player's socket until it is closed.
func (player *Player) listen(lobby *Lobby) {
	for {
		socketClosed, err := player.receiveMessage(lobby)
		if socketClosed {
			if err != nil {
				log.Println(fmt.Errorf("socket closed for player '%s': %w", player.String(), err))
			}
			break
		}
		if err != nil {
			log.Println(fmt.Errorf("message error for player '%s': %w", player.String(), err))
			player.SendError(err)
		}
	}
}

// Reads a message from the player's socket, and handles it appropriately.
func (player *Player) receiveMessage(lobby *Lobby) (socketClosed bool, err error) {
	player.lock.RLock()
	defer player.lock.RUnlock()

	_, receivedMsg, err := player.socket.ReadMessage()
	if err != nil {
		switch err := err.(type) {
		case *websocket.CloseError:
			return true, err
		default:
			return false, fmt.Errorf("socket connection error: %w", err)
		}
	}

	var messageWithType map[MessageType]json.RawMessage
	if err := json.Unmarshal(receivedMsg, &messageWithType); err != nil {
		return false, fmt.Errorf("failed to parse message: %w", err)
	}
	if len(messageWithType) != 1 {
		return false, errors.New("invalid message format")
	}

	var messageType MessageType
	var rawMsg json.RawMessage
	for messageType, rawMsg = range messageWithType {
		break
	}

	isLobbyMessage, err := player.receiveLobbyMessage(lobby, messageType, rawMsg)
	if err != nil {
		return false, fmt.Errorf("error in handling lobby message: %w", err)
	}

	if !isLobbyMessage {
		if player.gameID == "" {
			return false, errors.New("received game message before the player's game ID was set")
		} else {
			go player.gameMessageReceiver.receiveGameMessage(messageType, rawMsg)
		}
	}

	return false, nil
}

func (player *Player) receiveLobbyMessage(
	lobby *Lobby, messageType MessageType, rawMessage json.RawMessage,
) (isLobbyMsg bool, err error) {
	switch messageType {
	case messageTypeSelectGameID:
		var message SelectGameIDMessage
		if err := json.Unmarshal(rawMessage, &message); err != nil {
			return true, fmt.Errorf("failed to unmarshal %s message: %w", messageType, err)
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
	case messageTypeReady:
		var message ReadyToStartGameMessage
		if err := json.Unmarshal(rawMessage, &message); err != nil {
			return true, fmt.Errorf("failed to unmarshal %s message: %w", messageType, err)
		}

		if err := player.setReadyToStartGame(message.Ready); err != nil {
			return true, fmt.Errorf("failed to set ready status: %w", err)
		}

		if err := lobby.SendPlayerStatusMessage(player); err != nil {
			return true, fmt.Errorf("failed to update other players about ready status: %w", err)
		}

		return true, nil
	case messageTypeStartGame:
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

func newGameMessageReceiver() *GameMessageReceiver {
	return &GameMessageReceiver{
		orders:          make(chan SubmitOrdersMessage),
		winterVote:      make(chan WinterVoteMessage),
		sword:           make(chan SwordMessage),
		raven:           make(chan RavenMessage),
		supports:        nil,
		supportNotifier: sync.Cond{L: &sync.Mutex{}},
	}
}

func (receiver *GameMessageReceiver) receiveGameMessage(
	messageType MessageType, rawMessage json.RawMessage,
) {
	var err error // Error declared here in order to handle it after the switch

	switch messageType {
	case messageTypeSubmitOrders:
		var message SubmitOrdersMessage
		if err = json.Unmarshal(rawMessage, &message); err == nil {
			receiver.orders <- message
		}
	case messageTypeGiveSupport:
		var message GiveSupportMessage
		if err = json.Unmarshal(rawMessage, &message); err == nil {
			receiver.supportNotifier.L.Lock()
			receiver.supports = append(receiver.supports, message)
			receiver.supportNotifier.L.Unlock()
			receiver.supportNotifier.Broadcast()
		}
	case messageTypeWinterVote:
		var message WinterVoteMessage
		if err = json.Unmarshal(rawMessage, &message); err == nil {
			receiver.winterVote <- message
		}
	case messageTypeSword:
		var message SwordMessage
		if err = json.Unmarshal(rawMessage, &message); err == nil {
			receiver.sword <- message
		}
	case messageTypeRaven:
		var message RavenMessage
		if err = json.Unmarshal(rawMessage, &message); err == nil {
			receiver.raven <- message
		}
	}

	if err != nil {
		log.Println(fmt.Errorf("failed to parse message of type '%s': %w", messageType, err))
	}
}

func (lobby *Lobby) ReceiveOrders(fromPlayer string) ([]gametypes.Order, error) {
	player, ok := lobby.getPlayer(fromPlayer)
	if !ok {
		return nil, fmt.Errorf(
			"failed to get order message from player '%s': player not found", fromPlayer,
		)
	}

	orders := <-player.gameMessageReceiver.orders
	return orders.Orders, nil
}

func (lobby *Lobby) ReceiveSupport(
	fromPlayer string, supportingRegion string, embattledRegion string,
) (supportedPlayer string, err error) {
	player, ok := lobby.getPlayer(fromPlayer)
	if !ok {
		return "", fmt.Errorf(
			"failed to get support message from player '%s' in region '%s': player not found",
			fromPlayer, supportingRegion,
		)
	}

	receiver := player.gameMessageReceiver

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
