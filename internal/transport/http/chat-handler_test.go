package http

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestWsChatHandler(t *testing.T) {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint: true,
	})

	handler := NewHandler(logger)
	chatHandler := newChatHandler(logger, handler.wsUpgrader)
	go chatHandler.wsConnectionListener()

	server := httptest.NewServer(http.HandlerFunc(chatHandler.handleRequest))
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
				msg := ChatMsg{
					Message:    "Test Message 1 - " + strconv.Itoa(i),
					SenderName: "Connection One - " + strconv.Itoa(i),
				}
				err := connection1.WriteJSON(msg)
				require.Nil(t, err, err)

				msg1 := ChatMsg{}
				err = connection1.ReadJSON(&msg1)
				require.Nil(t, err, err)
				require.Exactly(t, msg.Message, msg1.Message)
				require.Exactly(t, msg.SenderName, msg1.SenderName)

				msg2 := ChatMsg{}
				err = connection2.ReadJSON(&msg2)
				require.Nil(t, err, err)
				require.Exactly(t, msg1, msg2)
			}
			for i := 0; i < 10; i++ {
				msg := ChatMsg{
					Message:    "Test Message 2 - " + strconv.Itoa(i),
					SenderName: "Connection Two - " + strconv.Itoa(i),
				}
				err := connection2.WriteJSON(msg)
				require.Nil(t, err, err)

				msg1 := ChatMsg{}
				err = connection1.ReadJSON(&msg1)
				require.Nil(t, err, err)
				require.Exactly(t, msg.Message, msg1.Message)
				require.Exactly(t, msg.SenderName, msg1.SenderName)

				msg2 := ChatMsg{}
				err = connection2.ReadJSON(&msg2)
				require.Nil(t, err, err)
				require.Exactly(t, msg1, msg2)
			}
		})
	}
}
