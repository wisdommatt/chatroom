package chatroom

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Meghee/kit/database/mongodb"
	"github.com/Meghee/kit/dotenv"
	"github.com/stretchr/testify/require"
	"github.com/wisdommatt/randgen"
)

func setupRepo(t *testing.T) (repo *ChatRoomRepo, tearDown func()) {
	randomDBname := "testDB_" + randgen.NewStringGenerator().GenerateFromSource(randgen.StringAlphabetsSource, 15)
	mongoDBClient, err := mongodb.Connect(os.Getenv("DATABASE_URI"))
	require.Nil(t, err, err)
	mongoDB := mongoDBClient.Database(randomDBname)
	return NewRepository(mongoDB), func() {
		err = mongoDB.Drop(context.TODO())
		require.Nil(t, err, err)
		err = mongoDBClient.Disconnect(context.TODO())
		require.Nil(t, err, err)
	}
}

func TestSaveChatRoom(t *testing.T) {
	dotenv.LoadEnvironmentVariables("../../")
	repo, tearDown := setupRepo(t)
	defer tearDown()
	testCases := map[string]ChatRoom{
		"empty entity": {},
		"complete entity": {
			Name:      "Sample Chat",
			DeletePin: randgen.NewStringGenerator().GenerateFromSource(randgen.StringAlphaNumericSource, 50),
			TimeAdded: time.Now(),
			Chats: []ChatMsg{
				{
					Message:    "Welcome",
					SenderName: "Wisdom Matt",
				},
				{
					Message:    "Message 2",
					SenderName: "Wisdom Matthew",
				},
			},
		},
	}
	for name, entity := range testCases {
		t.Run(name, func(t *testing.T) {
			err := repo.SaveChatRoom(&entity)
			require.Nil(t, err, err)
			require.NotEmpty(t, entity.ID)
		})
	}
}

func TestSaveMessage(t *testing.T) {
	dotenv.LoadEnvironmentVariables("../../")
	repo, tearDown := setupRepo(t)
	defer tearDown()

	testProcesses := map[string]func(t *testing.T){
		"Correct process": func(t *testing.T) {
			chatroom := ChatRoom{Name: "Welcome"}
			err := repo.SaveChatRoom(&chatroom)
			require.Nil(t, err, err)
			msg := ChatMsg{
				Message: "This is a message !",
			}
			err = repo.SaveMessage(chatroom.ID, &msg)
			require.Nil(t, err, err)
			require.NotEmpty(t, msg.ID)
		},
		"Incorrect process": func(t *testing.T) {
			chatRoomID := "helloWorld"
			msg := ChatMsg{
				Message: "This is a sample message !",
			}
			err := repo.SaveMessage(chatRoomID, &msg)
			require.NotNil(t, err, err)
			require.NotEmpty(t, msg.ID)
		},
	}
	for name, process := range testProcesses {
		t.Run(name, func(t *testing.T) {
			process(t)
		})
	}
}
