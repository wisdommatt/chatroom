package http

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type ChatMsg struct {
	Message    string `json:"message"`
	SenderName string `json:"senderName"`
}

type chatHandler struct {
	logger   *logrus.Logger
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
	join     chan *websocket.Conn
	leave    chan *websocket.Conn
	dispatch chan ChatMsg
}

func newChatHandler(logger *logrus.Logger, upgrader websocket.Upgrader) *chatHandler {
	return &chatHandler{
		logger:   logger,
		upgrader: upgrader,
		clients:  make(map[*websocket.Conn]bool),
		join:     make(chan *websocket.Conn),
		leave:    make(chan *websocket.Conn),
		dispatch: make(chan ChatMsg),
	}
}

func (h *chatHandler) handleEndpoint() http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		wsConn, err := h.upgrader.Upgrade(rw, r, nil)
		if err != nil {
			h.logger.WithError(err).Info("Chat handler error")
			return
		}
		h.join <- wsConn
		defer func() {
			h.leave <- wsConn
		}()
		go h.writer(wsConn, h.logger)
		h.reader(wsConn, h.logger)
	})
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

		case msg := <-h.dispatch:
			for client, _ := range h.clients {
				err := client.WriteJSON(msg)
				if err != nil {
					h.logger.WithError(err).Error("Dispatch chat message error")
				}
			}
		}
	}
}

func (h *chatHandler) reader(conn *websocket.Conn, logger *logrus.Logger) {
	defer conn.Close()
	for {
		msg := ChatMsg{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			logger.WithError(err).Debug("Chat websocket read message error ...")
			break
		}
		logger.WithFields(logrus.Fields{
			"message": msg.Message,
			"sender":  msg.SenderName,
		}).Info("New chat message received")
		h.dispatch <- msg
	}
}

func (h *chatHandler) writer(conn *websocket.Conn, logger *logrus.Logger) {

}
