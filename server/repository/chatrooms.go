package repository

import (
	"context"
	"fmt"
	"mig"
	"mig/db"

	"github.com/guregu/null/v5"
)

type ChatroomRepository interface {
	// GetChatroomsBySearchTermAndWorkflowStates returns chatrooms based on a search term and chatroom workflow states.
	// Returned result is paginated.
	GetChatroomsBySearchTermAndWorkflowStates(ctx context.Context, searchTerm string, chatroomWorkflowStates []string, pagination mig.Pagination) ([]mig.Chatroom, error)

	// GetMessages returns paginated messages for given chatroom.
	GetMessages(ctx context.Context, chatroomID string, pagination mig.Pagination) ([]mig.ChatroomMessage, error)

	// GetChatroom returns chatroom with given chatroom ID.
	GetChatroom(ctx context.Context, chatroomID string) (mig.Chatroom, error)

	// GetChatroomWithCreator returns chatroom with given chatroom ID including creator information.
	GetChatroomWithCreator(ctx context.Context, chatroomID string) (mig.ChatroomWithCreator, error)

	SaveChatroomMessage(ctx context.Context, arg SaveChatroomMessageParams) (mig.Message, error)
}

type ChatroomRepositoryPostgreSQL struct {
	queries *db.Queries
}

func NewChatroomRepositoryPostgreSQL(queries *db.Queries) (*ChatroomRepositoryPostgreSQL, error) {
	if queries == nil {
		return nil, fmt.Errorf("missing queries")
	}

	repo := &ChatroomRepositoryPostgreSQL{
		queries: queries,
	}

	return repo, nil
}

func getChatroomFromDBModel(chatroom db.Chatroom) mig.Chatroom {
	return mig.Chatroom{
		ID:            UUIDToString(chatroom.ID),
		Name:          chatroom.Name,
		WorkflowState: mig.ChatroomWorkflowState(chatroom.WorkflowState),
		Type:          mig.ChatroomType(chatroom.Type),
		CreatedBy:     UUIDToString(chatroom.CreatedBy),
	}
}

func getChatroomWithCreatorFromDBModel(chatroom db.GetChatroomWithCreatorRow) mig.ChatroomWithCreator {
	return mig.ChatroomWithCreator{
		ID:                   UUIDToString(chatroom.ID),
		Name:                 chatroom.Name,
		WorkflowState:        mig.ChatroomWorkflowState(chatroom.WorkflowState),
		Type:                 mig.ChatroomType(chatroom.Type),
		CreatedBy:            UUIDToString(chatroom.CreatedBy),
		CreatorUsername:      chatroom.CreatorUsername,
		CreatorEmail:         chatroom.CreatorEmail,
		CreatorWorkflowState: mig.UserWorkflowState(chatroom.CreatorWorkflowState),
	}
}

func getDbModelChatroomWorkflowStates(input []string) []db.ChatroomWorkflowState {
	result := make([]db.ChatroomWorkflowState, len(input))

	for i, str := range input {
		result[i] = db.ChatroomWorkflowState(str)
	}

	return result
}

func getChatroomsFromDBModel(chatrooms []db.Chatroom) []mig.Chatroom {
	result := make([]mig.Chatroom, len(chatrooms))

	for i, chatroom := range chatrooms {
		result[i] = getChatroomFromDBModel(chatroom)
	}

	return result
}

func getChatroomMessageFromDBModel(msg db.GetChatroomMessagesRow) mig.ChatroomMessage {
	return mig.ChatroomMessage{
		ID:                    UUIDToString(msg.ID),
		Content:               msg.Content,
		WorkflowState:         mig.MessageWorkflowState(msg.WorkflowState),
		Type:                  mig.MessageTypeChatroom,
		CreatedAt:             msg.CreatedAt.Time,
		SenderID:              UUIDToString(msg.SenderID),
		ChatroomID:            UUIDToString(msg.ChatroomID),
		ChatroomCreatorID:     UUIDToString(msg.ChatroomCreatorID),
		SenderEmail:           msg.SenderEmail,
		SenderUsername:        msg.SenderUsername,
		SenderWorkflowState:   mig.UserWorkflowState(msg.SenderWorkflowState),
		ChatroomName:          msg.ChatroomName,
		ChatroomWorkflowState: mig.ChatroomWorkflowState(msg.ChatroomWorkflowState),
	}
}

func getChatroomMessagesFromDBModel(msgs []db.GetChatroomMessagesRow) []mig.ChatroomMessage {
	result := make([]mig.ChatroomMessage, len(msgs))

	for i, msg := range msgs {
		result[i] = getChatroomMessageFromDBModel(msg)
	}

	return result
}

func (r *ChatroomRepositoryPostgreSQL) GetChatroomsBySearchTermAndWorkflowStates(ctx context.Context, searchTerm string, chatroomWorkflowStates []string, pagination mig.Pagination) ([]mig.Chatroom, error) {
	arg := db.GetChatroomsBySearchTermAndWorkflowStatesParams{
		SearchTerm:     searchTerm,
		WorkflowStates: getDbModelChatroomWorkflowStates(chatroomWorkflowStates),
		Page:           int32(pagination.Page),
		PageSize:       int32(pagination.PageSize),
	}

	result, err := r.queries.GetChatroomsBySearchTermAndWorkflowStates(ctx, arg)
	if err != nil {
		return nil, err
	}

	return getChatroomsFromDBModel(result), nil
}

func (r *ChatroomRepositoryPostgreSQL) GetMessages(ctx context.Context, chatroomID string, pagination mig.Pagination) ([]mig.ChatroomMessage, error) {
	chatroomUUID, err := StringToUUID(chatroomID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid %s", chatroomID)
	}

	result, err := r.queries.GetChatroomMessages(ctx, chatroomUUID)
	if err != nil {
		return nil, err
	}

	return getChatroomMessagesFromDBModel(result), nil
}

func (r *ChatroomRepositoryPostgreSQL) GetChatroom(ctx context.Context, chatroomID string) (mig.Chatroom, error) {
	chatroomUUID, err := StringToUUID(chatroomID)
	if err != nil {
		return mig.Chatroom{}, fmt.Errorf("invalid uuid %s", chatroomID)
	}

	result, err := r.queries.GetChatroom(ctx, chatroomUUID)
	if err != nil {
		return mig.Chatroom{}, err
	}

	return getChatroomFromDBModel(result), nil
}

func (r *ChatroomRepositoryPostgreSQL) GetChatroomWithCreator(ctx context.Context, chatroomID string) (mig.ChatroomWithCreator, error) {
	chatroomUUID, err := StringToUUID(chatroomID)
	if err != nil {
		return mig.ChatroomWithCreator{}, fmt.Errorf("invalid uuid %s", chatroomID)
	}

	result, err := r.queries.GetChatroomWithCreator(ctx, chatroomUUID)
	if err != nil {
		return mig.ChatroomWithCreator{}, err
	}

	return getChatroomWithCreatorFromDBModel(result), nil
}

type SaveChatroomMessageParams struct {
	ID         string
	SenderID   string
	ChatroomID string
	Content    string
}

func (r *ChatroomRepositoryPostgreSQL) SaveChatroomMessage(ctx context.Context, arg SaveChatroomMessageParams) (mig.Message, error) {
	id, err := StringToUUID(arg.ID)
	if err != nil {
		return mig.Message{}, err

	}

	senderID, err := StringToUUID(arg.SenderID)
	if err != nil {
		return mig.Message{}, err

	}

	chatroomID, err := StringToUUID(arg.ChatroomID)
	if err != nil {
		return mig.Message{}, err

	}

	msg, err := r.queries.CreateChatroomMessage(ctx, db.CreateChatroomMessageParams{
		ID:         id,
		SenderID:   senderID,
		ChatroomID: chatroomID,
		Content:    arg.Content,
	})

	if err != nil {
		return mig.Message{}, err
	}

	return mig.Message{
		ID:            UUIDToString(msg.ID),
		Content:       msg.Content,
		WorkflowState: mig.MessageWorkflowState(msg.WorkflowState),
		Type:          mig.MessageType(msg.MessageType),
		SenderID:      UUIDToString(msg.SenderID),
		RecipientID:   null.NewString(UUIDToString(msg.RecipientID), msg.RecipientID.Valid),
		ChatroomID:    null.NewString(UUIDToString(msg.ChatroomID), msg.ChatroomID.Valid),
		IsRead:        null.NewBool(msg.IsRead.Bool, msg.IsRead.Valid),
		CreatedAt:     msg.CreatedAt.Time,
	}, nil
}
