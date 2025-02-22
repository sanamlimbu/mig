package repository

import (
	"context"
	"errors"
	"fmt"
	"mig"
	"mig/db"
	"time"

	"github.com/guregu/null/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
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
	GetPrivateConversation(ctx context.Context, firstUserID, secondUserID string, pagination mig.Pagination) ([]mig.Message, error)

	// GetRecentPrivateMessages returns recent private messages of specified user.
	// Returned messages are paginated.
	GetRecentPrivateMessages(ctx context.Context, userID string, pagination mig.Pagination) ([]mig.Message, error)

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

	GetRefreshToken(ctx context.Context, token string) (mig.RefreshToken, error)

	SavePrivateMessage(ctx context.Context, arg SavePrivateMessageParams) (mig.Message, error)

	UpdateMessage(ctx context.Context, arg UpdateMessageParams) (mig.Message, error)

	DeleteMessage(ctx context.Context, id string) error

	GetLastReadMessage(ctx context.Context, senderID, recipientID string) (mig.Message, error)
}

type UserRepositoryPostgreSQL struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewUserRepositoryPostgreSQL(pool *pgxpool.Pool, queries *db.Queries) (*UserRepositoryPostgreSQL, error) {
	if pool == nil {
		return nil, fmt.Errorf("missing pool")
	}

	if queries == nil {
		return nil, fmt.Errorf("missing queries")
	}

	repo := &UserRepositoryPostgreSQL{
		pool:    pool,
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

func getPrivateMessageFromDBModel(msg db.GetPrivateConversationRow) mig.Message {
	senderID := UUIDToString(msg.SenderID)
	recipientID := null.NewString(UUIDToString(msg.RecipientID), msg.RecipientID.Valid)

	return mig.Message{
		ID:            UUIDToString(msg.ID),
		Content:       msg.Content,
		WorkflowState: mig.MessageWorkflowState(msg.WorkflowState),
		Type:          mig.MessageType(msg.MessageType),
		CreatedAt:     msg.CreatedAt.Time,
		SenderID:      senderID,
		RecipientID:   recipientID,
		IsRead:        null.NewBool(msg.IsRead.Bool, msg.IsRead.Valid),
		Sender: mig.User{
			ID:            senderID,
			Email:         msg.SenderEmail,
			Username:      msg.SenderUsername,
			WorkflowState: mig.UserWorkflowState(msg.SenderWorkflowState),
		},
		Recipient: mig.User{
			ID:            recipientID.String,
			Email:         msg.RecipientEmail,
			Username:      msg.RecipientUsername,
			WorkflowState: mig.UserWorkflowState(msg.RecipientWorkflowState),
		},
	}
}

func getPrivateMessagesFromDBModel(msgs []db.GetPrivateConversationRow) []mig.Message {
	result := make([]mig.Message, len(msgs))

	for i, msg := range msgs {
		result[i] = getPrivateMessageFromDBModel(msg)
	}

	return result
}

func (r *UserRepositoryPostgreSQL) GetPrivateConversation(ctx context.Context, firstUserID, secondUserID string, pagination mig.Pagination) ([]mig.Message, error) {
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

func (r *UserRepositoryPostgreSQL) GetRecentPrivateMessages(ctx context.Context, userID string, pagination mig.Pagination) ([]mig.Message, error) {
	userUUID, err := StringToUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid %s", userID)
	}

	arg := db.GetRecentUniquePrivateMessagesParams{
		UserID:   userUUID,
		Page:     int32(pagination.Page),
		PageSize: int32(pagination.PageSize),
	}

	msgs, err := r.queries.GetRecentUniquePrivateMessages(ctx, arg)
	if err != nil {
		return nil, err
	}

	result := make([]mig.Message, len(msgs))

	for i, msg := range msgs {
		senderID := UUIDToString(msg.SenderID)
		recipientID := null.NewString(UUIDToString(msg.RecipientID), msg.RecipientID.Valid)

		result[i] = mig.Message{
			ID:            UUIDToString(msg.ID),
			Content:       msg.Content,
			WorkflowState: mig.MessageWorkflowState(msg.WorkflowState),
			Type:          mig.MessageType(msg.MessageType),
			CreatedAt:     msg.CreatedAt.Time,
			SenderID:      senderID,
			RecipientID:   recipientID,
			IsRead:        null.NewBool(msg.IsRead.Bool, msg.ID.Valid),
			Sender: mig.User{
				ID:            senderID,
				Email:         msg.SenderEmail,
				Username:      msg.SenderUsername,
				WorkflowState: mig.UserWorkflowState(msg.SenderWorkflowState),
			},
			Recipient: mig.User{
				ID:            recipientID.String,
				Email:         msg.RecipientEmail,
				Username:      msg.RecipientUsername,
				WorkflowState: mig.UserWorkflowState(msg.RecipientWorkflowState),
			},
		}
	}

	return result, nil
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
	tx, err := r.pool.Begin(ctx)
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

func (r *UserRepositoryPostgreSQL) GetRefreshToken(ctx context.Context, token string) (mig.RefreshToken, error) {
	refresh, err := r.queries.GetRefreshToken(ctx, token)
	if err != nil {
		return mig.RefreshToken{}, err
	}

	return mig.RefreshToken{
		ID:        UUIDToString(refresh.ID),
		UserID:    UUIDToString(refresh.UserID),
		Token:     refresh.Token,
		ExpiresAt: refresh.ExpiresAt.Time,
		Revoked:   refresh.Revoked.Bool,
	}, nil
}

type SavePrivateMessageParams struct {
	ID          string
	SenderID    string
	RecipientID string
	Content     string
}

func (r *UserRepositoryPostgreSQL) SavePrivateMessage(ctx context.Context, arg SavePrivateMessageParams) (mig.Message, error) {
	id, err := StringToUUID(arg.ID)
	if err != nil {
		return mig.Message{}, err

	}

	senderID, err := StringToUUID(arg.SenderID)
	if err != nil {
		return mig.Message{}, err

	}

	recipientID, err := StringToUUID(arg.RecipientID)
	if err != nil {
		return mig.Message{}, err

	}

	msg, err := r.queries.CreatePrivateMessage(ctx, db.CreatePrivateMessageParams{
		ID:          id,
		SenderID:    senderID,
		RecipientID: recipientID,
		Content:     arg.Content,
		IsRead:      pgtype.Bool{Bool: false, Valid: true},
	})

	if err != nil {
		return mig.Message{}, err
	}

	return mig.Message{
		ID:            UUIDToString(msg.ID),
		Content:       msg.Content,
		WorkflowState: mig.MessageWorkflowState(msg.WorkflowState),
		Type:          mig.MessageType(msg.MessageType),
		SenderID:      UUIDToString(msg.SenderID),
		RecipientID:   null.NewString(UUIDToString(msg.RecipientID), msg.RecipientID.Valid),
		ChatroomID:    null.NewString(UUIDToString(msg.ChatroomID), msg.ChatroomID.Valid),
		IsRead:        null.NewBool(msg.IsRead.Bool, msg.IsRead.Valid),
		CreatedAt:     msg.CreatedAt.Time,
	}, nil
}

type UpdateMessageParams struct {
	ID           string
	Content      string
	WorflowState mig.MessageWorkflowState
	IsRead       null.Bool
}

func (r *UserRepositoryPostgreSQL) UpdateMessage(ctx context.Context, arg UpdateMessageParams) (mig.Message, error) {
	uuid, err := StringToUUID(arg.ID)
	if err != nil {
		return mig.Message{}, err

	}

	msg, err := r.queries.UpdateMessage(ctx, db.UpdateMessageParams{
		Content:       arg.Content,
		WorkflowState: db.MessageWorkflowState(arg.WorflowState),
		IsRead:        pgtype.Bool{Bool: arg.IsRead.Bool, Valid: arg.IsRead.Valid},
		ID:            uuid,
	})

	if err != nil {
		return mig.Message{}, err
	}

	return mig.Message{
		ID:            UUIDToString(msg.ID),
		Content:       msg.Content,
		WorkflowState: mig.MessageWorkflowState(msg.WorkflowState),
		Type:          mig.MessageType(msg.MessageType),
		SenderID:      UUIDToString(msg.SenderID),
		RecipientID:   null.NewString(UUIDToString(msg.RecipientID), msg.RecipientID.Valid),
		ChatroomID:    null.NewString(UUIDToString(msg.ChatroomID), msg.ChatroomID.Valid),
		IsRead:        null.NewBool(msg.IsRead.Bool, msg.IsRead.Valid),
		CreatedAt:     msg.CreatedAt.Time,
	}, nil
}

func (r *UserRepositoryPostgreSQL) DeleteMessage(ctx context.Context, id string) error {
	uuid, err := StringToUUID(id)
	if err != nil {
		return err

	}

	_, err = r.queries.DeleteMessage(ctx, db.DeleteMessageParams{
		DeletedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		ID:        uuid,
	})

	return err
}

func (r *UserRepositoryPostgreSQL) GetLastReadMessage(ctx context.Context, senderID, recipientID string) (mig.Message, error) {
	senderUuid, err := StringToUUID(senderID)
	if err != nil {
		return mig.Message{}, err
	}

	recipientUuid, err := StringToUUID(recipientID)
	if err != nil {
		return mig.Message{}, err
	}

	message, err := r.queries.GetLastReadMessage(ctx, db.GetLastReadMessageParams{
		SenderID:    senderUuid,
		RecipientID: recipientUuid,
	})
	if err != nil {
		return mig.Message{}, err
	}

	return mig.Message{
		ID:            message.ID.String(),
		Content:       message.Content,
		WorkflowState: mig.MessageWorkflowState(message.WorkflowState),
		SenderID:      message.SenderID.String(),
		RecipientID:   null.NewString(message.RecipientID.String(), message.RecipientID.Valid),
		CreatedAt:     message.CreatedAt.Time,
	}, nil
}
