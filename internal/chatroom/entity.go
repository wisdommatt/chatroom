package chatroom

import (
	"time"

	"github.com/go-playground/validator"
)

type ChatRoom struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	Name      string    `json:"name" bson:"name,omitempty" validate:"required"`
	DeletePin string    `json:"deletePin" bson:"deletePin,omitempty" validate:"required,min=30"`
	TimeAdded time.Time `json:"timeAdded" bson:"timeAdded,omitempty" validate:"required"`
	Chats     []ChatMsg `json:"chats" bson:"chats,omitempty"`
}

func (entity *ChatRoom) Validate() error {
	return validator.New().Struct(entity)
}

type ChatMsg struct {
	ID         string    `json:"id" bson:"_id,omitempty"`
	Message    string    `json:"message" bson:"message,omitempty" validate:"required"`
	SenderName string    `json:"senderName" bson:"senderName,omitempty" validate:"required"`
	TimeSent   time.Time `json:"timeSent" bson:"timeSent,omitempty" validate:"required"`
}

func (entity *ChatMsg) Validate() error {
	return validator.New().Struct(entity)
}
