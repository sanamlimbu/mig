package mig

import "time"

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

type ChatroomMessage struct {
	ID                    string                `json:"id"`
	Content               string                `json:"content"`
	WorkflowState         MessageWorkflowState  `json:"workflow_state"`
	Type                  MessageType           `json:"type"`
	CreatedAt             time.Time             `json:"created_at"`
	SenderID              string                `json:"sender_id"`
	ChatroomID            string                `json:"chatroom_id"`
	ChatroomCreatorID     string                `json:"chatroom_creator_id"`
	SenderEmail           string                `json:"sender_email"`
	SenderUsername        string                `json:"sender_username"`
	SenderWorkflowState   UserWorkflowState     `json:"sender_workflow_state"`
	ChatroomName          string                `json:"chatroom_name"`
	ChatroomWorkflowState ChatroomWorkflowState `json:"chatroom_workflow_state"`
}

type PrivateMessage struct {
	ID                     string               `json:"id"`
	Content                string               `json:"content"`
	WorkflowState          MessageWorkflowState `json:"workflow_state"`
	Type                   MessageType          `json:"type"`
	CreatedAt              time.Time            `json:"created_at"`
	SenderID               string               `json:"sender_id"`
	SenderUsername         string               `json:"sender_username"`
	SenderEmail            string               `json:"sender_email"`
	SenderWorkflowState    UserWorkflowState    `json:"sender_workflow_state"`
	RecipientID            string               `json:"recipient_id"`
	RecipientUsername      string               `json:"recipient_username"`
	RecipientEmail         string               `json:"recipient_email"`
	RecipientWorkflowState UserWorkflowState    `json:"recipient_workflow_state"`
}
