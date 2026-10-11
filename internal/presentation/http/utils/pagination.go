package presentationhttputils

import (
	presentationhttpresponse "github.com/MateWorkspace/lockmate/internal/presentation/http/response"
	"github.com/gin-gonic/gin"
)

const (
	defaultPage  int64 = 1
	defaultLimit int64 = 10
	maxLimit     int64 = 100
)

type PaginationArgs struct {
	Page   int64
	Limit  int64
	Search *string
}

func PageArgs(c *gin.Context) (PaginationArgs, error) {
	page, err := QueryInt64(c, "page")
	if err != nil {
		return PaginationArgs{}, err
	}
	limit, err := QueryInt64(c, "limit")
	if err != nil {
		return PaginationArgs{}, err
	}
	args := PaginationArgs{Page: defaultPage, Limit: defaultLimit, Search: QueryString(c, "search")}
	if page != nil && *page > 0 {
		args.Page = *page
	}
	if limit != nil && *limit > 0 {
		args.Limit = *limit
	}
	if args.Limit > maxLimit {
		args.Limit = maxLimit
	}
	return args, nil
}

func PageResponse(args PaginationArgs, total int64) presentationhttpresponse.PageResponse {
	return presentationhttpresponse.PageResponse{Page: args.Page, Limit: args.Limit, TotalItems: total}
}
