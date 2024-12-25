package api

import (
	"fmt"
	"mig/chatroom"
	"mig/user"
)

type HttpApiController struct {
	hub             *WsHub
	userService     *user.Service
	chatroomService *chatroom.Service
}

func NewHttpApiController(hub *WsHub, userService *user.Service, chatroomService *chatroom.Service) (*HttpApiController, error) {
	if hub == nil {
		return nil, fmt.Errorf("missing websocket hub")
	}
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
