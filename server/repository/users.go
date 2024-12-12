package repository

import (
	"context"
	"encoding/hex"
	"fmt"
	"mig"
	"mig/db"

	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository interface {
	// GetChatroomsByCreatorID returns chatrooms created by specified user.
	// Returned chatrooms are filtered based on given chatroom workflow states and paginated.
	GetChatroomsByCreatorID(ctx context.Context, creatorID string, chatroomWorkflowStates []string, pagination mig.Pagination) ([]mig.Chatroom, error)

	// GetFriends returns active friends of specified user.
	// Returned result is paginated.
	GetFriends(ctx context.Context, userID string, pagination mig.Pagination) ([]mig.User, error)

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
}

type UserRepositoryPostgreSQL struct {
	queries *db.Queries
}

func NewUserRepositoryPostgreSQL(queries *db.Queries) (*UserRepositoryPostgreSQL, error) {
	if queries == nil {
		return nil, fmt.Errorf("missing queries")
	}

	repo := &UserRepositoryPostgreSQL{
		queries: queries,
	}

	return repo, nil
}

func uuidToString(uuid pgtype.UUID) string {
	if !uuid.Valid {
		return ""
	}

	return encodeUUID(uuid.Bytes)
}

func stringToUUID(str string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	err := uuid.Scan(str)
	return uuid, err
}

func encodeUUID(src [16]byte) string {
	var buf [36]byte

	hex.Encode(buf[0:8], src[:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], src[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], src[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], src[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], src[10:])

	return string(buf[:])
}

func getUserFromDBModel(user db.User) mig.User {
	return mig.User{
		ID:            uuidToString(user.ID),
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
	creatorUUID, err := stringToUUID(creatorID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid %s", creatorID)
	}

	arg := db.GetChatroomsByCreatorIDParams{
		ID:       creatorUUID,
		Page:     int32(pagination.Page),
		PageSize: int32(pagination.PageSize),
	}

	result, err := r.queries.GetChatroomsByCreatorID(ctx, arg)
	if err != nil {
		return nil, err
	}

	return getChatroomsFromDBModel(result), nil
}

func (r *UserRepositoryPostgreSQL) GetFriends(ctx context.Context, userID string, friendshipWorkflowStates []string, pagination mig.Pagination) ([]mig.User, error) {
	userUUID, err := stringToUUID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid %s", userID)
	}

	arg := db.GetFriendsParams{
		ID:       userUUID,
		Page:     int32(pagination.Page),
		PageSize: int32(pagination.PageSize),
	}

	result, err := r.queries.GetFriends(ctx, arg)
	if err != nil {
		return nil, err
	}

	return getUsersFromDBModel(result), nil
}

func (r *UserRepositoryPostgreSQL) GetFriend(ctx context.Context, userID, friendID string) (mig.User, error) {
	userUUID, err := stringToUUID(userID)
	if err != nil {
		return mig.User{}, fmt.Errorf("invalid uuid %s", userID)
	}

	friendUUID, err := stringToUUID(friendID)
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
	userUUID, err := stringToUUID(userID)
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
		ID:                     uuidToString(msg.ID),
		Content:                msg.Content,
		WorkflowState:          mig.MessageWorkflowState(msg.WorkflowState),
		Type:                   mig.MessageType(msg.MessageType),
		CreatedAt:              msg.CreatedAt.Time,
		SenderID:               uuidToString(msg.SenderID),
		RecipientID:            uuidToString(msg.RecipientID),
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
	firstUserUUID, err := stringToUUID(firstUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid %s", firstUserID)
	}

	secondUserUUID, err := stringToUUID(secondUserID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid %s", secondUserID)
	}

	arg := db.GetPrivateConversationParams{
		FirstUserID:  firstUserUUID,
		SecondUserID: secondUserUUID,
	}

	result, err := r.queries.GetPrivateConversation(ctx, arg)
	if err != nil {
		return nil, err
	}

	return getPrivateMessagesFromDBModel(result), nil
}
