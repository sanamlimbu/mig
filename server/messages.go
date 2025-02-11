package mig

import (
	"time"

	"github.com/guregu/null/v5"
)

type MessageType string

const (
	MessageTypePrivate  MessageType = "private"
	MessageTypeChatroom MessageType = "chatroom"
)

type MessageWorkflowState string

const (
	MessageWorkflowStateCreated MessageWorkflowState = "created"
	MessageWorkflowStateUpdated MessageWorkflowState = "updated"
	MessageWorkflowStateDeleted MessageWorkflowState = "deleted"
	MessageWorkflowStateRead    MessageWorkflowState = "read"
)

type Message struct {
	ID            string               `json:"id"`
	Content       string               `json:"content"`
	WorkflowState MessageWorkflowState `json:"workflow_state"`
	Type          MessageType          `json:"type"`
	SenderID      string               `json:"sender_id"`
	RecipientID   null.String          `json:"recipient_id"`
	ChatroomID    null.String          `json:"chatroom_id"`
	IsRead        null.Bool            `json:"is_read"`
	Sender        *User                `json:"sender,omitempty"`
	Chatroom      *Chatroom            `json:"chatroom,omitempty"`
	Recipient     *User                `json:"recipient,omitempty"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
	DeletedAt     null.Time            `json:"deleted_at"`
}
