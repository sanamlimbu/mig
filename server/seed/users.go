package seed

import (
	"context"
	"errors"
	"fmt"
	"mig"
	"mig/db"
	"mig/repository"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

// Users will create total of (count + 2) users.
// `count` number of user are random users.
// Two users Jack(username: jack, email: jack@example.com, password: jack123) and
// Rose(username: rose, email:rose@example.com, password: rose123) are also created.
func (s *SeederPostgreSQL) Users(ctx context.Context, uuids []string) ([]mig.User, error) {
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

	length := len(uuids)

	if length <= 2 {
		return nil, fmt.Errorf("uuids slice length is small")
	}

	users := make([]mig.User, length)

	for i := 0; i < len(uuids)-2; i++ {
		username := s.faker.Username()

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(username), 8)
		if err != nil {
			return nil, err
		}

		uuid, err := repository.StringToUUID(uuids[i])
		if err != nil {
			return nil, err
		}

		args := db.CreateUserParams{
			ID:            uuid,
			Username:      username,
			Email:         s.faker.Email(),
			Password:      string(passwordHash),
			WorkflowState: db.UserWorkflowStateActive,
		}

		user, err := qtx.CreateUser(ctx, args)
		if err != nil {
			return nil, err
		}

		users[i] = mig.User{
			ID:            repository.UUIDToString(user.ID),
			Email:         user.Email,
			Username:      user.Username,
			WorkflowState: mig.UserWorkflowState(user.WorkflowState),
		}
	}

	jackPasswordHash, err := bcrypt.GenerateFromPassword([]byte("jack123"), 8)
	if err != nil {
		return nil, err
	}

	jackUUID, err := repository.StringToUUID(UsersUUIDs[length-2])
	if err != nil {
		return nil, err
	}

	jackArgs := db.CreateUserParams{
		ID:            jackUUID,
		Username:      "jack",
		Email:         "jack@example.com",
		Password:      string(jackPasswordHash),
		WorkflowState: db.UserWorkflowStateActive,
	}

	jack, err := qtx.CreateUser(ctx, jackArgs)
	if err != nil {
		return nil, err
	}

	rosePasswordHash, err := bcrypt.GenerateFromPassword([]byte("rose123"), 8)
	if err != nil {
		return nil, err
	}

	roseUUID, err := repository.StringToUUID(UsersUUIDs[length-1])
	if err != nil {
		return nil, err
	}
	roseArgs := db.CreateUserParams{
		ID:            roseUUID,
		Username:      "rose",
		Email:         "rose@example.com",
		Password:      string(rosePasswordHash),
		WorkflowState: db.UserWorkflowStateActive,
	}

	rose, err := qtx.CreateUser(ctx, roseArgs)
	if err != nil {
		return nil, err
	}

	users[length-2] = mig.User{
		ID:            repository.UUIDToString(jack.ID),
		Email:         jack.Email,
		Username:      jack.Username,
		WorkflowState: mig.UserWorkflowState(jack.WorkflowState),
	}

	users[length-1] = mig.User{
		ID:            repository.UUIDToString(rose.ID),
		Email:         rose.Email,
		Username:      rose.Username,
		WorkflowState: mig.UserWorkflowState(rose.WorkflowState),
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return users, nil
}
