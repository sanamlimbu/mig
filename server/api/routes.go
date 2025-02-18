package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewHttpRouter(c *HttpApiController, allowedOrigins string) (*chi.Mux, error) {
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
		AllowedOrigins:   strings.Split(allowedOrigins, ","),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.HandleFunc("/ws", c.hub.serveWebSockets)

	r.Route("/api", func(r chi.Router) {
		r.Post("/login", c.Login)
		r.Post("/signup", c.Signup)
		r.Post("/refresh-token", c.RefreshToken)

		r.With(paginate).Get("/users/{user_id}/friends", withAuth(c, c.GetFriends))
		r.With(paginate).Get("/users/{user_id}/private-conversation/{recipient_id}", withAuth(c, c.GetPrivateConversation))
		r.Get("/users/{user_id}", withAuth(c, c.GetUser))
		r.With(paginate).Get("/users/{user_id}/recent-private-messages", withAuth(c, c.GetRecentPrivateMessages))

		r.With(paginate).Get("/chatrooms", c.GetChatrooms)
		r.Get("/chatrooms/{chatroom_id}", c.GetChatroom)
		r.With(paginate).Get("/chatrooms/{chatroom_id}/messages", c.GetChatroomMessages)

		r.Get("/hello", hello)
	})

	return r, nil
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"message":"Hello World","time":"%s"}`, time.Now())))
}
