package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestWsChatHandler(t *testing.T) {
	logger := &logrus.Logger{}
	handler := NewHandler(logger)
	chatHandler := newChatHandler(logger, handler.wsUpgrader)
	go chatHandler.wsConnectionListener()

	server := httptest.NewServer(http.HandlerFunc(chatHandler.handleEndpoint()))
	defer server.Close()

	var connection1 *websocket.Conn
	var connection2 *websocket.Conn
	testCases := map[string]struct {
		connectionValid bool
		url             string
	}{
		"Valid connection": {
			connectionValid: true,
			url:             "ws" + strings.TrimPrefix(server.URL, "http"),
		},
		"Invalid connection": {
			connectionValid: false,
			url:             server.URL,
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			var err error
			connection1, _, err = websocket.DefaultDialer.Dial(testCase.url, nil)
			if !testCase.connectionValid {
				require.NotNil(t, err, err)
				return
			}
			require.Nil(t, err, err)
			defer connection1.Close()

			connection2, _, err = websocket.DefaultDialer.Dial(testCase.url, nil)
			require.Nil(t, err, err)
			defer connection2.Close()

			for i := 0; i < 10; i++ {
				err := connection1.WriteJSON(ChatMsg{
					Message:    "Test Message",
					SenderName: "Connection One",
				})
				require.Nil(t, err, err)

				msg := ChatMsg{}
				err = connection1.ReadJSON(&msg)
				require.Nil(t, err, err)
				err = connection2.ReadJSON(&msg)
				require.Nil(t, err, err)
			}
			for i := 0; i < 10; i++ {
				err := connection2.WriteJSON(ChatMsg{
					Message:    "Test Message",
					SenderName: "Connection Two",
				})
				require.Nil(t, err, err)

				msg := ChatMsg{}
				err = connection1.ReadJSON(&msg)
				require.Nil(t, err, err)
				err = connection2.ReadJSON(&msg)
				require.Nil(t, err, err)
			}
		})
	}
}
