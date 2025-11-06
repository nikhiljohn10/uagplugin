package utils

import (
	"encoding/base64"
	"strconv"
)

// PaginateOffset paginates using offset/page/perPage and returns paged items and base64 nextCursor
func PaginateOffset[T any](items []T, page, perPage int) ([]T, *string) {
	if perPage <= 0 {
		perPage = 20
	}
	if page <= 0 {
		page = 1
	}
	start := min((page-1)*perPage, len(items))
	end := min(start+perPage, len(items))
	paged := items[start:end]
	var nextCursor *string
	if end < len(items) {
		nc := strconv.Itoa(end)
		enc := base64.URLEncoding.EncodeToString([]byte(nc))
		nextCursor = &enc
	}
	return paged, nextCursor
}

// PaginateCursor paginates using base64 cursor and returns paged items and base64 nextCursor
func PaginateCursor[T any](items []T, cursor string, pageSize int) ([]T, *string) {
	if pageSize <= 0 {
		pageSize = 20
	}
	start := 0
	if cursor != "" {
		decoded, err := base64.URLEncoding.DecodeString(cursor)
		if err == nil {
			if n, err := strconv.Atoi(string(decoded)); err == nil && n >= 0 {
				start = n
			}
		}
	}
	end := min(start+pageSize, len(items))
	paged := items[start:end]
	var nextCursor *string
	if end < len(items) {
		nc := strconv.Itoa(end)
		enc := base64.URLEncoding.EncodeToString([]byte(nc))
		nextCursor = &enc
	}
	return paged, nextCursor
}
