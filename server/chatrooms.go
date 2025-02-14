package mig

import (
	"time"

	"github.com/guregu/null/v5"
)

type Chatroom struct {
	ID            string                `json:"id"`
	Name          string                `json:"name"`
	WorkflowState ChatroomWorkflowState `json:"workflow_state"`
	Type          ChatroomType          `json:"type"`
	CreatedBy     string                `json:"created_by"`
	CreatedAt     time.Time             `json:"created_at,omitzero"`
	UpdatedAt     time.Time             `json:"updated_at,omitzero"`
	DeletedAt     null.Time             `json:"deleted_at,omitzero"`
	Creator       User                  `json:"creator,omitzero"`
}

type ChatroomWorkflowState string

const (
	ChatroomWorkflowStateActive  ChatroomWorkflowState = "active"
	ChatroomWorkflowStateDeleted ChatroomWorkflowState = "deleted"
)

func AllChatroomWorkflowState() []ChatroomWorkflowState {
	return []ChatroomWorkflowState{
		ChatroomWorkflowStateActive,
		ChatroomWorkflowStateDeleted,
	}
}

type ChatroomType string

const (
	ChatroomTypePrivate ChatroomType = "private"
	ChatroomTypePublic  ChatroomType = "public"
)

func AllChatroomType() []ChatroomType {
	return []ChatroomType{
		ChatroomTypePrivate,
		ChatroomTypePublic,
	}
}
