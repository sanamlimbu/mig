package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mig"
	"mig/repository"
)

type Service struct {
	userRepo repository.UserRepository
}

func NewService(userRepo repository.UserRepository) (*Service, error) {
	if userRepo == nil {
		return nil, fmt.Errorf("missing user repository")
	}

	service := &Service{
		userRepo: userRepo,
	}

	return service, nil
}

// GetChatroomsByCreatorID returns chatrooms created by given user.
// Chatrooms are filtered based on given chatroom workflow states.
// Result is paginated.
func (s *Service) GetChatroomsByCreatorID(ctx context.Context, creatorID string, chatroomWorkflowStates []string, pagination mig.Pagination) ([]mig.Chatroom, error) {
	_, err := s.userRepo.GetUser(ctx, creatorID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, mig.NewError(fmt.Sprintf("not found user of id %s", creatorID), err, mig.NotFoundError)
	}

	if err != nil {
		return nil, mig.NewError(fmt.Sprintf("unable to fetch user of id %s", creatorID), err, mig.InternalServerError)
	}

	result, err := s.userRepo.GetChatroomsByCreatorID(ctx, creatorID, chatroomWorkflowStates, pagination)
	if err != nil {
		return nil, mig.NewError(fmt.Sprintf("unable to fetch chatrooms created by user id %s", creatorID), err, mig.InternalServerError)
	}

	return result, nil
}

// GetFriends returns active friends of given user.
// Result is paginated.
func (s *Service) GetFriends(ctx context.Context, userID string, pagination mig.Pagination) ([]mig.User, error) {
	_, err := s.userRepo.GetUser(ctx, userID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, mig.NewError(fmt.Sprintf("not found user of id %s", userID), err, mig.NotFoundError)
	}

	if err != nil {
		return nil, mig.NewError(fmt.Sprintf("unable to fetch user of id %s", userID), err, mig.InternalServerError)
	}

	result, err := s.userRepo.GetFriends(ctx, userID, pagination)
	if err != nil {
		return nil, mig.NewError(fmt.Sprintf("unable to fetch friends of user of id %s", userID), err, mig.InternalServerError)
	}

	return result, nil
}

// GetPrivateMessages returns messages communicated between two given users.
// Result is paginated.
func (s *Service) GetPrivateMessages(ctx context.Context, firstUserID, secondUserID string, pagination mig.Pagination) ([]mig.PrivateMessage, error) {
	_, err := s.userRepo.GetUser(ctx, firstUserID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, mig.NewError(fmt.Sprintf("not found user of id %s", firstUserID), err, mig.NotFoundError)
	}

	if err != nil {
		return nil, mig.NewError(fmt.Sprintf("unable to fetch user of id %s", firstUserID), err, mig.InternalServerError)
	}

	_, err = s.userRepo.GetUser(ctx, secondUserID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, mig.NewError(fmt.Sprintf("not found user of id %s", secondUserID), err, mig.NotFoundError)
	}

	if err != nil {
		return nil, mig.NewError(fmt.Sprintf("unable to fetch user of id %s", secondUserID), err, mig.InternalServerError)
	}

	result, err := s.userRepo.GetPrivateConversation(ctx, firstUserID, secondUserID, pagination)
	if err != nil {
		return nil, mig.NewError(fmt.Sprintf("unable to fetch private messages between users of ids %s and %s", firstUserID, secondUserID), err, mig.InternalServerError)
	}

	return result, nil
}

// GetUser returns a user with given id.
func (s *Service) GetUser(ctx context.Context, userID string) (mig.User, error) {
	result, err := s.userRepo.GetUser(ctx, userID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return mig.User{}, mig.NewError(fmt.Sprintf("not found user of id %s", userID), err, mig.NotFoundError)
	}

	if err != nil {
		return mig.User{}, mig.NewError(fmt.Sprintf("unable to fetch user of id %s", userID), err, mig.InternalServerError)
	}

	return result, nil
}
