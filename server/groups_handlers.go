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

type GroupDTO struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	WorkflowState string `json:"workflow_state"`
	Type          string `json:"type"`
	CreatedBy     int64  `json:"created_by"`
}

func groupDTO(g Group) GroupDTO {
	return GroupDTO{
		ID:            g.ID,
		Name:          g.Name,
		WorkflowState: g.WorkflowState,
		Type:          g.Type,
		CreatedBy:     g.CreatedBy,
	}
}

func groupsDTO(groups []Group) []GroupDTO {
	results := []GroupDTO{}

	for _, g := range groups {
		results = append(results, GroupDTO{
			ID:            g.ID,
			Name:          g.Name,
			WorkflowState: g.WorkflowState,
			Type:          g.Type,
			CreatedBy:     g.CreatedBy,
		})
	}

	return results
}

// query params
//   - user_id : int64 (required)
//   - state[] : array of states - 'pending','active','rejected','cancelled' (required)
func (c *APIController) getGroups(w http.ResponseWriter, r *http.Request, pagination Pagination) (int, error) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "user_id"), 10, 64)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("invalid or missing user_id params: %s")
	}

	states := r.URL.Query()["state[]"]

	if len(states) == 0 {
		return http.StatusBadRequest, fmt.Errorf("missing state[] parmas")
	}

	for _, state := range states {
		if !slices.Contains(models.AllGroupUsersWorkflowState(), models.GroupUsersWorkflowState(state)) {
			return http.StatusBadRequest, fmt.Errorf("invalid state params: %s", state)
		}
	}

	data, err := c.groupsRepo.getGroupsByWorflowStatesAndUserID(context.Background(), pagination, states, userID)
	if err != nil {
		return http.StatusBadRequest, err
	}

	results := groupsDTO(data)

	if err := json.NewEncoder(w).Encode(results); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}
