package mig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
)

type Pagination struct {
	page     int
	pageSize int
}

func withPagination(next func(w http.ResponseWriter, r *http.Request, pagination Pagination) (int, error)) func(w http.ResponseWriter, r *http.Request) (int, error) {
	fn := func(w http.ResponseWriter, r *http.Request) (int, error) {
		page, err := strconv.Atoi(r.URL.Query().Get("page"))

		if err != nil {
			return http.StatusBadRequest, fmt.Errorf("invalid or missing page")
		}

		pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))

		if err != nil {
			return http.StatusBadRequest, fmt.Errorf("invalid or missing page_size")
		}

		pagination := Pagination{
			page:     page,
			pageSize: pageSize,
		}

		return next(w, r, pagination)
	}

	return fn
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func withError(next func(w http.ResponseWriter, r *http.Request) (int, error)) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		contents, err := io.ReadAll(r.Body)
		if err != nil {
			log.Err(err)
		}
		r.Body = io.NopCloser(bytes.NewReader(contents))
		defer r.Body.Close()

		code, err := next(w, r)
		if err != nil {
			errResponse := ErrorResponse{
				Code:    fmt.Sprintf("%d", code),
				Message: err.Error(),
			}

			if code == http.StatusInternalServerError {
				log.Error().Msg(err.Error())
				errResponse.Message = http.StatusText(http.StatusInternalServerError)
			}

			jsonErr, err := json.Marshal(code)
			if err != nil {
				log.Err(err)
				http.Error(w, `{"code":"00001","message":"JSON failed, please contact IT."}`, code)
				return
			}

			http.Error(w, string(jsonErr), code)
			return
		}
	}

	return fn
}

func withAuth(c *APIController, next func(u User, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := User{
			ID:       1,
			Username: "sudosanam",
		}
		next(user, w, r)

	}

	return fn
}
