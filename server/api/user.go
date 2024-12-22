package api

import (
	"encoding/json"
	"mig"
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"
)

// GetValidUserWorkflowStates returns strings slice of valid user workflow states.
func GetValidUserWorkflowStates(input []string) []string {
	result := make([]string, 0, len(input))
	allWorkflowStates := mig.AllUserWorkflowState()

	for _, str := range input {
		if slices.Contains(allWorkflowStates, mig.UserWorkflowState(str)) {
			result = append(result, str)
		}
	}

	return result
}

// GetValidFriendshipWorkflowStates returns strings slice of valid friendship workflow states.
func GetValidFriendshipWorkflowStates(input []string) []string {
	result := make([]string, 0, len(input))
	allWorkflowStates := mig.AllFriendshipWorkflowState()

	for _, str := range input {
		if slices.Contains(allWorkflowStates, mig.FriendshipWorkflowState(str)) {
			result = append(result, str)
		}
	}

	return result
}

// GetUser returns user of given id.
func (c *HttpApiController) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")

	result, err := c.userService.GetUser(r.Context(), userID)
	if err != nil {
		mig.HttpErrorReply(w, err)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		mig.HttpErrorReply(w, mig.NewError(mig.ErrMsgUnableToJsonEnode, err, mig.InternalServerError))
	}
}

// GetFriends handler returns friends for specified user, with optional filtering by friendship workflow states.
// It accepts an optional query parameter `state[]`, which allows filtering by following friendship
// workflow states: 'pending', 'active', 'rejected', 'cancelled', and 'deleted'.
// If no valid states are provided, default state "active" is used.
// Result is paginated based on provided `pagination` query parameters.
// Pagination is optional: if not provided, default pagination settings will be used.
func (c *HttpApiController) GetFriends(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")

	friendshipWorkflowStates := GetValidFriendshipWorkflowStates(r.URL.Query()["state[]"])

	if len(friendshipWorkflowStates) == 0 {
		friendshipWorkflowStates = append(friendshipWorkflowStates, string(mig.FriendshipWorkflowStateActive))
	}

	pagination := mig.NewPagination(r)

	result, err := c.userService.GetFriendsByFriendshipWorkflowStates(r.Context(), userID, friendshipWorkflowStates, pagination)
	if err != nil {
		mig.HttpErrorReply(w, err)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		mig.HttpErrorReply(w, mig.NewError(mig.ErrMsgUnableToJsonEnode, err, mig.InternalServerError))
	}
}

// GetPrivateMessages handler returns messages exchanged between two users.
// It expects user ID and recipient ID to be passed as URL parameters.
// Result is paginated based on the provided `pagination` query parameters.
// Pagination is optional: if not provided, default pagination settings will be used.
func (c *HttpApiController) GetPrivateMessages(w http.ResponseWriter, r *http.Request) {
	firstUserID := chi.URLParam(r, "user_id")

	secondUserID := chi.URLParam(r, "recipient_id")

	pagination := mig.NewPagination(r)

	result, err := c.userService.GetPrivateMessages(r.Context(), firstUserID, secondUserID, pagination)
	if err != nil {
		mig.HttpErrorReply(w, err)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		mig.HttpErrorReply(w, mig.NewError(mig.ErrMsgUnableToJsonEnode, err, mig.InternalServerError))
	}
}
