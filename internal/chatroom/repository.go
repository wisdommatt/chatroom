package chatroom

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Repository is the interface that describes a chatroom repository object.
type Repository interface {
	SaveChatRoom(chatRoom *ChatRoom) error
}

type ChatRoomRepo struct {
	collection *mongo.Collection
}

// NewRepository returns a new chatroom repository object that implements the Repository
// interface.
func NewRepository(db *mongo.Database) *ChatRoomRepo {
	return &ChatRoomRepo{
		collection: db.Collection("chatrooms"),
	}
}

// SaveChatRoom saves a chat room to the database.
func (repo *ChatRoomRepo) SaveChatRoom(chatRoom *ChatRoom) error {
	chatRoom.ID = primitive.NewObjectID().Hex()
	ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second)
	defer cancel()
	_, err := repo.collection.InsertOne(ctx, chatRoom)
	return err
}
