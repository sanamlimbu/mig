package repository

import (
	"context"
	"errors"
	"fmt"
	"mig"
	"mig/db"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
)

type UserRepository interface {
	// GetChatroomsByCreatorID returns chatrooms created by specified user.
	// Returned chatrooms are filtered based on given chatroom workflow states and paginated.
	GetChatroomsByCreatorID(ctx context.Context, creatorID string, chatroomWorkflowStates []string, pagination mig.Pagination) ([]mig.Chatroom, error)

	// GetFriendsByFriendshipWorkflowStates returns active friends of specified user.
	// Result is filtered based on provided friendship workflow states.
	// Returned result is paginated.
	GetFriendsByFriendshipWorkflowStates(ctx context.Context, userID string, friendshipWorkflowStates []string, pagination mig.Pagination) ([]mig.User, error)

	// GetFriend returns a friend for specified user, identified by their userID and friendID.
	GetFriend(ctx context.Context, userID, friendID string) (mig.User, error)

	// GetUser returns user for specified user ID.
	GetUser(ctx context.Context, userID string) (mig.User, error)

	// GetPrivateConversation returns private messages exchanged between specified users.
	// Returned messages are paginated.
	GetPrivateConversation(ctx context.Context, firstUserID, secondUserID string, pagination mig.Pagination) ([]mig.PrivateMessage, error)

	// GetUserByEmail returns user for specified email address.
	GetUserByEmail(ctx context.Context, email string) (mig.User, error)

	// GetUserByUsername returns user for specified username.
	GetUserByUsername(ctx context.Context, username string) (mig.User, error)

	// GetPassword returns password for specified user id.
	GetPassword(ctx context.Context, userID string) (string, error)

	// CreateUserWithRefreshToken creates user and refresh token.
	CreateUserWithRefreshToken(ctx context.Context, arg CreateUserWithRefreshTokenParams) (mig.User, mig.RefreshToken, error)

	// UpsertRefreshToken creates or updates refresh token.
	UpsertRefreshToken(ctx context.Context, arg UpsertRefreshTokenParams) (mig.RefreshToken, error)
}

type UserRepositoryPostgreSQL struct {
	conn    *pgx.Conn
	queries *db.Queries
}

func NewUserRepositoryPostgreSQL(conn *pgx.Conn, queries *db.Queries) (*UserRepositoryPostgreSQL, error) {
	if conn == nil {
		return nil, fmt.Errorf("missing conn")
	}

	if queries == nil {
		return nil, fmt.Errorf("missing queries")
	}

	repo := &UserRepositoryPostgreSQL{
		conn:    conn,
		queries: queries,
	}

	return repo, nil
}

func getUserFromDBModel(user db.User) mig.User {
	return mig.User{
		ID:            UUIDToString(user.ID),
		Email:         user.Email,
		Username:      user.Username,
		WorkflowState: mig.UserWorkflowState(user.WorkflowState),
	}
}

func getUsersFromDBModel(users []db.User) []mig.User {
	result := make([]mig.User, len(users))

	for i, user := range users {
		result[i] = getUserFromDBModel(user)
	}

	return result
}

func (r *UserRepositoryPostgreSQL) GetChatroomsByCreatorID(ctx context.Context, creatorID string, chatroomWorkflowStates []string, pagination mig.Pagination) ([]mig.Chatroom, error) {
	creatorUUID, err := StringToUUID(creatorID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid %s", creatorID)
	}

	arg := db.GetChatroomsByCreatorIDAndWorkflowStatesParams{
		ID:             creatorUUID,
		WorkflowStates: getDbModelChatroomWorkflowStates(chatroomWorkflowStates),
		Page:           int32(pagination.Page),
		PageSize:       int32(pagination.PageSize),
	}

	result, err := r.queries.GetChatroomsByCreatorIDAndWorkflowStates(ctx, arg)
	if err != nil {
		return nil, err
	}

	return getChatroomsFromDBModel(result), nil
}

func getDbModelFriendshipsWorkflowStates(input []string) []db.FriendshipWorkflowState {
	result := make([]db.FriendshipWorkflowState, len(input))

	for i, str := range input {
		result[i] = db.FriendshipWorkflowState(str)
	}

	return result
}

func (r *UserRepositoryPostgreSQL) GetFriendsByFriendshipWorkflowStates(ctx context.Context, userID string, friendshipWorkflowStates []string, pagination mig.Pagination) ([]mig.User, error) {
	userUUID, err := StringToUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid %s", userID)
	}

	arg := db.GetFriendsByFriendshipWorkflowStatesParams{
		ID:                       userUUID,
		FriendshipWorkflowStates: getDbModelFriendshipsWorkflowStates(friendshipWorkflowStates),
		Page:                     int32(pagination.Page),
		PageSize:                 int32(pagination.PageSize),
	}

	result, err := r.queries.GetFriendsByFriendshipWorkflowStates(ctx, arg)
	if err != nil {
		return nil, err
	}

	return getUsersFromDBModel(result), nil
}

func (r *UserRepositoryPostgreSQL) GetFriend(ctx context.Context, userID, friendID string) (mig.User, error) {
	userUUID, err := StringToUUID(userID)
	if err != nil {
		return mig.User{}, fmt.Errorf("invalid uuid %s", userID)
	}

	friendUUID, err := StringToUUID(friendID)
	if err != nil {
		return mig.User{}, fmt.Errorf("invalid uuid %s", friendID)
	}

	arg := db.GetFriendParams{
		UserID:   userUUID,
		FriendID: friendUUID,
	}

	result, err := r.queries.GetFriend(ctx, arg)
	if err != nil {
		return mig.User{}, err
	}

	return getUserFromDBModel(result), nil
}

func (r *UserRepositoryPostgreSQL) GetUser(ctx context.Context, userID string) (mig.User, error) {
	userUUID, err := StringToUUID(userID)
	if err != nil {
		return mig.User{}, fmt.Errorf("invalid uuid %s", userID)
	}

	result, err := r.queries.GetUser(ctx, userUUID)
	if err != nil {
		return mig.User{}, err
	}

	return getUserFromDBModel(result), nil
}

func (r *UserRepositoryPostgreSQL) GetUserByEmail(ctx context.Context, email string) (mig.User, error) {
	result, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return mig.User{}, err
	}

	return getUserFromDBModel(result), nil
}

func (r *UserRepositoryPostgreSQL) GetUserByUsername(ctx context.Context, username string) (mig.User, error) {
	result, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		return mig.User{}, err
	}

	return getUserFromDBModel(result), nil
}

func getPrivateMessageFromDBModel(msg db.GetPrivateConversationRow) mig.PrivateMessage {
	return mig.PrivateMessage{
		ID:                     UUIDToString(msg.ID),
		Content:                msg.Content,
		WorkflowState:          mig.MessageWorkflowState(msg.WorkflowState),
		Type:                   mig.MessageType(msg.MessageType),
		CreatedAt:              msg.CreatedAt.Time,
		SenderID:               UUIDToString(msg.SenderID),
		RecipientID:            UUIDToString(msg.RecipientID),
		SenderUsername:         msg.SenderUsername,
		SenderEmail:            msg.SenderEmail,
		SenderWorkflowState:    mig.UserWorkflowState(msg.SenderWorkflowState),
		RecipientUsername:      msg.RecipientUsername,
		RecipientEmail:         msg.RecipientEmail,
		RecipientWorkflowState: mig.UserWorkflowState(msg.RecipientWorkflowState),
	}
}

func getPrivateMessagesFromDBModel(msgs []db.GetPrivateConversationRow) []mig.PrivateMessage {
	result := make([]mig.PrivateMessage, len(msgs))

	for i, msg := range msgs {
		result[i] = getPrivateMessageFromDBModel(msg)
	}

	return result
}

func (r *UserRepositoryPostgreSQL) GetPrivateConversation(ctx context.Context, firstUserID, secondUserID string, pagination mig.Pagination) ([]mig.PrivateMessage, error) {
	firstUserUUID, err := StringToUUID(firstUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid %s", firstUserID)
	}

	secondUserUUID, err := StringToUUID(secondUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid %s", secondUserID)
	}

	arg := db.GetPrivateConversationParams{
		FirstUserID:  firstUserUUID,
		SecondUserID: secondUserUUID,
		Page:         int32(pagination.Page),
		PageSize:     int32(pagination.PageSize),
	}

	result, err := r.queries.GetPrivateConversation(ctx, arg)
	if err != nil {
		return nil, err
	}

	return getPrivateMessagesFromDBModel(result), nil
}

func (r *UserRepositoryPostgreSQL) GetPassword(ctx context.Context, userID string) (string, error) {
	uuid, err := StringToUUID(userID)
	if err != nil {
		return "", err
	}

	return r.queries.GetPassword(ctx, uuid)
}

func getRefreshTokenFromDBModel(refreshToken db.RefreshToken) mig.RefreshToken {
	return mig.RefreshToken{
		ID:        UUIDToString(refreshToken.ID),
		UserID:    UUIDToString(refreshToken.UserID),
		Token:     refreshToken.Token,
		ExpiresAt: refreshToken.ExpiresAt.Time,
		Revoked:   refreshToken.Revoked.Bool,
	}
}

type CreateUserWithRefreshTokenParams struct {
	ID            string
	Email         string
	Username      string
	WorkflowState mig.UserWorkflowState
	Password      string
	RefreshToken  string
	ExpiresAt     time.Time
}

func (r *UserRepositoryPostgreSQL) CreateUserWithRefreshToken(ctx context.Context, arg CreateUserWithRefreshTokenParams) (mig.User, mig.RefreshToken, error) {
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return mig.User{}, mig.RefreshToken{}, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Error().Msg(err.Error())
		}
	}()

	qtx := r.queries.WithTx(tx)

	uuid, err := StringToUUID(arg.ID)
	if err != nil {
		return mig.User{}, mig.RefreshToken{}, err
	}

	user, err := qtx.CreateUser(ctx, db.CreateUserParams{
		ID:            uuid,
		Email:         arg.Email,
		Username:      arg.Username,
		Password:      arg.Password,
		WorkflowState: db.UserWorkflowState(arg.WorkflowState),
	})
	if err != nil {
		return mig.User{}, mig.RefreshToken{}, err
	}

	refreshToken, err := qtx.UpsertRefreshToken(ctx, db.UpsertRefreshTokenParams{
		UserID: uuid,
		Token:  arg.RefreshToken,
		ExpiresAt: pgtype.Timestamp{
			Time:  arg.ExpiresAt,
			Valid: true,
		},
		Revoked: pgtype.Bool{
			Bool:  false,
			Valid: true,
		},
	})
	if err != nil {
		return mig.User{}, mig.RefreshToken{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return mig.User{}, mig.RefreshToken{}, err
	}

	return getUserFromDBModel(user), getRefreshTokenFromDBModel(refreshToken), nil
}

type UpsertRefreshTokenParams struct {
	UserID    string
	Token     string
	ExpiresAt time.Time
	Revoked   bool
}

func (r *UserRepositoryPostgreSQL) UpsertRefreshToken(ctx context.Context, arg UpsertRefreshTokenParams) (mig.RefreshToken, error) {

	uuid, err := StringToUUID(arg.UserID)
	if err != nil {
		return mig.RefreshToken{}, err

	}
	refreshToken, err := r.queries.UpsertRefreshToken(ctx, db.UpsertRefreshTokenParams{
		UserID: uuid,
		Token:  arg.Token,
		ExpiresAt: pgtype.Timestamp{
			Time:  arg.ExpiresAt,
			Valid: true,
		},
		Revoked: pgtype.Bool{
			Bool:  false,
			Valid: true,
		},
	})

	if err != nil {
		return mig.RefreshToken{}, err
	}

	return getRefreshTokenFromDBModel(refreshToken), nil
}
