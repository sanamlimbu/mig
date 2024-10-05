package mig

import (
	"context"
)

type UsersService struct {
	usersRepo UsersRepository
}

func NewUsersService(usersRepo UsersRepository) *UsersService {
	return &UsersService{
		usersRepo: usersRepo,
	}
}

type Friendship struct {
	ID                  int64  `json:"id"`
	RequesterID         int64  `json:"requester_id"`
	UserID              int64  `json:"user_id"`
	WorkflowState       string `json:"workflow_state"`
	WorkflowCompletedBy int64  `json:"workflow_completed_by"`
}

func (s *UsersService) GetChatroomsByCreatorID(creatorID int64, states []string, pagination Pagination) ([]Chatroom, error) {
	return s.usersRepo.getChatroomsByCreatorIDAndWorkflowStates(context.Background(), creatorID, states, pagination)
}

func (s *UsersService) GetFriends(userID int64, states []string, pagination Pagination) ([]User, error) {
	return s.usersRepo.getFriendsByWorkflowStates(context.Background(), userID, states, pagination)
}
