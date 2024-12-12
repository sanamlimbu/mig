package chatroom

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mig"
	"mig/repository"
)

type Service struct {
	chatroomRepo repository.ChatroomRepository
}

func NewService(chatroomRepo repository.ChatroomRepository) (*Service, error) {
	if chatroomRepo == nil {
		return nil, fmt.Errorf("missig chatroom repository")
	}

	service := &Service{
		chatroomRepo: chatroomRepo,
	}

	return service, nil
}

// GetChatroomWithCreator returns chatroom for given chatroom ID.
func (s *Service) GetChatroomWithCreator(ctx context.Context, chatroomID string) (mig.ChatroomWithCreator, error) {
	result, err := s.chatroomRepo.GetChatroomWithCreator(ctx, chatroomID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return mig.ChatroomWithCreator{}, mig.NewError(fmt.Sprintf("not found chatroom of id %s", chatroomID), err, mig.NotFoundError)
	}

	if err != nil {
		return mig.ChatroomWithCreator{}, mig.NewError(fmt.Sprintf("unable to fetch chatroom of id %s", chatroomID), err, mig.InternalServerError)
	}

	return result, nil
}

// GetChatroomsBySearchTermAndWorkflowStates returns chatrooms based on given search term and chatroom workflow states.
// Result is paginated.
func (s *Service) GetChatroomsBySearchTermAndWorkflowStates(ctx context.Context, searchTerm string, chatroomWorkflowStates []string, pagination mig.Pagination) ([]mig.Chatroom, error) {
	result, err := s.chatroomRepo.GetChatroomsBySearchTermAndWorkflowStates(ctx, searchTerm, chatroomWorkflowStates, pagination)
	if err != nil {
		return nil, mig.NewError("unable to fetch chatrooms", err, mig.InternalServerError)
	}

	return result, nil
}

// GetMessages returns messages sent in given chatroom.
// Returned messages are paginated.
func (s *Service) GetMessages(ctx context.Context, chatroomID string, pagination mig.Pagination) ([]mig.ChatroomMessage, error) {
	_, err := s.chatroomRepo.GetChatroom(ctx, chatroomID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, mig.NewError(fmt.Sprintf("not found chatroom of id %s", chatroomID), err, mig.NotFoundError)
	}

	if err != nil {
		return nil, mig.NewError(fmt.Sprintf("unable to fetch chatroom of id %s", chatroomID), err, mig.InternalServerError)
	}

	result, err := s.chatroomRepo.GetMessages(ctx, chatroomID, pagination)
	if err != nil {
		return nil, mig.NewError(fmt.Sprintf("unable to fetch messages of chatroom of id %s", chatroomID), err, mig.InternalServerError)
	}

	return result, nil
}
