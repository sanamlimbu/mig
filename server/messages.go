package mig

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
)

type MessagesService struct {
	usersRepo     UsersRepository
	chatroomsRepo ChatroomsRepository
}

func NewMessagesSerive(usersRepo UsersRepository, chatroomsRepo ChatroomsRepository) *MessagesService {
	return &MessagesService{
		usersRepo:     usersRepo,
		chatroomsRepo: chatroomsRepo,
	}
}

func (s *MessagesService) isValidMessage(payload []byte) (bool, error) {
	var msg Message
	if err := json.Unmarshal(payload, &msg); err != nil {
		log.Error().Msg(err.Error())
		return false, err
	}

	if msg.MessageType == MessageTypePrivate {
		return canSendPrivateMesage(s, msg)

	} else if msg.MessageType == MessageTypeChatroom {
		return canSendChatroomMessage(s, msg)

	} else {
		return false, fmt.Errorf("invalid message type: %s", msg.MessageType)
	}
}

func canSendPrivateMesage(s *MessagesService, msg Message) (bool, error) {
	_, err := s.usersRepo.getFriendByID(context.Background(), msg.SenderID, msg.RecipientID)
	if err != nil {
		return false, fmt.Errorf("unauthorised message operation")
	}

	return true, nil
}

func canSendChatroomMessage(s *MessagesService, msg Message) (bool, error) {
	return true, nil
}
