package mig

import (
	"time"
)

type UserRole string

const (
	SuperAdminUserRole UserRole = "superadmin"
	AdminUserRole      UserRole = "admin"
	MemberUserRole     UserRole = "member"
)

type User struct {
	ID            string            `json:"id"`
	Email         string            `json:"email"`
	Username      string            `json:"username"`
	WorkflowState UserWorkflowState `json:"workflow_state"`
	Role          UserRole          `json:"role"`
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
