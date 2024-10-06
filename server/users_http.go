package mig

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

// query params
//   - state[] : array of states - 'pending','active', 'rejected', 'cancelled' (required)
func (c *APIController) getFriends(w http.ResponseWriter, r *http.Request, pagination Pagination) (int, error) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("invalid user id: %d", userID)
	}

	states := r.URL.Query()["state[]"]

	if len(states) == 0 {
		return http.StatusBadRequest, fmt.Errorf("missing state[] params")
	}

	friends, err := c.usersService.GetFriends(userID, states, pagination)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if err := json.NewEncoder(w).Encode(friends); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}
