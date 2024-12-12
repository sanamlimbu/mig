package mig

import (
	"net/http"
)

type PaginationCtxValue string

const (
	PagePaginationCtxValue     PaginationCtxValue = "page"
	PageSizePaginationCtxValue PaginationCtxValue = "page-size"
)

type Pagination struct {
	Page     int
	PageSize int
}

func NewPagination(r *http.Request) Pagination {
	return Pagination{
		Page:     r.Context().Value(PagePaginationCtxValue).(int),
		PageSize: r.Context().Value(PageSizePaginationCtxValue).(int),
	}
}
