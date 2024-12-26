package api

import (
	"encoding/json"
	"mig"
	"net/http"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (c *HttpApiController) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid login input.", http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		http.Error(w, "Missing username.", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, "Missing password.", http.StatusBadRequest)
		return
	}

	resp, err := c.authService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		mig.HttpErrorReply(w, err)
		return
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		mig.HttpErrorReply(w, mig.NewError(mig.ErrMsgSomethingWentWrong, err, mig.InternalServerError))
	}
}

type signupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (c *HttpApiController) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid signup input.", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Missing email.", http.StatusBadRequest)
		return
	}

	if req.Username == "" {
		http.Error(w, "Missing username.", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		http.Error(w, "Missing password.", http.StatusBadRequest)
		return
	}

	resp, err := c.authService.Signup(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		mig.HttpErrorReply(w, err)
		return
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		mig.HttpErrorReply(w, mig.NewError(mig.ErrMsgSomethingWentWrong, err, mig.InternalServerError))
	}
}
