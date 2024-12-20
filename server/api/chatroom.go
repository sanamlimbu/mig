package api

import (
	"encoding/json"
	"fmt"
	"mig"
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"
)

// GetValidChatroomWorkflowStates returns strings slice of valid chatroom workflow states.
func GetValidChatroomWorkflowStates(input []string) []string {
	result := make([]string, 0, len(input))
	allWorkflowStates := mig.AllChatroomWorkflowState()

	for _, str := range input {
		if slices.Contains(allWorkflowStates, mig.ChatroomWorkflowState(str)) {
			result = append(result, str)
		}
	}

	return result
}

// GetChatrooms returns chatrooms.
// It accepts an optional query parameter `state[]`, which allows filtering by chatroom workflow states.
// If no valid states are provided, default state "active" is used.
// It also accepts optional query parameter `search_term`.
// If `search_term` is missing no filtering is applied.
// Result is paginated based on provided `pagination` query parameters.
// Pagination is optional: if not provided, default pagination settings will be used.
func (c *HttpApiController) GetChatrooms(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("search_term")

	chatroomWorkflowStates := GetValidChatroomWorkflowStates(r.URL.Query()["state[]"])

	if len(chatroomWorkflowStates) == 0 {
		chatroomWorkflowStates = append(chatroomWorkflowStates, string(mig.ChatroomWorkflowStateActive))
	}

	pagination := mig.NewPagination(r)

	wildCardSearchTerm := fmt.Sprintf("%%%s%%", searchTerm)

	result, err := c.chatroomService.GetChatroomsBySearchTermAndWorkflowStates(r.Context(), wildCardSearchTerm, chatroomWorkflowStates, pagination)
	if err != nil {
		mig.HttpErrorReply(w, err)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		mig.HttpErrorReply(w, mig.NewError(mig.ErrMsgUnableToJsonEnode, err, mig.InternalServerError))
	}
}

// GetChatroom returns chatroom for given chatroom ID.
func (c *HttpApiController) GetChatroom(w http.ResponseWriter, r *http.Request) {
	chatroomID := chi.URLParam(r, "chatroom_id")

	result, err := c.chatroomService.GetChatroomWithCreator(r.Context(), chatroomID)
	if err != nil {
		mig.HttpErrorReply(w, err)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		mig.HttpErrorReply(w, mig.NewError(mig.ErrMsgUnableToJsonEnode, err, mig.InternalServerError))
	}
}

// GetChatroomMessages returns messages sent in given chatroom.
// Returned result is paginated.
func (c *HttpApiController) GetChatroomMessages(w http.ResponseWriter, r *http.Request) {
	chatroomID := chi.URLParam(r, "chatroom_id")

	pagination := mig.NewPagination(r)

	result, err := c.chatroomService.GetMessages(r.Context(), chatroomID, pagination)
	if err != nil {
		mig.HttpErrorReply(w, err)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		mig.HttpErrorReply(w, mig.NewError(mig.ErrMsgUnableToJsonEnode, err, mig.InternalServerError))
	}
}
