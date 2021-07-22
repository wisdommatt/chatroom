package http

import (
	"strings"

	"github.com/go-playground/validator"
)

type createChatRoomPayload struct {
	Name       string `json:"name" validate:"required"`
	InviteCode string `json:"inviteCode"`
}

func (entity *createChatRoomPayload) validate() error {
	entity.InviteCode = strings.ToLower(entity.InviteCode)
	return validator.New().Struct(entity)
}
