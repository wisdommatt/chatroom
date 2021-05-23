package http

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
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
func (h *Handler) SetupRoutes() {
	h.logger.Info("Setting up routes")
	h.Router = chi.NewRouter()
	h.Router.Use(middleware.RealIP)
	h.Router.Use(middleware.RequestID)
	h.Router.Use(middleware.Logger)

	// wsChannels = map[string]chan *websocket.Conn{
	// 	"join": make(chan *websocket.Conn),
	// 	"leave": make(chan *websocket.Conn),
	// 	"getClients": make(chan *websocket.Conn),
	// }

	chatHandler := newChatHandler(h.logger, h.wsUpgrader)

	// joinChatChan, leaveChatChan := make(chan *websocket.Conn), make(chan *websocket.Conn)
	// getClients := make(chan map[*websocket.Conn]bool)
	go chatHandler.wsConnectionListener()
	// go wsConnectionListener(h.logger, joinChatChan, leaveChatChan, getClients)(make(map[*websocket.Conn]bool))
	h.Router.Get("/websocket/chat", chatHandler.handleEndpoint())
}
