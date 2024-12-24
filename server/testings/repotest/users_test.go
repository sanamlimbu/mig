package repotest

import (
	"context"
	"testing"
)

func TestGetUserByEmail(t *testing.T) {
	ctx := context.Background()

	email := "jack@example.com"

	user, err := userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		t.Errorf("unable to find user: %s", err)
	}

	if user.Email != email {
		t.Errorf("expected user with email: %s but got: %s", email, user.Email)
	}
}
