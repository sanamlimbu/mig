package mig

import (
	"fmt"
	"mig/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type APIController struct {
	auther           Auther
	hub              *Hub
	chatroomsService *ChatroomsService
	usersService     *UsersService
}

func NewAPIController(auther Auther, hub *Hub, chatroomsService *ChatroomsService, usersService *UsersService) *APIController {
	return &APIController{
		auther:           auther,
		hub:              hub,
		chatroomsService: chatroomsService,
		usersService:     usersService,
	}
}

type Auther struct {
	jwtKey []byte
	issuer string
}

type JWTClaims struct {
	ID            int64                     `json:"id"`
	Username      string                    `json:"username"`
	WorkflowState models.UsersWorkflowState `json:"workflow_state"`
	jwt.RegisteredClaims
}

type JWTUser struct {
	id            int64
	username      string
	workflowState models.UsersWorkflowState
}

func NewAuther(secret, issuer string) Auther {
	return Auther{
		jwtKey: []byte(secret),
		issuer: issuer,
	}
}

func (a *Auther) New(user JWTUser, audience []string) (string, error) {
	claims := JWTClaims{
		ID:            user.id,
		Username:      user.username,
		WorkflowState: user.workflowState,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    a.issuer,
			Subject:   fmt.Sprintf("%v", user.id),
			ID:        fmt.Sprintf("%v", user.id),
			Audience:  audience,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(a.jwtKey)
	if err != nil {
		return "", err
	}

	return ss, nil
}
