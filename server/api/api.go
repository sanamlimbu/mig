package api

import (
	"fmt"
	"mig/chatroom"
	"mig/user"
)

type HttpApiController struct {
	userHandler     *user.Handler
	chatroomHandler *chatroom.Handler
}

func NewHttpApiController(userHandler *user.Handler, chatroomHandler *chatroom.Handler) (*HttpApiController, error) {
	if userHandler == nil {
		return nil, fmt.Errorf("missing user handler")
	}

	if chatroomHandler == nil {
		return nil, fmt.Errorf("missing chatroom handler")
	}

	controller := &HttpApiController{
		userHandler:     userHandler,
		chatroomHandler: chatroomHandler,
	}

	return controller, nil
}
