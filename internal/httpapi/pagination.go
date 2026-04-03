package httpapi

import (
	"fmt"
	"math"
	"net/url"

	"github.com/gin-gonic/gin"
)

type PageRequest struct {
	Page     int32 `form:"page"      binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=100"`
}

type PageParams struct {
	Limit  int32
	Offset int32
}

type PageResponse[T any] struct {
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Count    int64   `json:"count"`
	PageSize int32   `json:"page_size"`
	Results  []T     `json:"results"`
}

func (r PageRequest) Params() PageParams {
	return PageParams{
		Limit:  r.PageSize,
		Offset: (r.Page - 1) * r.PageSize,
	}
}

func NewPageResponse[T any](
	ctx *gin.Context,
	req PageRequest,
	results []T,
	totalCount int64,
) PageResponse[T] {
	totalPages := int32(math.Ceil(float64(totalCount) / float64(req.PageSize)))
	baseURL := *ctx.Request.URL

	var nextLink *string
	if req.Page < totalPages {
		next := createPaginationURL(&baseURL, req.Page+1, req)
		nextLink = &next
	}

	var prevLink *string
	if req.Page > 1 {
		prev := createPaginationURL(&baseURL, req.Page-1, req)
		prevLink = &prev
	}

	return PageResponse[T]{
		Next:     nextLink,
		Previous: prevLink,
		Count:    totalCount,
		PageSize: req.PageSize,
		Results:  results,
	}
}

func createPaginationURL(baseURL *url.URL, page int32, req PageRequest) string {
	q := baseURL.Query()
	q.Set("page", fmt.Sprintf("%d", page))
	q.Set("page_size", fmt.Sprintf("%d", req.PageSize))
	baseURL.RawQuery = q.Encode()
	return baseURL.String()
}
