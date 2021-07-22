package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/wisdommatt/chatroom/internal/chatroom"
)

type chatHandler struct {
	logger          *logrus.Logger
	chatroomRepo    chatroom.Repository
	activeChatRooms map[string]*room
}

func newChatHandler(logger *logrus.Logger, repo chatroom.Repository) *chatHandler {
	return &chatHandler{
		logger:          logger,
		chatroomRepo:    repo,
		activeChatRooms: make(map[string]*room),
	}
}

func (h *chatHandler) getRoom(id string) *room {
	if _, exist := h.activeChatRooms[id]; !exist {
		r := &room{
			join:      make(chan *websocket.Conn),
			leave:     make(chan *websocket.Conn),
			broadcast: make(chan chatroom.ChatMsg),
			handler:   h,
		}
		h.activeChatRooms[id] = r
		go r.run(id)
	}
	return h.activeChatRooms[id]
}

func (h *chatHandler) handleRequest(upgrader websocket.Upgrader) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		wsConn, err := upgrader.Upgrade(rw, r, nil)
		if err != nil {
			h.logger.WithError(err).Debug("Chat handler error")
			return
		}
		defer wsConn.Close()
		roomId := r.URL.Query().Get("roomId")
		if roomId == "" {
			return
		}
		chatRoom := h.getRoom(roomId)
		chatRoom.join <- wsConn
		defer func() {
			chatRoom.leave <- wsConn
		}()
		for {
			msg := chatroom.ChatMsg{}
			err := wsConn.ReadJSON(&msg)
			if err != nil {
				h.logger.WithError(err).Debug("Chat websocket read message error ...")
				break
			}
			h.logger.Info("New chat message received")
			msg.TimeSent = time.Now()
			chatRoom.broadcast <- msg
		}
	}
}

type room struct {
	join      chan *websocket.Conn
	leave     chan *websocket.Conn
	broadcast chan chatroom.ChatMsg
	handler   *chatHandler
}

func (room *room) run(roomID string) {
	logger := room.handler.logger
	chatroomRepo := room.handler.chatroomRepo
	logger.Info("Listening for websocket clients connections / disconnections ")
	clients := make(map[*websocket.Conn]bool)
	for {
		select {
		case client := <-room.join:
			logger.Info("New client joining ...")
			clients[client] = true

		case client := <-room.leave:
			logger.Info("Client leaving ...")
			delete(clients, client)
			if len(clients) == 0 {
				delete(room.handler.activeChatRooms, roomID)
				break
			}

		// broadcast channel broadcasts a message and also save it in the database.
		case msg := <-room.broadcast:
			// converting msg to JSON bytes to make the message broadcast process faster.
			msgJSONBytes, _ := json.Marshal(msg)
			for client := range clients {
				err := client.WriteMessage(websocket.TextMessage, msgJSONBytes)
				if err != nil {
					logger.WithError(err).Debug("Broadcast chat message error")
				}
			}
			logger.Info("Message broadcast successful")
			err := chatroomRepo.SaveMessage(roomID, &msg)
			if err != nil {
				logger.WithError(err).Error("Save chat message error !")
			}
		}
	}
}
