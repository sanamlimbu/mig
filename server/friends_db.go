package mig

import (
	"context"
	"database/sql"
	"mig/models"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func getFriendsByWorkflowState(ctx context.Context, db *sql.DB, pagination Pagination, userID int64, state models.FriendshipsWorkflowState) ([]*models.User, error) {
	// avoid OR in INNER JOIN
	firstResults, err := models.Users(
		qm.InnerJoin(models.TableNames.Friendships+" ON "+models.FriendshipColumns.RequesterID+" = "+models.UserColumns.ID),
		models.FriendshipWhere.WorkflowState.EQ(state),
		models.FriendshipWhere.RequesterID.EQ(userID),
		qm.Limit(pagination.pageSize),
	).All(ctx, db)
	if err != nil {
		return nil, err
	}

	secondResults, err := models.Users(
		qm.InnerJoin(models.TableNames.Friendships+" ON "+models.FriendshipColumns.UserID+" = "+models.UserColumns.ID),
		models.FriendshipWhere.WorkflowState.EQ(state),
		models.FriendshipWhere.UserID.EQ(userID),
		qm.Limit(pagination.pageSize),
	).All(ctx, db)
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

	return results[start:end], nil
}
