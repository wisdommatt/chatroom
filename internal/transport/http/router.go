package http

import (
	"net/http"

	"github.com/Meghee/kit/router"
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
	chatHandler := newChatHandler(h.logger, chatroomRepo)

	h.Router.Get("/assets/*", router.FileRouter("./static/assets", "/assets/"))
	h.Router.Get("/", handleIndexPage)
	h.Router.Get("/websocket/chat/{roomId}", chatHandler.handleRequest(h.wsUpgrader))
	h.Router.Post("/chatroom/", handleCreateChatRoom(chatroomRepo, h.logger))
	h.Router.Get("/cr/{roomId}", handleOpenChatRoomPage(chatHandler.logger))
}
