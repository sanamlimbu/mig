package seed

import (
	"context"
	"errors"
	"mig/db"
	"mig/repository"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

// Friendships will create friendships between each users.
func (s *SeederPostgreSQL) Friendships(ctx context.Context, userUUIDs []string) error {
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

	for i := 0; i < len(userUUIDs)-1; i++ {
		for j := i + 1; j < len(userUUIDs); j++ {

			requestorID, err := repository.StringToUUID(userUUIDs[i])
			if err != nil {
				return err
			}

			userID, err := repository.StringToUUID(userUUIDs[j])
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
