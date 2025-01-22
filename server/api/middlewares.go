package api

import (
	"context"
	"mig"
	"mig/auth"
	"net/http"
	"strconv"
	"strings"
)

const pageSize int = 10

// paginate is a middleware that ensures the "page" and "page_size" query parameters are valid and greater than or equal to 1.
// If the parameters are invalid or not provided, it falls back to default values: page = 1 and page_size = 10.
// These values are then stored in the request context for use by subsequent handlers.
func paginate(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil || page <= 0 {
			page = 1
		}

		size, err := strconv.Atoi(r.URL.Query().Get("page_size"))
		if err != nil || size <= 0 {
			size = pageSize
		}

		ctx := context.WithValue(r.Context(), mig.PagePaginationCtxValue, page)
		ctx = context.WithValue(ctx, mig.PageSizePaginationCtxValue, size)

		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}

func withAuth(c *HttpApiController, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(fingerprintCookie)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		const prefix = "Bearer "

		if !strings.HasPrefix(authHeader, prefix) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, prefix)
		if token == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		valid, claims, err := c.authService.VerifyAccessToken(token)
		if err != nil || !valid {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		hash := auth.GetHash(cookie.Value)

		if hash != claims.UserFingerprint {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		next(w, r)
	}

	return http.HandlerFunc(fn)
}
