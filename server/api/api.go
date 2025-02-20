package api

import (
	"fmt"
	"mig/auth"
	"mig/chatroom"
	"mig/user"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HttpApiController struct {
	hub             *WsHub
	userService     *user.Service
	chatroomService *chatroom.Service
	authService     *auth.Service
	db              *pgxpool.Pool
}

func NewHttpApiController(hub *WsHub, userService *user.Service, chatroomService *chatroom.Service, authService *auth.Service, db *pgxpool.Pool) (*HttpApiController, error) {
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

	if db == nil {
		return nil, fmt.Errorf("missing pgxpool")
	}

	controller := &HttpApiController{
		hub:             hub,
		userService:     userService,
		chatroomService: chatroomService,
		authService:     authService,
		db:              db,
	}

	return controller, nil
}
