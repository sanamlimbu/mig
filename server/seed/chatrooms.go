package seed

import (
	"context"
	"database/sql"
	"errors"
	"mig"
	"mig/db"
	"mig/repository"
	"strings"
)

// Chatrooms will create one public chatroom per user as creator.
func (s *SeederPostgreSQL) Chatrooms(ctx context.Context, users []mig.User) ([]mig.Chatroom, error) {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)
	chatrooms := make([]mig.Chatroom, len(users))

	for i, user := range users {
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

		createdBy, err := repository.StringToUUID(user.ID)
		if err != nil {
			return nil, err
		}

		chatroomArgs := db.CreateChatroomParams{
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
