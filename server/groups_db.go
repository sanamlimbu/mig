package mig

import (
	"context"
	"database/sql"
	"mig/models"
	"time"

	"github.com/guregu/null"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Group struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	WorkflowState string    `json:"workflow_state"`
	Type          string    `json:"type"`
	CreatedBy     int64     `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeletedAt     null.Time `json:"deleted_at"`
}

type GroupsRepository interface {
	getGroupsByWorflowStatesAndUserID(ctx context.Context, pagination Pagination, states []string, userID int64) ([]Group, error)
}

type GroupsRepositoryPostgreSQL struct {
	db *sql.DB
}

func NewGroupsRepositoryPostgreSQL(db *sql.DB) *GroupsRepositoryPostgreSQL {
	return &GroupsRepositoryPostgreSQL{
		db: db,
	}
}

func (r *GroupsRepositoryPostgreSQL) getGroupsByWorflowStatesAndUserID(ctx context.Context, pagination Pagination, states []string, userID int64) ([]Group, error) {
	ws := []models.GroupUsersWorkflowState{}

	for _, state := range states {
		ws = append(ws, models.GroupUsersWorkflowState(state))
	}
	groups, err := models.Groups(
		qm.InnerJoin(models.TableNames.GroupUsers+" ON "+models.GroupUserColumns.GroupID+" = "+models.GroupColumns.ID),
		models.GroupUserWhere.WorkflowState.IN(ws),
		qm.Or2(models.GroupUserWhere.RequesterID.EQ(userID)),
		qm.Or2(models.GroupUserWhere.UserID.EQ(userID)),
		qm.Limit(pagination.pageSize),
		qm.Offset(pagination.page),
	).All(ctx, r.db)

	results := []Group{}

	for _, g := range groups {
		results = append(results, Group{
			ID:            g.ID,
			Name:          g.Name,
			WorkflowState: string(g.WorkflowState),
			Type:          string(g.Type),
			CreatedBy:     g.CreatedBy,
			CreatedAt:     g.CreatedAt,
			UpdatedAt:     g.UpdatedAt,
			DeletedAt:     null.TimeFrom(g.DeletedAt.Time),
		})
	}

	return results, err
}
