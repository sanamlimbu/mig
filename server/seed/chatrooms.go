package seed

import (
	"context"
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

	chatroomNames := make(map[string]bool, len(userUUIDs))

	for i, userUUID := range userUUIDs {
		var name string
		for {
			name = strings.ToLower(s.faker.Animal())

			if !chatroomNames[name] {
				chatroomNames[name] = true
				break
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
