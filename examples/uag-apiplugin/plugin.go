// Package main implements a reference API-backed plugin for UAG.
//
// It demonstrates the minimal plugin "contract" expected by the host:
//   - Meta()    -> map[string]any     : describes the plugin and auth type
//   - Contacts(auth, params)          : fetches contacts with filter/sort/pagination
//   - Health()  -> string             : quick health probe ("ok" on success)
//   - Auth(...) -> error              : optional, validates credentials when needed
//
// This example fetches contacts from a public demo API (jsonplaceholder), and
// shows how to support filters, sorting, and cursor pagination.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/nikhiljohn10/uagplugin/models"
	"github.com/nikhiljohn10/uagplugin/typing"
	"github.com/nikhiljohn10/uagplugin/utils"
)

// Default public demo API (can be overridden via params.Extra["base_url"] or
// environment variable API_BASE_URL).
const defaultBaseURL = "https://jsonplaceholder.typicode.com"

// Meta returns basic information about the plugin such as id, name, version
// and the kind of authentication it requires ("none" for this demo plugin).
type apiPlugin struct{}

// Export typed Plugin symbol
var Plugin typing.Plugin = apiPlugin{}

func (apiPlugin) Meta() map[string]any {
	return map[string]any{
		"platform_id":      "apiplugin",
		"platform_name":    "API Plugin",
		"version":          "1.0.0",
		"author":           "UAG",
		"auth_type":        "none",
		"contract_version": typing.ContractVersion,
	}
}

// Health returns a simple constant to indicate the plugin is responsive.
func (apiPlugin) Health() string { return "ok" }

// Auth validates credentials if the target API requires them. This demo API is
// public, so Auth is a no-op and returns nil.
func Auth(_ models.AuthCredentials, _ models.Params) error { return nil }

// API user shape we care about
type apiUser struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// baseURL resolves the API endpoint, in this order:
//  1. params.Extra["base_url"]
//  2. API_BASE_URL environment variable
//  3. defaultBaseURL
func baseURL(params models.Params) string {
	if params.Extra != nil {
		if v, ok := params.Extra["base_url"]; ok && strings.TrimSpace(v) != "" {
			return v
		}
	}
	if v := strings.TrimSpace(os.Getenv("API_BASE_URL")); v != "" {
		return v
	}
	return defaultBaseURL
}

// Contacts fetches contacts from the demo API and applies filtering/sorting/pagination.
//
// Supported Params:
//   - SearchText: case-insensitive substring match on contact name or email
//   - SearchIDs:  restrict results to the given list of string IDs
//   - Sort (+ SortOrder): sort by Name (asc|desc), default asc
//   - Cursor: cursor-based pagination position (base64 encoded index)
//
// If the API requires authentication in the future, headers can be set based on
// the provided AuthCredentials (e.g., bearer token).
func (apiPlugin) Contacts(auth models.AuthCredentials, params models.Params) (*models.Contacts, error) {
	// Build URL of the users endpoint
	url := strings.TrimRight(baseURL(params), "/") + "/users"

	// Fetch users from API
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	// If future auth is required, read from auth and set headers here
	// Example: if token, ok := auth["token"]; ok { req.Header.Set("Authorization", "Bearer "+token) }
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var users []apiUser
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}

	// Map API users to plugin contacts
	var contacts []models.Contact
	for _, u := range users {
		contacts = append(contacts, models.Contact{
			ID:    strconv.Itoa(u.ID),
			Name:  u.Name,
			Email: u.Email,
		})
	}

	// Filter by SearchText (name or email)
	if st := strings.ToLower(strings.TrimSpace(params.SearchText)); st != "" {
		filtered := make([]models.Contact, 0, len(contacts))
		for _, c := range contacts {
			if strings.Contains(strings.ToLower(c.Name), st) || strings.Contains(strings.ToLower(c.Email), st) {
				filtered = append(filtered, c)
			}
		}
		contacts = filtered
	}

	// Filter by explicit IDs, if provided
	if len(params.SearchIDs) > 0 {
		allowed := map[string]struct{}{}
		for _, id := range params.SearchIDs {
			allowed[strings.TrimSpace(id)] = struct{}{}
		}
		filtered := make([]models.Contact, 0, len(contacts))
		for _, c := range contacts {
			if _, ok := allowed[c.ID]; ok {
				filtered = append(filtered, c)
			}
		}
		contacts = filtered
	}

	// Optional sorting by name
	if params.Sort {
		asc := strings.ToLower(params.SortOrder) != "desc"
		sort.SliceStable(contacts, func(i, j int) bool {
			if asc {
				return contacts[i].Name < contacts[j].Name
			}
			return contacts[i].Name > contacts[j].Name
		})
	}

	// Cursor-based pagination (page size 20)
	items, next := utils.PaginateCursor(contacts, params.Cursor, 20)
	src, _ := Meta()["platform_id"].(string)
	return &models.Contacts{
		Source:     src,
		Items:      items,
		Count:      len(items),
		Total:      len(contacts),
		NextCursor: next,
	}, nil
}

// Ledger implements the core interface; demo returns an empty ledger.
func (apiPlugin) Ledger(auth models.AuthCredentials, params models.Params) (*models.Ledger, error) {
	return &models.Ledger{Entries: nil, CustomerName: "", OpeningBalance: "0"}, nil
}

// Back-compat: keep top-level functions delegating to the instance
func Meta() map[string]any { return Plugin.Meta() }
func Health() string       { return Plugin.Health() }
func Contacts(a models.AuthCredentials, p models.Params) (*models.Contacts, error) {
	return Plugin.Contacts(a, p)
}
func Ledger(a models.AuthCredentials, p models.Params) (*models.Ledger, error) {
	return Plugin.Ledger(a, p)
}
