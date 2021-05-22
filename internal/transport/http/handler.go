package http

import (
	"net/http"

	"github.com/go-chi/chi"
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
	h.Router.Get("/websocket/chat", wsChatHandler())
}

func wsChatHandler() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
	})
}
