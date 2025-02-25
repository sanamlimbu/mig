package repotest

import (
	"context"
	"database/sql"
	"errors"
	"mig"
	"mig/seed"
	"testing"
)

func TestGetUser(t *testing.T) {
	ctx := context.Background()
	id := seed.UsersUUIDs[0]

	user, err := userRepo.GetUser(ctx, id)
	if err != nil {
		t.Errorf("unable to find user: %s", err)
	}

	if user.ID != id {
		t.Errorf("expected user id: %s but got: %s", id, user.ID)
	}

	user, err = userRepo.GetUser(ctx, "cb53935a-27fd-49f6-947b-bff0370243ef")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows but got: %s", err)
	}

	user, err = userRepo.GetUser(ctx, "09ffa0c8-d142-4281-8f76-59dde626a85fffff")
	if err == nil {
		t.Errorf("expected err: invalid uuid but got: nil")
	}
}

func TestGetUserByEmail(t *testing.T) {
	ctx := context.Background()
	email := "jack@limbu.dev"

	user, err := userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		t.Errorf("unable to find user: %s", err)
	}

	if user.Email != email {
		t.Errorf("expected email: %s but got: %s", email, user.Email)
	}

	user, err = userRepo.GetUserByEmail(ctx, "")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows but got: %s", err)
	}

	user, err = userRepo.GetUserByEmail(ctx, "elon@tesla.com")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows but got: %s", err)
	}
}

func TestGetUserByUsername(t *testing.T) {
	ctx := context.Background()
	username := "jack"

	user, err := userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		t.Errorf("unable to find user: %s", err)
	}

	if user.Username != username {
		t.Errorf("expected username: %s but got: %s", username, user.Username)
	}
}

func TestGetChatroomsByCreatorID(t *testing.T) {
	ctx := context.Background()
	creatorID := seed.UsersUUIDs[0]

	chatrooms, err := userRepo.GetChatroomsByCreatorID(ctx, creatorID, []string{"active", "deleted"}, mig.Pagination{Page: 0, PageSize: 10})
	if err != nil {
		t.Errorf("unable to fetch chatrooms: %s", err)
	}

	if len(chatrooms) == 0 {
		t.Errorf("expected: 1 but got: %d", len(chatrooms))
	}

	chatrooms, err = userRepo.GetChatroomsByCreatorID(ctx, creatorID, []string{"active", "deleted"}, mig.Pagination{Page: 1, PageSize: 10})
	if err != nil {
		t.Errorf("unable to fetch chatrooms: %s", err)
	}

	if len(chatrooms) != 0 {
		t.Errorf("expected: 0 but got: %d", len(chatrooms))
	}

	chatrooms, err = userRepo.GetChatroomsByCreatorID(ctx, creatorID, []string{"deleted"}, mig.Pagination{Page: 0, PageSize: 10})
	if err != nil {
		t.Errorf("unable to fetch chatrooms: %s", err)
	}

	if len(chatrooms) != 0 {
		t.Errorf("expected: 0 but got: %d", len(chatrooms))
	}
}

func TestGetFriendsByFriendshipWorkflowStates(t *testing.T) {
	ctx := context.Background()
	userID := seed.UsersUUIDs[0]

	friends, err := userRepo.GetFriendsByFriendshipWorkflowStates(ctx, userID, []string{"active", "pending"}, mig.Pagination{Page: 0, PageSize: 22})
	if err != nil {
		t.Errorf("unable to friends: %s", err)
	}

	if len(friends) != 21 {
		t.Errorf("expected: 21 but got: %d", len(friends))
	}
}

func TestGetFriend(t *testing.T) {
	ctx := context.Background()
	userID := seed.UsersUUIDs[0]
	friendID := seed.UsersUUIDs[1]

	friend, err := userRepo.GetFriend(ctx, userID, friendID)
	if err != nil {
		t.Errorf("unable to fetch friend: %s", err)
	}

	if friend.ID != friendID {
		t.Errorf("expected friend ID: %s but got: %s", friendID, friend.ID)
	}
}

func TestGetPrivateConversation(t *testing.T) {
	ctx := context.Background()
	userID1 := seed.UsersUUIDs[0]
	userID2 := seed.UsersUUIDs[1]
	pagination := mig.Pagination{Page: 0, PageSize: 20}

	messages, err := userRepo.GetPrivateConversation(ctx, userID1, userID2, pagination)
	if err != nil {
		t.Errorf("unable to fetch private conversation: %s", err)
	}

	if len(messages) == 0 {
		t.Errorf("expected some messages but got zero")
	}
}
