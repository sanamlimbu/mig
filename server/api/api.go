package api

import (
	"fmt"
	"mig/auth"
	"mig/chatroom"
	"mig/user"
)

type HttpApiController struct {
	hub             *WsHub
	userService     *user.Service
	chatroomService *chatroom.Service
	authService     *auth.Service
}

func NewHttpApiController(hub *WsHub, userService *user.Service, chatroomService *chatroom.Service, authService *auth.Service) (*HttpApiController, error) {
	if hub == nil {
		return nil, fmt.Errorf("missing websocket hub")
	}

	if userService == nil {
		return nil, fmt.Errorf("missing user service")
	}

	if chatroomService == nil {
		return nil, fmt.Errorf("missing chatroom service")
	}

	if authService == nil {
		return nil, fmt.Errorf("missing auth service")
	}

	controller := &HttpApiController{
		hub:             hub,
		userService:     userService,
		chatroomService: chatroomService,
		authService:     authService,
	}

	return controller, nil
}
