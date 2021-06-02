package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/require"
	"github.com/wisdommatt/chatroom/internal/chatroom"
	"github.com/wisdommatt/chatroom/test/mocks"
)

func TestHandleCreateChatRoom(t *testing.T) {
	testCases := map[string]struct {
		payload            createChatRoomPayload
		chatroomRepo       chatroom.Repository
		expectedStatusCode int
	}{
		"correct payload": {
			payload: createChatRoomPayload{Name: "Correct chatroom"},
			chatroomRepo: &mocks.ChatRoomRepo{
				SaveChatRoomFunc: func(chatRoom *chatroom.ChatRoom) error {
					chatRoom.ID = "helloWorld"
					return nil
				},
			},
			expectedStatusCode: http.StatusOK,
		},
		"payload without name": {
			payload:            createChatRoomPayload{},
			expectedStatusCode: http.StatusBadRequest,
		},
		"correct payload with err chatroomRepo implementation": {
			payload: createChatRoomPayload{Name: "Correct chatroom"},
			chatroomRepo: &mocks.ChatRoomRepo{
				SaveChatRoomFunc: func(chatRoom *chatroom.ChatRoom) error {
					return errors.New("An error occured")
				},
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			payloadJSON, err := json.Marshal(testCase.payload)
			require.Nil(t, err, err)

			r := chi.NewRouter()
			rec := httptest.NewRecorder()
			r.Post("/chatroom", HandleCreateChatRoom(testCase.chatroomRepo))
			req := httptest.NewRequest("POST", "/chatroom", bytes.NewBuffer(payloadJSON))
			r.ServeHTTP(rec, req)
			require.Exactly(t, testCase.expectedStatusCode, rec.Result().StatusCode)

			var apiResponse map[string]interface{}
			err = json.Unmarshal(rec.Body.Bytes(), &apiResponse)
			require.NotEmpty(t, apiResponse)
			if rec.Result().StatusCode == http.StatusOK {
				require.Exactly(t, "success", apiResponse["status"])
				require.NotEmpty(t, apiResponse["message"])
				require.NotEmpty(t, apiResponse["chatroom"].(map[string]interface{})["deletePin"])
			}
		})
	}
}
