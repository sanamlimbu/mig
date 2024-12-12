package api

import (
	"context"
	"mig"
	"net/http"
	"strconv"
)

const pageSize int = 10

// paginate is a middleware that ensures the "page" and "page_size" query parameters are valid and greater than or equal to 1.
// If the parameters are invalid or not provided, it falls back to default values: page = 1 and page_size = 10.
// These values are then stored in the request context for use by subsequent handlers.
func paginate(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil || page <= 0 {
			page = 0
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
