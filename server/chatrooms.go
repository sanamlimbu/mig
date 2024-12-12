package mig

type Chatroom struct {
	ID            string                `json:"id"`
	Name          string                `json:"name"`
	WorkflowState ChatroomWorkflowState `json:"workflow_state"`
	Type          ChatroomType          `json:"type"`
	CreatedBy     string                `json:"created_by"`
}

type ChatroomWithCreator struct {
	ID                   string                `json:"id"`
	Name                 string                `json:"name"`
	WorkflowState        ChatroomWorkflowState `json:"workflow_state"`
	Type                 ChatroomType          `json:"type"`
	CreatedBy            string                `json:"created_by"`
	CreatorUsername      string                `json:"creator_username"`
	CreatorEmail         string                `json:"creator_email"`
	CreatorWorkflowState UserWorkflowState     `json:"creator_workflow_state"`
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
