package utils

import (
	"encoding/base64"
	"reflect"
	"strconv"
	"testing"
)

func TestPaginateOffset(t *testing.T) {
	items := make([]int, 100)
	for i := 0; i < 100; i++ {
		items[i] = i + 1
	}

	t.Run("should paginate with valid page and perPage", func(t *testing.T) {
		paged, nextCursor := PaginateOffset(items, 2, 10)
		if len(paged) != 10 {
			t.Errorf("Expected 10 items, got %d", len(paged))
		}
		if paged[0] != 11 {
			t.Errorf("Expected first item to be 11, got %d", paged[0])
		}
		if nextCursor == nil {
			t.Errorf("Expected next cursor, got nil")
		} else {
			decoded, _ := base64.URLEncoding.DecodeString(*nextCursor)
			if string(decoded) != "20" {
				t.Errorf("Expected next cursor to be '20', got '%s'", string(decoded))
			}
		}
	})

	t.Run("should handle last page", func(t *testing.T) {
		paged, nextCursor := PaginateOffset(items, 10, 10)
		if len(paged) != 10 {
			t.Errorf("Expected 10 items, got %d", len(paged))
		}
		if paged[0] != 91 {
			t.Errorf("Expected first item to be 91, got %d", paged[0])
		}
		if nextCursor != nil {
			t.Errorf("Expected no next cursor, got %v", *nextCursor)
		}
	})

	t.Run("should handle invalid page and perPage", func(t *testing.T) {
		paged, _ := PaginateOffset(items, 0, 0)
		if len(paged) != 20 {
			t.Errorf("Expected 20 items, got %d", len(paged))
		}
	})

	t.Run("should handle empty slice", func(t *testing.T) {
		paged, nextCursor := PaginateOffset([]int{}, 1, 10)
		if len(paged) != 0 {
			t.Errorf("Expected 0 items, got %d", len(paged))
		}
		if nextCursor != nil {
			t.Errorf("Expected no next cursor, got %v", *nextCursor)
		}
	})
}

func TestPaginateCursor(t *testing.T) {
	items := make([]int, 100)
	for i := 0; i < 100; i++ {
		items[i] = i + 1
	}

	t.Run("should paginate with no cursor", func(t *testing.T) {
		paged, nextCursor := PaginateCursor(items, "", 10)
		expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		if !reflect.DeepEqual(paged, expected) {
			t.Errorf("Expected %v, got %v", expected, paged)
		}
		if nextCursor == nil {
			t.Fatal("Expected next cursor")
		}
		decoded, _ := base64.URLEncoding.DecodeString(*nextCursor)
		if val, _ := strconv.Atoi(string(decoded)); val != 10 {
			t.Errorf("Expected next cursor to be 10, got %d", val)
		}
	})

	t.Run("should paginate with valid cursor", func(t *testing.T) {
		cursor := base64.URLEncoding.EncodeToString([]byte("10"))
		paged, nextCursor := PaginateCursor(items, cursor, 10)
		expected := []int{11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
		if !reflect.DeepEqual(paged, expected) {
			t.Errorf("Expected %v, got %v", expected, paged)
		}
		if nextCursor == nil {
			t.Fatal("Expected next cursor")
		}
		decoded, _ := base64.URLEncoding.DecodeString(*nextCursor)
		if val, _ := strconv.Atoi(string(decoded)); val != 20 {
			t.Errorf("Expected next cursor to be 20, got %d", val)
		}
	})

	t.Run("should handle last page", func(t *testing.T) {
		cursor := base64.URLEncoding.EncodeToString([]byte("90"))
		paged, nextCursor := PaginateCursor(items, cursor, 10)
		expected := []int{91, 92, 93, 94, 95, 96, 97, 98, 99, 100}
		if !reflect.DeepEqual(paged, expected) {
			t.Errorf("Expected %v, got %v", expected, paged)
		}
		if nextCursor != nil {
			t.Errorf("Expected no next cursor, got %v", *nextCursor)
		}
	})

	t.Run("should handle invalid cursor", func(t *testing.T) {
		paged, _ := PaginateCursor(items, "invalid-cursor", 10)
		expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		if !reflect.DeepEqual(paged, expected) {
			t.Errorf("Expected %v, got %v", expected, paged)
		}
	})

	t.Run("should handle empty slice", func(t *testing.T) {
		paged, nextCursor := PaginateCursor([]int{}, "", 10)
		if len(paged) != 0 {
			t.Errorf("Expected 0 items, got %d", len(paged))
		}
		if nextCursor != nil {
			t.Errorf("Expected no next cursor, got %v", *nextCursor)
		}
	})
}
