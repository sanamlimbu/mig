package seed

import (
	"context"
	"errors"
	"mig"
	"mig/db"
	"mig/repository"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

// Friendships will create friendships between each users.
func (s *SeederPostgreSQL) Friendships(ctx context.Context, users []mig.User) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Error().Msg(err.Error())
		}
	}()

	qtx := s.queries.WithTx(tx)

	for i := 0; i < len(users)-1; i++ {
		for j := i + 1; j < len(users); j++ {

			requestorID, err := repository.StringToUUID(users[i].ID)
			if err != nil {
				return err
			}

			userID, err := repository.StringToUUID(users[j].ID)
			if err != nil {
				return err
			}

			friendshipArgs := db.UpsertFriendshipParams{
				RequesterID:         requestorID,
				UserID:              userID,
				WorkflowState:       db.FriendshipWorkflowStateActive,
				WorkflowCompletedBy: userID,
			}

			_, err = qtx.UpsertFriendship(ctx, friendshipArgs)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}
