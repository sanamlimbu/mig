package api

import (
	"fmt"
	"mig/user"
)

type HttpApiController struct {
	userHandler *user.Handler
}

func NewHttpApiController(userHandler *user.Handler) (*HttpApiController, error) {
	if userHandler == nil {
		return nil, fmt.Errorf("missing user handler")
	}

	controller := &HttpApiController{
		userHandler: userHandler,
	}

	return controller, nil
}
