package mig

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/guregu/null/v5"
)

type UserRole string

const (
	UserRoleSuperAdmin UserRole = "superadmin"
	UserRoleAdmin      UserRole = "admin"
	UserRoleMember     UserRole = "member"
)

type User struct {
	ID            string            `json:"id"`
	Email         string            `json:"email"`
	Username      string            `json:"username"`
	WorkflowState UserWorkflowState `json:"workflow_state"`
	Role          UserRole          `json:"role"`
	CreatedAt     time.Time         `json:"created_at,omitzero"`
	UpdatedAt     time.Time         `json:"updated_at,omitzero"`
	DeletedAt     null.Time         `json:"deleted_at,omitzero"`
}

type UserWorkflowState string

const (
	UserWorkflowStateActive     UserWorkflowState = "active"
	UserWorkflowStateSuspended  UserWorkflowState = "suspended"
	UserWorkflowStateDeleted    UserWorkflowState = "deleted"
	UserWorkflowStateUnverified UserWorkflowState = "unverified"
)

func AllUserWorkflowState() []UserWorkflowState {
	return []UserWorkflowState{
		UserWorkflowStateActive,
		UserWorkflowStateSuspended,
		UserWorkflowStateUnverified,
		UserWorkflowStateDeleted,
	}
}

type FriendshipWorkflowState string

const (
	FriendshipWorkflowStatePending   FriendshipWorkflowState = "pending"
	FriendshipWorkflowStateActive    FriendshipWorkflowState = "active"
	FriendshipWorkflowStateRejected  FriendshipWorkflowState = "rejected"
	FriendshipWorkflowStateCancelled FriendshipWorkflowState = "cancelled"
	FriendshipWorkflowStateDeleted   FriendshipWorkflowState = "deleted"
)

func AllFriendshipWorkflowState() []FriendshipWorkflowState {
	return []FriendshipWorkflowState{
		FriendshipWorkflowStatePending,
		FriendshipWorkflowStateActive,
		FriendshipWorkflowStateRejected,
		FriendshipWorkflowStateCancelled,
		FriendshipWorkflowStateDeleted,
	}
}

type RefreshToken struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	Revoked   bool
}

type NullUser struct {
	User  User
	Valid bool
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this NullUser is null.
func (u *NullUser) MarshalJSON() ([]byte, error) {
	if !u.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(u.User)
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports User and null input.
func (u *NullUser) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == 'n' {
		u.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &u.User); err != nil {
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	u.Valid = true
	return nil
}

type NullChatroom struct {
	Chatroom Chatroom
	Valid    bool
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this NullChatroom is null.
func (c *NullChatroom) MarshalJSON() ([]byte, error) {
	if !c.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(c.Chatroom)
}

// UnmarshalJSON implements json.Unmarshaler.
// It supports Chatroom and null input.
func (c *NullChatroom) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == 'n' {
		c.Valid = false
		return nil
	}

	if err := json.Unmarshal(data, &c.Chatroom); err != nil {
		return fmt.Errorf("null: couldn't unmarshal JSON: %w", err)
	}

	c.Valid = true
	return nil
}
