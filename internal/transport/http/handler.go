package http

import (
	"net/http"
	"time"

	"github.com/Meghee/kit/web"

	"github.com/Meghee/kit/json"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-playground/validator"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/wisdommatt/chatroom/internal/chatroom"
	"github.com/wisdommatt/randgen"
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

	h.Router.Get("/websocket/chat", chatHandler.handleRequest(h.wsUpgrader))
}

type createChatRoomPayload struct {
	Name string `json:"name" validate:"required"`
}

func (entity *createChatRoomPayload) validate() error {
	return validator.New().Struct(entity)
}

// HandleCreateChatRoom is the route handler for creat chatroom endpoint.
func HandleCreateChatRoom(chatroomRepo chatroom.Repository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var chatRoom createChatRoomPayload
		json.DecodeAndEscapeHTML(r.Body, &chatRoom)
		err := chatRoom.validate()
		if err != nil {
			web.JSONErrorResponse(rw, http.StatusBadRequest, "Invalid JSON payload !")
			return
		}
		newChatRoom := chatroom.ChatRoom{
			Name:      chatRoom.Name,
			DeletePin: randgen.NewStringGenerator().GenerateFromSource(randgen.StringAlphaNumericSource, 50),
			TimeAdded: time.Now(),
		}
		err = chatroomRepo.SaveChatRoom(&newChatRoom)
		if err != nil {
			web.JSONErrorResponse(rw, http.StatusInternalServerError, "An error occured while creating chatroom !")
			return
		}
		web.JSONResponse(rw, http.StatusOK, map[string]interface{}{
			"status":   "success",
			"message":  "Chat room created successfully !",
			"chatroom": newChatRoom,
		})
	}
}
