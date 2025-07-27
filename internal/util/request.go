package util

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func GetParamUUID(r *http.Request, param string) (uuid.UUID, error) {
	return uuid.Parse(mux.Vars(r)[param])
}

func GetQueryStr(r *http.Request, query string) string {
	q := r.URL.Query()
	return q.Get(query)
}

func GetQueryUUID(r *http.Request, query string) (uuid.UUID, error) {
	q := r.URL.Query()
	return uuid.Parse(q.Get(query))
}

func GetQueryInt(r *http.Request, query string) (int, error) {
	q := r.URL.Query()
	val, err := strconv.ParseInt(q.Get(query), 10, 64)
	if err != nil {
		return 0, err
	}
	return int(val), nil
}

// GetPaginationQuery retrieves and composes the skip(offset) and limit params. It defaults to the one specified if none is provided
func GetPaginationQuery(r *http.Request, pageDefault, pageSizeDefault int) (int, int) {
	page, err := GetQueryInt(r, "page")
	if err != nil {
		page = pageDefault
	}
	pageSize, err := GetQueryInt(r, "pageSize")
	if err != nil {
		pageSize = pageSizeDefault
	}

	// absolute defaults
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// compose page
	skip := (page - 1) * pageSize
	return skip, pageSize
}
