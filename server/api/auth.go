package api

import (
	"encoding/json"
	"mig"
	"net/http"
)

const fingerprintCookie string = "Fingerprint"

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

	cookie := http.Cookie{
		Name:     fingerprintCookie,
		Value:    resp.UserFingerprint,
		Path:     "/",
		MaxAge:   (24 * 60 * 60) - 10, // Refresh token expires in 24 hours.
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}

	http.SetCookie(w, &cookie)

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

type refreshTokenRequest struct {
	RefreshToken    string `json:"refresh_token"`
	UserFingerprint string `json:"user_fingerprint"`
}

func (c *HttpApiController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(fingerprintCookie)
	if err != nil {
		http.Error(w, "Missing user fingerprint cookie.", http.StatusBadRequest)
		return
	}

	var req refreshTokenRequest

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid refresh token request.", http.StatusBadRequest)
		return
	}

	if req.UserFingerprint == "" {
		http.Error(w, "Missing user fingerprint.", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, "Missing refresh token.", http.StatusBadRequest)
		return
	}

	resp, err := c.authService.RefreshToken(r.Context(), req.RefreshToken, req.UserFingerprint, cookie.Value)
	if err != nil {
		http.Error(w, "Unable to generate access token.", http.StatusUnauthorized)
		return
	}

	cookie = &http.Cookie{
		Name:     fingerprintCookie,
		Value:    resp.UserFingerprint,
		Path:     "/",
		MaxAge:   (24 * 60 * 60) - 10, // Refresh token expires in 24 hours.
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}

	http.SetCookie(w, cookie)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		mig.HttpErrorReply(w, mig.NewError(mig.ErrMsgSomethingWentWrong, err, mig.InternalServerError))
	}
}
