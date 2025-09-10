package main

import (
	"slices"

	"github.com/nikhiljohn10/uagplugin/models"

	_ "embed"
	"encoding/csv"
	"sort"
	"strings"
)

//go:embed contact.csv
var contactCSV string

func Meta() map[string]any {
	return map[string]any{
		"platform_id":   "fileplugin",
		"platform_name": "File Plugin",
		"version":       "1.0.0",
		"author":        "Nikhil John",
		"auth_type":     "none",
	}
}

func Contacts(auth models.AuthCredentials, params models.Params) (*models.Contacts, error) {
	// 'auth' is not used as this is a file-based plugin
	src, _ := Meta()["platform_id"].(string)
	r := csv.NewReader(strings.NewReader(contactCSV))
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	var contacts []models.Contact
	for _, rec := range records {
		if len(rec) < 3 {
			continue
		}
		if len(params.SearchIDs) > 0 {
			if i := slices.IndexFunc(params.SearchIDs, func(n string) bool { return n == rec[0] }); i < 0 {
				continue
			}
		}
		if params.SearchText != "" && !strings.Contains(
			strings.ToLower(rec[1]),
			strings.ToLower(params.SearchText),
		) {
			continue
		}
		contacts = append(contacts, models.Contact{
			ID:    rec[0],
			Name:  rec[1],
			Email: rec[2],
		})
	}

	if params.Sort {
		if params.SortOrder == "" || strings.ToLower(params.SortOrder) == "asc" {
			sort.SliceStable(contacts, func(i, j int) bool {
				return contacts[i].Name < contacts[j].Name
			})
		} else {
			sort.SliceStable(contacts, func(i, j int) bool {
				return contacts[i].Name > contacts[j].Name
			})
		}
	}

	pagedContacts, nextCursor := models.PaginateCursor(contacts, params.Cursor, 20)
	return &models.Contacts{
		Source:     src,
		Items:      pagedContacts,
		Count:      len(pagedContacts),
		Total:      len(contacts),
		NextCursor: nextCursor,
	}, nil
}

func Health() string { return "ok" }
