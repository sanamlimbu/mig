package mig

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/go-chi/chi"
)

// query params
//   - search_term : string (required)
//   - state[] : array of states - 'active','deleted' (required)
func (c *HttpApiController) getChatrooms(w http.ResponseWriter, r *http.Request, pagination Pagination) (int, error) {
	searchTerm := chi.URLParam(r, "search_term")
	if searchTerm == "" {
		return http.StatusBadRequest, fmt.Errorf("missing search_term params")
	}

	states := r.URL.Query()["state[]"]

	if len(states) == 0 {
		return http.StatusBadRequest, fmt.Errorf("missing state[] parmas")
	}

	for _, state := range states {
		if !slices.Contains(AllChatroomsWorkflowState(), ChatroomsWorkflowState(state)) {
			return http.StatusBadRequest, fmt.Errorf("invalid state params: %s", state)
		}
	}

	chatrooms, err := c.chatroomsService.getChatroomsBySearchTermAndWorkflowState(searchTerm, states, pagination)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if err := json.NewEncoder(w).Encode(chatrooms); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}
