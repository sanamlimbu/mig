package api

import (
	"fmt"
	"mig/chatroom"
	"mig/user"
)

type HttpApiController struct {
	userService     *user.Service
	chatroomService *chatroom.Service
}

func NewHttpApiController(userService *user.Service, chatroomService *chatroom.Service) (*HttpApiController, error) {
	if userService == nil {
		return nil, fmt.Errorf("missing user service")
	}

	if chatroomService == nil {
		return nil, fmt.Errorf("missing chatroom service")
	}

	controller := &HttpApiController{
		userService:     userService,
		chatroomService: chatroomService,
	}

	return controller, nil
}
