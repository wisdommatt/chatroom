package mocks

import "github.com/wisdommatt/chatroom/internal/chatroom"

type ChatRoomRepo struct {
	chatroom.ChatRoomRepo
	SaveChatRoomFunc func(chatRoom *chatroom.ChatRoom) error
	SaveMessageFunc  func(chatRoomID string, msg *chatroom.ChatMsg) error
}

func (repo *ChatRoomRepo) SaveChatRoom(room *chatroom.ChatRoom) error {
	return repo.SaveChatRoomFunc(room)
}

func (repo *ChatRoomRepo) SaveMessage(roomID string, msg *chatroom.ChatMsg) error {
	return repo.SaveMessageFunc(roomID, msg)
}
