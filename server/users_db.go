package mig

import (
	"context"
	"database/sql"
	"mig/models"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type User struct {
	ID            int64  `json:"id"`
	UUID          string `json:"uuid"`
	Username      string `json:"username"`
	WorkflowState string `json:"workflow_state"`
}

func UserFromDBModel(u models.User) User {
	return User{
		ID:            u.ID,
		UUID:          u.UUID,
		Username:      u.Username,
		WorkflowState: u.WorkflowState.String(),
	}
}

func UsersFromDBModel(users models.UserSlice) []User {
	results := []User{}

	for _, u := range users {
		results = append(results, UserFromDBModel(*u))
	}

	return results
}

func FriendshipsWorkflowStates(states []string) []models.FriendshipsWorkflowState {
	ws := []models.FriendshipsWorkflowState{}

	for _, state := range states {
		ws = append(ws, models.FriendshipsWorkflowState(state))
	}

	return ws
}

type UsersRepository interface {
	getChatroomsByCreatorIDAndWorkflowStates(ctx context.Context, creatorID int64, states []string, paination Pagination) ([]Chatroom, error)
	getFriendsByWorkflowStates(ctx context.Context, userID int64, states []string, pagination Pagination) ([]User, error)
}

type UsersRepositoryPostgreSQL struct {
	db *sql.DB
}

func NewUsersRepositoryPostgreSQL(db *sql.DB) *UsersRepositoryPostgreSQL {
	return &UsersRepositoryPostgreSQL{
		db: db,
	}
}

func (r *UsersRepositoryPostgreSQL) getChatroomsByCreatorIDAndWorkflowStates(ctx context.Context, creatorID int64, states []string, pagination Pagination) ([]Chatroom, error) {
	ws := ChatroomsWorkflowStates(states)

	chatrooms, err := models.Chatrooms(
		models.ChatroomWhere.CreatedBy.EQ(creatorID),
		models.ChatroomWhere.WorkflowState.IN(ws),
		qm.Limit(pagination.pageSize),
		qm.Offset(pagination.page),
	).All(ctx, r.db)

	if err != nil {
		return nil, err
	}

	return ChatroomsFromDBModel(chatrooms), nil
}

func (r *UsersRepositoryPostgreSQL) getFriendsByWorkflowStates(ctx context.Context, userID int64, states []string, pagination Pagination) ([]User, error) {
	ws := FriendshipsWorkflowStates(states)

	// avoid OR in INNER JOIN
	firstResults, err := models.Users(
		qm.InnerJoin("%s ON %s = %s", models.TableNames.Friendships, models.FriendshipColumns.RequesterID, models.UserColumns.ID),
		models.FriendshipWhere.WorkflowState.IN(ws),
		models.FriendshipWhere.RequesterID.EQ(userID),
		qm.Limit(pagination.pageSize),
	).All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	secondResults, err := models.Users(
		qm.InnerJoin("%s ON %s = %s", models.TableNames.Friendships, models.FriendshipColumns.UserID, models.UserColumns.ID),
		models.FriendshipWhere.WorkflowState.IN(ws),
		models.FriendshipWhere.UserID.EQ(userID),
		qm.Limit(pagination.pageSize),
	).All(ctx, r.db)
	if err != nil {
		return nil, err
	}

	results := append(firstResults, secondResults...)

	total := len(results)
	start := pagination.page
	end := pagination.page + pagination.pageSize

	if start >= total {
		return nil, nil
	}

	if end > total {
		end = total
	}

	return UsersFromDBModel(results[start:end]), nil
}
