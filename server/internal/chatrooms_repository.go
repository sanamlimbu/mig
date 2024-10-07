package mig

import (
	"context"
	"database/sql"
	"fmt"
	"mig/models"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type ChatroomsRepository interface {
	getChatroomsBySearchTermAndWorkflowState(ctx context.Context, searchTerm string, states []string, pagination Pagination) ([]Chatroom, error)
}

type ChatroomsRepositoryPostgreSQL struct {
	db *sql.DB
}

func NewChatroomsRepositoryPostgreSQL(db *sql.DB) *ChatroomsRepositoryPostgreSQL {
	return &ChatroomsRepositoryPostgreSQL{
		db: db,
	}
}

func ChatroomFromDBModel(c models.Chatroom) Chatroom {
	return Chatroom{
		ID:            c.ID,
		Name:          c.Name,
		WorkflowState: c.WorkflowState.String(),
		Type:          c.Type.String(),
		CreatedBy:     c.CreatedBy,
	}
}

func ChatroomsFromDBModel(chatrooms models.ChatroomSlice) []Chatroom {
	results := []Chatroom{}

	for _, c := range chatrooms {
		results = append(results, ChatroomFromDBModel(*c))
	}

	return results
}

func ChatroomsWorkflowStates(states []string) []models.ChatroomsWorkflowState {
	ws := []models.ChatroomsWorkflowState{}

	for _, state := range states {
		ws = append(ws, models.ChatroomsWorkflowState(state))
	}

	return ws
}

func (r *ChatroomsRepositoryPostgreSQL) getChatroomsBySearchTermAndWorkflowState(ctx context.Context, searchTerm string, states []string, pagination Pagination) ([]Chatroom, error) {
	ws := ChatroomsWorkflowStates(states)

	searchPattern := fmt.Sprintf("%%%s%%", searchTerm)

	chatrooms, err := models.Chatrooms(
		models.ChatroomWhere.WorkflowState.IN(ws),
		qm.Where("? ILIKE ?", models.ChatroomColumns.Name, searchPattern),
		qm.Limit(pagination.pageSize),
		qm.Offset(pagination.page),
	).All(ctx, r.db)

	return ChatroomsFromDBModel(chatrooms), err
}
