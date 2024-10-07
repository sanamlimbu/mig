package mig

import (
	"context"
)

type User struct {
	ID            int64  `json:"id"`
	UUID          string `json:"uuid"`
	Username      string `json:"username"`
	WorkflowState string `json:"workflow_state"`
}

type UsersService struct {
	usersRepo UsersRepository
}

func NewUsersService(usersRepo UsersRepository) *UsersService {
	return &UsersService{
		usersRepo: usersRepo,
	}
}

func (s *UsersService) GetChatroomsByCreatorID(creatorID int64, states []string, pagination Pagination) ([]Chatroom, error) {
	return s.usersRepo.getChatroomsByCreatorIDAndWorkflowStates(context.Background(), creatorID, states, pagination)
}

func (s *UsersService) GetFriends(userID int64, states []string, pagination Pagination) ([]User, error) {
	return s.usersRepo.getFriendsByWorkflowStates(context.Background(), userID, states, pagination)
}
