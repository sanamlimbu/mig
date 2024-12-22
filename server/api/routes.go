package api

import (
	"fmt"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewHttpRouter(c *HttpApiController) (*chi.Mux, error) {
	if c == nil {
		return nil, fmt.Errorf("missing http api controller")
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(time.Second * 15))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Route("/api", func(r chi.Router) {
		r.With(paginate).Get("/users/{user_id}/friends", c.GetFriends)
		r.With(paginate).Get("/users/{user_id}/private-messages/{recipient_id}", c.GetPrivateMessages)
		r.Get("/users/{user_id}", c.GetUser)

		r.With(paginate).Get("/chatrooms", c.GetChatrooms)
		r.Get("/chatrooms/{chatroom_id}", c.GetChatroom)
		r.With(paginate).Get("/chatrooms/{chatroom_id}/messages", c.GetChatroomMessages)
	})

	return r, nil
}
