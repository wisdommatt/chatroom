package http

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/wisdommatt/chatroom/internal/chatroom"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	Router     chi.Router
	wsUpgrader websocket.Upgrader
	logger     *logrus.Logger
}

// NewHandler returns a new HTTP transport handler.
func NewHandler(logger *logrus.Logger) *Handler {
	return &Handler{
		wsUpgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		logger: logger,
	}
}

// SetupRoutes sets up all the application routes.
func (h *Handler) SetupRoutes(database *mongo.Database) {
	h.logger.Info("Setting up routes")
	h.Router = chi.NewRouter()
	h.Router.Use(middleware.RealIP)
	h.Router.Use(middleware.RequestID)
	h.Router.Use(middleware.Logger)

	chatroomRepo := chatroom.NewRepository(database)
	chatHandler := newChatHandler(h.logger, h.wsUpgrader)
	go chatHandler.wsConnectionListener()

	h.Router.Get("/websocket/chat", chatHandler.handleRequest(chatroomRepo))
}
