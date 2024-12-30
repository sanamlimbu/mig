package seed

import (
	"context"
	"errors"
	"mig/db"
	"mig/repository"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

// Messages creates two messages for each user in each chatroom and also creates messages
// between all unique pairs of users.
func (s *SeederPostgreSQL) Messages(ctx context.Context, userUUIDs, chatroomUUIDs []string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Error().Msg(err.Error())
		}
	}()

	qtx := s.queries.WithTx(tx)

	for _, userUUID := range userUUIDs {
		for _, chatroomUUID := range chatroomUUIDs {
			senderID, err := repository.StringToUUID(userUUID)
			if err != nil {
				return err
			}

			chatroomID, err := repository.StringToUUID(chatroomUUID)
			if err != nil {
				return err
			}

			msgArgs := []db.CreateChatroomMessageParams{
				{
					SenderID:      senderID,
					ChatroomID:    chatroomID,
					Content:       s.faker.Paragraph(1, 2, 10, ","),
					WorkflowState: db.MessageWorkflowStateCreated,
				},
				{
					SenderID:      senderID,
					ChatroomID:    chatroomID,
					Content:       s.faker.Paragraph(1, 2, 10, ","),
					WorkflowState: db.MessageWorkflowStateCreated,
				},
			}

			for _, msgArg := range msgArgs {
				_, err = qtx.CreateChatroomMessage(ctx, msgArg)
				if err != nil {
					return err
				}
			}
		}
	}

	for i := 0; i < len(userUUIDs)-1; i++ {
		for j := i + 1; j < len(userUUIDs); j++ {
			senderID, err := repository.StringToUUID(UsersUUIDs[i])
			if err != nil {
				return err
			}

			recipientID, err := repository.StringToUUID(userUUIDs[j])
			if err != nil {
				return err
			}

			msgArg := db.CreatePrivateMessageParams{
				SenderID:      senderID,
				RecipientID:   recipientID,
				Content:       s.faker.Paragraph(1, 2, 10, ","),
				WorkflowState: db.MessageWorkflowStateCreated,
			}

			_, err = qtx.CreatePrivateMessage(ctx, msgArg)
			if err != nil {
				return err
			}
		}
	}

	for i := len(userUUIDs) - 1; i >= 1; i-- {
		for j := i - 1; j >= 0; j-- {
			senderID, err := repository.StringToUUID(userUUIDs[i])
			if err != nil {
				return err
			}

			recipientID, err := repository.StringToUUID(userUUIDs[j])
			if err != nil {
				return err
			}

			msgArg := db.CreatePrivateMessageParams{
				SenderID:      senderID,
				RecipientID:   recipientID,
				Content:       s.faker.Paragraph(1, 2, 10, ","),
				WorkflowState: db.MessageWorkflowStateCreated,
			}

			_, err = qtx.CreatePrivateMessage(ctx, msgArg)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}
