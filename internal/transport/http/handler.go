package http

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/Meghee/kit/router"
	"golang.org/x/crypto/bcrypt"

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

	h.Router.Get("/assets/*", router.FileRouter("./static/assets", "/assets/"))
	h.Router.Get("/", handleIndexPage)
	h.Router.Get("/websocket/chat", chatHandler.handleRequest(h.wsUpgrader))
	h.Router.Post("/chatroom/", handleCreateChatRoom(chatroomRepo, h.logger))
}

// handleIndexPage is the route handler for index page.
func handleIndexPage(rw http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("index.html").ParseFiles("./static/templates/index.html"))
	t.Execute(rw, nil)
}

type createChatRoomPayload struct {
	Name       string `json:"name" validate:"required"`
	InviteCode string `json:"inviteCode" validate:"required"`
}

func (entity *createChatRoomPayload) validate() error {
	entity.InviteCode = strings.ToLower(entity.InviteCode)
	return validator.New().Struct(entity)
}

// handleCreateChatRoom is the route handler for creat chatroom endpoint.
func handleCreateChatRoom(chatroomRepo chatroom.Repository, logger *logrus.Logger) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var chatRoom createChatRoomPayload
		json.DecodeAndEscapeHTML(r.Body, &chatRoom)
		err := chatRoom.validate()
		if err != nil {
			web.JSONErrorResponse(rw, http.StatusBadRequest, "Invalid JSON payload !")
			return
		}
		// hashing invite code.
		if chatRoom.InviteCode != "" {
			hashedCode, err := bcrypt.GenerateFromPassword([]byte(chatRoom.InviteCode), bcrypt.MinCost)
			if err != nil {
				logger.WithField("inviteCode", chatRoom.InviteCode).WithError(err).
					Error("An error occured while hashing invite code")
				web.JSONErrorResponse(rw, http.StatusInternalServerError, "An error occured, please try again later !")
				return
			}
			chatRoom.InviteCode = string(hashedCode)
		}
		strGen := randgen.NewStringGenerator()
		rawRoomPin := strGen.GenerateFromSource(randgen.StringAlphaNumericSource, 40)
		// hashing room pin.
		hashedPin, _ := bcrypt.GenerateFromPassword([]byte(rawRoomPin), bcrypt.MinCost)
		roomPin := string(hashedPin)
		newChatRoom := chatroom.ChatRoom{
			Name:       chatRoom.Name,
			URL:        strGen.GenerateFromSource(randgen.StringAlphaNumericSource, 25),
			InviteCode: chatRoom.InviteCode,
			RoomPin:    roomPin,
			TimeAdded:  time.Now(),
		}
		err = chatroomRepo.SaveChatRoom(&newChatRoom)
		if err != nil {
			logger.WithError(err).Error("An error occured while saving chatroom in DB")
			web.JSONErrorResponse(rw, http.StatusInternalServerError, "An error occured while creating chatroom !")
			return
		}
		logger.Info(fmt.Sprintf("Chatroom created successfully %s %s", newChatRoom.Name, newChatRoom.ID))
		web.JSONResponse(rw, http.StatusOK, map[string]interface{}{
			"status":    "success",
			"message":   "Chat room created successfully !",
			"roomUrl":   newChatRoom.URL,
			"actionPin": rawRoomPin,
			"chatroom":  newChatRoom,
		})
	}
}
