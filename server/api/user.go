package api

import (
	"database/sql"
	"encoding/json"
	"errors"
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

// GetPrivateConversation handler returns messages exchanged between two users.
// It expects user ID and recipient ID to be passed as URL parameters.
// Result is paginated based on the provided `pagination` query parameters.
// Pagination is optional: if not provided, default pagination settings will be used.
func (c *HttpApiController) GetPrivateConversation(w http.ResponseWriter, r *http.Request) {
	firstUserID := chi.URLParam(r, "user_id")

	secondUserID := chi.URLParam(r, "recipient_id")

	pagination := mig.NewPagination(r)

	result, err := c.userService.GetPrivateConversation(r.Context(), firstUserID, secondUserID, pagination)
	if err != nil {
		mig.HttpErrorReply(w, err)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		mig.HttpErrorReply(w, mig.NewError(mig.ErrMsgUnableToJsonEnode, err, mig.InternalServerError))
	}
}

func (c *HttpApiController) GetRecentPrivateMessages(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	pagination := mig.NewPagination(r)

	messages, err := c.userService.GetRecentPrivateMessagesWithUniqueParticipant(r.Context(), userID, pagination)
	if err != nil {
		mig.HttpErrorReply(w, err)
		return
	}

	result := make([]struct {
		Message        mig.Message   `json:"message"`
		UnreadMessages []mig.Message `json:"unread_messages"`
	}, len(messages))

	for i, msg := range messages {
		senderID := msg.SenderID
		if msg.SenderID == userID {
			senderID = msg.RecipientID.String
		}

		messages, err := c.userService.GetUnreadMessages(r.Context(), senderID, userID)
		if err != nil {
			mig.HttpErrorReply(w, err)
			return
		}

		result[i] = struct {
			Message        mig.Message   `json:"message"`
			UnreadMessages []mig.Message `json:"unread_messages"`
		}{
			Message:        msg,
			UnreadMessages: messages,
		}
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		mig.HttpErrorReply(w, mig.NewError(mig.ErrMsgUnableToJsonEnode, err, mig.InternalServerError))
	}
}

type UpdateReadMessagesRequest struct {
	SenderID string `json:"sender_id"`
}

func (c *HttpApiController) UpdateReadMessages(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")

	var req UpdateReadMessagesRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid input.", http.StatusBadRequest)
		return
	}

	if req.SenderID == "" {
		http.Error(w, "Missing sender id.", http.StatusBadRequest)
		return
	}

	msg, err := c.userService.GetLastReadMessage(r.Context(), req.SenderID, userID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		query := `
			UPDATE messages 
			SET is_read = TRUE
			WHERE sender_id = $1 AND recipient_id = $2;
		`
		_, err := c.db.Exec(r.Context(), query, req.SenderID, userID)
		if err != nil {
			http.Error(w, "Unable to handle request.", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	if err != nil {
		http.Error(w, "Unable to handle request.", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE messages 
		SET is_read = TRUE
		WHERE sender_id = $1 AND 
			recipient_id = $2 AND
			created_at > $3;
		`
	_, err = c.db.Exec(r.Context(), query, req.SenderID, userID, msg.CreatedAt)
	if err != nil {
		http.Error(w, "Unable to handle request.", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *HttpApiController) GetUnreadMessages(w http.ResponseWriter, r *http.Request) {
	recipientID := chi.URLParam(r, "user_id")
	senderID := chi.URLParam(r, "sender_id")

	result, err := c.userService.GetUnreadMessages(r.Context(), senderID, recipientID)
	if err != nil {
		mig.HttpErrorReply(w, err)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		mig.HttpErrorReply(w, mig.NewError(mig.ErrMsgUnableToJsonEnode, err, mig.InternalServerError))
	}
}
