package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/wisdommatt/chatroom/internal/chatroom"
	"github.com/wisdommatt/chatroom/test/mocks"
)

func TestWsChatHandler(t *testing.T) {
	outFile, _ := os.Create("test.logs")
	logger := logrus.New()
	logger.SetOutput(outFile)
	logger.SetFormatter(&logrus.JSONFormatter{PrettyPrint: true})

	chatroomRepo := &mocks.ChatRoomRepo{
		SaveMessageFunc: func(chatRoomID string, msg *chatroom.ChatMsg) error {
			return nil
		},
	}
	testCases := map[string]struct {
		connectionValid bool
		url             func(server *httptest.Server) string
	}{
		"Valid connection url": {
			connectionValid: true,
			url: func(server *httptest.Server) string {
				return "ws" + strings.TrimPrefix(server.URL, "http")
			},
		},
		"Invalid connection url": {
			connectionValid: false,
			url:             func(server *httptest.Server) string { return server.URL },
		},
		"SaveMessage err chatroom repo": {
			connectionValid: true,
			url: func(server *httptest.Server) string {
				return "ws" + strings.TrimPrefix(server.URL, "http")
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			handler := NewHandler(logger)
			chatHandler := newChatHandler(logger, chatroomRepo)

			connections := []*websocket.Conn{}
			server := httptest.NewServer(http.HandlerFunc(chatHandler.handleRequest(handler.wsUpgrader)))
			defer server.Close()

			// setting up 20 connections .
			for i := 0; i < 20; i++ {
				connection, _, err := websocket.DefaultDialer.Dial(testCase.url(server), nil)
				if !testCase.connectionValid {
					require.NotNil(t, err, err)
					return
				}
				require.Nil(t, err, err)
				defer connection.Close()
				connections = append(connections, connection)
			}

			for _, connection := range connections {
				for i := 0; i < 10; i++ {
					msg := chatroom.ChatMsg{
						Message:    "Test Message - " + strconv.Itoa(i),
						SenderName: "Connection - " + strconv.Itoa(i),
					}
					err := connection.WriteJSON(msg)
					require.Nil(t, err, err)

					for _, conn := range connections {
						msg := chatroom.ChatMsg{}
						err = conn.ReadJSON(&msg)
						require.Nil(t, err, err)
						require.Exactly(t, msg.Message, msg.Message)
						require.Exactly(t, msg.SenderName, msg.SenderName)
					}
				}
			}
		})
	}
}
