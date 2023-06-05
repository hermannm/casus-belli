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
				log.Println(err)
			}
			break
		}
		if err != nil {
			log.Println(fmt.Errorf("message error for player %s: %w", player.String(), err))
			player.SendError(err)
		}
	}
}

// Reads a message from the player's socket, and handles it appropriately.
// Returns socketClosed=true if the socket closed, and an error if message handling failed.
func (player *Player) receiveMessage(lobby *Lobby) (socketClosed bool, err error) {
	player.lock.RLock()
	defer player.lock.RUnlock()

	_, receivedMsg, err := player.socket.ReadMessage()
	if err != nil {
		switch err := err.(type) {
		case *websocket.CloseError:
			return true, fmt.Errorf("socket closed: %w", err)
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
			return false, fmt.Errorf(
				"received game message from player '%s' before their game ID was set",
				player.String(),
			)
		} else {
			go player.gameMessageReceiver.receiveGameMessage(messageType, rawMsg)
		}
	}

	return false, nil
}

// Receives a lobby-specific message, and handles it according to its ID.
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
	orders     chan SubmitOrdersMessage
	winterVote chan WinterVoteMessage
	sword      chan SwordMessage
	raven      chan RavenMessage

	supports          []GiveSupportMessage // Must hold supportsCondition.L to access safely.
	supportsCondition sync.Cond
}

func newGameMessageReceiver() *GameMessageReceiver {
	return &GameMessageReceiver{
		orders:            make(chan SubmitOrdersMessage),
		winterVote:        make(chan WinterVoteMessage),
		sword:             make(chan SwordMessage),
		raven:             make(chan RavenMessage),
		supports:          nil,
		supportsCondition: sync.Cond{L: &sync.Mutex{}},
	}
}

// Takes a message ID and an unserialized JSON message.
// Unmarshals the message according to its type, and sends it to the appropraite receiver channel.
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
			receiver.supportsCondition.L.Lock()
			receiver.supports = append(receiver.supports, message)
			receiver.supportsCondition.L.Unlock()
			receiver.supportsCondition.Broadcast()
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

	receiver.supportsCondition.L.Lock()
	for {
		var supportedPlayer string
		remainingSupports := make([]GiveSupportMessage, 0, cap(receiver.supports))
		for _, support := range receiver.supports {
			if support.SupportingRegion == supportingRegion &&
				support.EmbattledRegion == embattledRegion {
				supportedPlayer = *support.SupportedPlayer
			} else {
				remainingSupports = append(remainingSupports, support)
			}
		}

		if supportedPlayer != "" {
			receiver.supports = remainingSupports
			receiver.supportsCondition.L.Unlock()
			return supportedPlayer, nil
		}

		receiver.supportsCondition.Wait()
	}
}
