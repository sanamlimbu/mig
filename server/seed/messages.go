package seed

import (
	"context"
	"mig"
	"mig/db"
	"mig/repository"
)

// Messages creates two messages for each user in each chatroom and also creates messages
// between all unique pairs of users.
func (s *SeederPostgreSQL) Messages(ctx context.Context, users []mig.User, chatrooms []mig.Chatroom) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	for _, user := range users {
		for _, chatroom := range chatrooms {
			senderID, err := repository.StringToUUID(user.ID)
			if err != nil {
				return err
			}

			chatroomID, err := repository.StringToUUID(chatroom.ID)
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

	for i := 0; i < len(users)-1; i++ {
		for j := i + 1; j < len(users); j++ {
			senderID, err := repository.StringToUUID(users[i].ID)
			if err != nil {
				return err
			}

			recipientID, err := repository.StringToUUID(users[j].ID)
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

	for i := len(users) - 1; i >= 1; i-- {
		for j := i - 1; j >= 0; j-- {
			senderID, err := repository.StringToUUID(users[i].ID)
			if err != nil {
				return err
			}

			recipientID, err := repository.StringToUUID(users[j].ID)
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
