package mig

import "context"

type Chatroom struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	WorkflowState string `json:"workflow_state"`
	Type          string `json:"type"`
	CreatedBy     int64  `json:"created_by"`
}

type ChatroomsWorkflowState string

const (
	ChatroomsWorkflowStateActive  ChatroomsWorkflowState = "active"
	ChatroomsWorkflowStateDeleted ChatroomsWorkflowState = "deleted"
)

func AllChatroomsWorkflowState() []ChatroomsWorkflowState {
	return []ChatroomsWorkflowState{
		ChatroomsWorkflowStateActive,
		ChatroomsWorkflowStateDeleted,
	}
}

type ChatroomsService struct {
	chatroomsRepo ChatroomsRepository
}

func NewChatroomsService(chatroomsRepo ChatroomsRepository) *ChatroomsService {
	return &ChatroomsService{
		chatroomsRepo: chatroomsRepo,
	}
}

func (s *ChatroomsService) getChatroomsBySearchTermAndWorkflowState(searchTerm string, states []string, pagination Pagination) ([]Chatroom, error) {
	return s.chatroomsRepo.getChatroomsBySearchTermAndWorkflowState(context.Background(), searchTerm, states, pagination)
}
