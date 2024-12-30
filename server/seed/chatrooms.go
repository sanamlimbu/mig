package seed

import (
	"context"
	"database/sql"
	"errors"
	"mig"
	"mig/db"
	"mig/repository"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

// Chatrooms will create one public chatroom per user as creator.
func (s *SeederPostgreSQL) Chatrooms(ctx context.Context, userUUIDs, chatroomUUIDs []string) ([]mig.Chatroom, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Error().Msg(err.Error())
		}
	}()

	qtx := s.queries.WithTx(tx)

	chatrooms := make([]mig.Chatroom, len(userUUIDs))

	for i, userUUID := range userUUIDs {
		name := ""
		for {
			name = strings.ToLower(s.faker.Animal())

			_, err := s.queries.GetChatroomByName(ctx, name)
			if err != nil && errors.Is(err, sql.ErrNoRows) {
				break
			}

			if err != nil {
				return nil, err
			}
		}

		createdBy, err := repository.StringToUUID(userUUID)
		if err != nil {
			return nil, err
		}

		chatroomUUID, err := repository.StringToUUID(chatroomUUIDs[i])
		if err != nil {
			return nil, err
		}

		chatroomArgs := db.CreateChatroomParams{
			ID:            chatroomUUID,
			Name:          name,
			WorkflowState: db.ChatroomWorkflowStateActive,
			Type:          db.ChatroomTypePublic,
			CreatedBy:     createdBy,
		}

		chatroom, err := qtx.CreateChatroom(ctx, chatroomArgs)
		if err != nil {
			return nil, err
		}

		chatrooms[i] = mig.Chatroom{
			ID:            repository.UUIDToString(chatroom.ID),
			Name:          chatroom.Name,
			WorkflowState: mig.ChatroomWorkflowState(chatroom.WorkflowState),
			Type:          mig.ChatroomType(chatroom.Type),
			CreatedBy:     repository.UUIDToString(chatroom.CreatedBy),
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return chatrooms, nil
}
