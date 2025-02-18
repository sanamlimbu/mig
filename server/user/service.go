package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"mig"
	"mig/repository"

	"github.com/guregu/null/v5"
)

type Service struct {
	userRepo     repository.UserRepository
	chatroomRepo repository.ChatroomRepository
}

func NewService(userRepo repository.UserRepository, chatroomRepo repository.ChatroomRepository) (*Service, error) {
	if userRepo == nil {
		return nil, fmt.Errorf("missing user repository")
	}

	if chatroomRepo == nil {
		return nil, fmt.Errorf("missing chatroom repository")
	}

	service := &Service{
		userRepo:     userRepo,
		chatroomRepo: chatroomRepo,
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

// GetFriendsByFriendshipWorkflowStates returns active friends of given user based on given friendship workflow states.
// Result is paginated.
func (s *Service) GetFriendsByFriendshipWorkflowStates(ctx context.Context, userID string, friendshipWorkflowStates []string, pagination mig.Pagination) ([]mig.User, error) {
	_, err := s.userRepo.GetUser(ctx, userID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, mig.NewError(fmt.Sprintf("not found user of id %s", userID), err, mig.NotFoundError)
	}

	if err != nil {
		return nil, mig.NewError(fmt.Sprintf("unable to fetch user of id %s", userID), err, mig.InternalServerError)
	}

	result, err := s.userRepo.GetFriendsByFriendshipWorkflowStates(ctx, userID, friendshipWorkflowStates, pagination)
	if err != nil {
		return nil, mig.NewError(fmt.Sprintf("unable to fetch friends of user of id %s", userID), err, mig.InternalServerError)
	}

	return result, nil
}

// GetPrivateConversation returns messages communicated between two given users.
// Result is paginated.
func (s *Service) GetPrivateConversation(ctx context.Context, firstUserID, secondUserID string, pagination mig.Pagination) ([]mig.Message, error) {
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

// GetRecentPrivateMessages returns recent private messages of given user.
// Result is paginated.
func (s *Service) GetRecentPrivateMessages(ctx context.Context, userID string, pagination mig.Pagination) ([]mig.Message, error) {
	_, err := s.userRepo.GetUser(ctx, userID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, mig.NewError(fmt.Sprintf("not found user of id %s", userID), err, mig.NotFoundError)
	}

	if err != nil {
		return nil, mig.NewError(fmt.Sprintf("unable to fetch user of id %s", userID), err, mig.InternalServerError)
	}

	result, err := s.userRepo.GetRecentPrivateMessages(ctx, userID, pagination)
	if err != nil {
		return nil, mig.NewError(fmt.Sprintf("unable to fetch private messages of user id %s", userID), err, mig.InternalServerError)
	}

	return result, nil
}

type SaveMessageParams struct {
	ID          string
	SenderID    string
	RecipientID string
	Content     string
	Type        mig.MessageType
}

func (s *Service) SaveMessage(ctx context.Context, arg SaveMessageParams) error {
	if arg.Type == mig.MessageTypePrivate {
		_, err := s.SavePrivateMessage(ctx, SavePrivateMessageParams{
			ID:          arg.ID,
			SenderID:    arg.SenderID,
			RecipientID: arg.RecipientID,
			Content:     arg.Content,
		})

		return err

	} else if arg.Type == mig.MessageTypeChatroom {
		_, err := s.SaveChatroomMessage(ctx, SaveChatroomMessageParams{
			ID:         arg.ID,
			SenderID:   arg.SenderID,
			ChatroomID: arg.RecipientID,
			Content:    arg.Content,
		})

		return err
	}

	return fmt.Errorf("invalid message type %s", arg.Type)
}

type SavePrivateMessageParams struct {
	ID          string
	SenderID    string
	RecipientID string
	Content     string
}

func (s *Service) SavePrivateMessage(ctx context.Context, arg SavePrivateMessageParams) (mig.Message, error) {
	sender, err := s.userRepo.GetUser(ctx, arg.SenderID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return mig.Message{}, mig.NewError(fmt.Sprintf("not found user of id %s", arg.SenderID), err, mig.NotFoundError)
	}

	if err != nil {
		return mig.Message{}, mig.NewError(fmt.Sprintf("unable to fetch user of id %s", arg.SenderID), err, mig.InternalServerError)
	}

	recipient, err := s.userRepo.GetUser(ctx, arg.RecipientID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return mig.Message{}, mig.NewError(fmt.Sprintf("not found user of id %s", arg.RecipientID), err, mig.NotFoundError)
	}

	if err != nil {
		return mig.Message{}, mig.NewError(fmt.Sprintf("unable to fetch user of id %s", arg.RecipientID), err, mig.InternalServerError)
	}

	msg, err := s.userRepo.SavePrivateMessage(ctx, repository.SavePrivateMessageParams{
		ID:          arg.ID,
		SenderID:    arg.SenderID,
		RecipientID: arg.RecipientID,
		Content:     arg.Content,
	})

	if err != nil {
		return mig.Message{}, mig.NewError("unable to save private message", err, mig.InternalServerError)
	}

	return mig.Message{
		ID:            msg.ID,
		Content:       msg.Content,
		WorkflowState: msg.WorkflowState,
		Type:          msg.Type,
		CreatedAt:     msg.CreatedAt,
		SenderID:      msg.SenderID,
		RecipientID:   msg.RecipientID,
		Sender: mig.User{
			ID:            sender.ID,
			Email:         sender.Email,
			Username:      sender.Username,
			WorkflowState: sender.WorkflowState,
		},
		Recipient: mig.User{
			ID:            recipient.ID,
			Email:         recipient.Email,
			Username:      recipient.Username,
			WorkflowState: recipient.WorkflowState,
		},
	}, nil
}

type SaveChatroomMessageParams struct {
	ID         string
	SenderID   string
	ChatroomID string
	Content    string
}

func (s *Service) SaveChatroomMessage(ctx context.Context, arg SaveChatroomMessageParams) (mig.Message, error) {
	sender, err := s.userRepo.GetUser(ctx, arg.SenderID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return mig.Message{}, mig.NewError(fmt.Sprintf("not found user of id %s", arg.SenderID), err, mig.NotFoundError)
	}

	if err != nil {
		return mig.Message{}, mig.NewError(fmt.Sprintf("unable to fetch user of id %s", arg.SenderID), err, mig.InternalServerError)
	}

	chatroom, err := s.chatroomRepo.GetChatroom(ctx, arg.ChatroomID, false)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return mig.Message{}, mig.NewError(fmt.Sprintf("not found chatroom of id %s", arg.ChatroomID), err, mig.NotFoundError)
	}

	if err != nil {
		return mig.Message{}, mig.NewError(fmt.Sprintf("unable to fetch chatroom of id %s", arg.ChatroomID), err, mig.InternalServerError)
	}

	msg, err := s.chatroomRepo.SaveChatroomMessage(ctx, repository.SaveChatroomMessageParams{
		ID:         arg.ID,
		SenderID:   arg.SenderID,
		ChatroomID: arg.ChatroomID,
		Content:    arg.Content,
	})

	if err != nil {
		return mig.Message{}, mig.NewError("unable to save chatroom message", err, mig.InternalServerError)
	}

	return mig.Message{
		ID:            msg.ID,
		Content:       msg.Content,
		WorkflowState: msg.WorkflowState,
		Type:          msg.Type,
		CreatedAt:     msg.CreatedAt,
		SenderID:      msg.SenderID,
		ChatroomID:    msg.ChatroomID,
		Sender: mig.User{
			ID:            sender.ID,
			Email:         sender.Email,
			Username:      sender.Username,
			WorkflowState: sender.WorkflowState,
		},
		Chatroom: mig.Chatroom{
			ID:            chatroom.ID,
			Name:          chatroom.Name,
			Type:          chatroom.Type,
			WorkflowState: chatroom.WorkflowState,
			CreatedBy:     chatroom.CreatedBy,
		},
	}, nil
}

type UpdateMessageParams struct {
	ID            string
	Content       string
	WorkflowState mig.MessageWorkflowState
	IsRead        null.Bool
}

func (s *Service) UpdateMessage(ctx context.Context, arg UpdateMessageParams) (mig.Message, error) {
	msg, err := s.userRepo.UpdateMessage(ctx, repository.UpdateMessageParams{
		ID:           arg.ID,
		Content:      arg.Content,
		WorflowState: arg.WorkflowState,
		IsRead:       arg.IsRead,
	})

	if err != nil {
		return mig.Message{}, mig.NewError("unable to update message", err, mig.InternalServerError)
	}

	return msg, nil
}

func (s *Service) DeleteMessage(ctx context.Context, id string) error {
	err := s.userRepo.DeleteMessage(ctx, id)

	if err != nil {
		return mig.NewError("unable to delete message", err, mig.InternalServerError)
	}

	return nil
}
