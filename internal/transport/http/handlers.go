package http

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/Meghee/kit/web"

	"github.com/Meghee/kit/json"
	"github.com/sirupsen/logrus"
	"github.com/wisdommatt/chatroom/internal/chatroom"
	"github.com/wisdommatt/randgen"
)

// handleIndexPage is the route handler for index page.
func handleIndexPage(rw http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("index.html").ParseFiles("./static/templates/index.html"))
	t.Execute(rw, nil)
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
		})
	}
}

// handleOpenChatRoomPage is the route handler for chatroom view page.
func handleOpenChatRoomPage(logger *logrus.Logger) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		t := template.Must(template.New("chatroom.html").ParseFiles("./static/templates/chatroom.html"))
		t.Execute(rw, nil)
	}
}
