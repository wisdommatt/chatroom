package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/wisdommatt/chatroom/internal/chatroom"
)

type chatHandler struct {
	logger    *logrus.Logger
	upgrader  websocket.Upgrader
	clients   map[*websocket.Conn]bool
	join      chan *websocket.Conn
	leave     chan *websocket.Conn
	broadcast chan chatroom.ChatMsg
}

func newChatHandler(logger *logrus.Logger, upgrader websocket.Upgrader) *chatHandler {
	return &chatHandler{
		logger:    logger,
		upgrader:  upgrader,
		clients:   make(map[*websocket.Conn]bool),
		join:      make(chan *websocket.Conn),
		leave:     make(chan *websocket.Conn),
		broadcast: make(chan chatroom.ChatMsg),
	}
}

func (h *chatHandler) handleRequest(chatRoomRepo chatroom.Repository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		wsConn, err := h.upgrader.Upgrade(rw, r, nil)
		if err != nil {
			h.logger.WithError(err).Debug("Chat handler error")
			return
		}
		defer wsConn.Close()
		h.join <- wsConn
		defer func() {
			h.leave <- wsConn
		}()
		for {
			msg := chatroom.ChatMsg{}
			err := wsConn.ReadJSON(&msg)
			if err != nil {
				h.logger.WithError(err).Debug("Chat websocket read message error ...")
				break
			}
			h.logger.WithFields(logrus.Fields{
				"message": msg.Message,
				"sender":  msg.SenderName,
			}).Info("New chat message received")
			h.broadcast <- msg
			err = chatRoomRepo.SaveMessage("", &msg)
			if err != nil {
				h.logger.WithError(err).Error("Save chat message error !")
			}
		}
	}
}

func (h *chatHandler) wsConnectionListener() {
	h.logger.Info("Listening for websocket clients connections / disconnections ")
	for {
		select {
		case client := <-h.join:
			h.logger.Info("New client joining ...")
			h.clients[client] = true

		case client := <-h.leave:
			h.logger.Info("Client leaving ...")
			delete(h.clients, client)

		case msg := <-h.broadcast:
			// converting msg to JSON bytes to make the message broadcast process faster.
			msgJSONBytes, _ := json.Marshal(msg)
			for client := range h.clients {
				err := client.WriteMessage(websocket.TextMessage, msgJSONBytes)
				if err != nil {
					h.logger.WithError(err).Debug("Broadcast chat message error")
				}
			}
			h.logger.WithFields(logrus.Fields{
				"message":    msg.Message,
				"senderName": msg.SenderName,
			}).Info("Message broadcast successful")
		}
	}
}
