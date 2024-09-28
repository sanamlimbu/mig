package mig

import (
	"context"
	"encoding/json"
	"fmt"
	"mig/models"
	"net/http"
	"slices"
	"strconv"

	"github.com/go-chi/chi"
)

func userDTO(u *models.User) User {
	return User{
		ID:            u.ID,
		Username:      u.Username,
		WorkflowState: u.WorkflowState.String(),
	}
}

func usersDTO(users []*models.User) []User {
	results := []User{}

	for _, u := range users {
		results = append(results, User{
			ID:            u.ID,
			Username:      u.Username,
			WorkflowState: u.WorkflowState.String(),
		})
	}

	return results
}

// query params: state (optional)
func (c *APIController) getFriends(w http.ResponseWriter, r *http.Request, pagination Pagination) (int, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("invalid user id")
	}

	state := r.URL.Query().Get("state") // 'pending','active','rejected','cancelled'

	if state == "" {
		state = "active"
	} else if slices.Contains(models.AllFriendshipsWorkflowState(), models.FriendshipsWorkflowState(state)) {
	} else {
		return http.StatusBadRequest, fmt.Errorf("invalid state")
	}

	data, err := getFriendsByWorkflowState(context.Background(), c.db, pagination, id, models.FriendshipsWorkflowState(state))

	results := usersDTO(data)

	if err := json.NewEncoder(w).Encode(results); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}
