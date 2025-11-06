package main

import (
	"encoding/csv"
	"slices"
	"strconv"
	"strings"

	"github.com/nikhiljohn10/uagplugin/models"
	"github.com/nikhiljohn10/uagplugin/typing"
	"github.com/nikhiljohn10/uagplugin/utils"
)

// Meta returns basic information about the plugin such as id, name, version
// and the kind of authentication it requires ("none" for this demo plugin).
type filePlugin struct{}

// Exported symbol for typed host loading
var Plugin typing.Plugin = filePlugin{}

var _ typing.Plugin = (*filePlugin)(nil)

func (filePlugin) Meta() *models.MetaData {
	return &models.MetaData{
		ID:              "fileplugin",
		Name:            "File Plugin",
		Version:         "1.0.0",
		Author:          "Nikhil John",
		AuthType:        "none",
		ContractVersion: typing.ContractVersion,
	}
}

// Contacts reads the embedded CSV and returns a paginated list of contacts.
//
// Supported Params:
//   - SearchText: case-insensitive substring match on contact name
//   - SearchIDs:  restrict results to the given list of IDs
//   - Sort (+ SortOrder): sort by Name (asc|desc), default asc
//   - Cursor: cursor-based pagination position (base64 encoded index)
//
// The auth parameter is unused here because this plugin is file-based.
func (filePlugin) Contacts(auth models.AuthCredentials, params models.ContactQueryParams) (*models.Contacts, error) {
	// 'auth' is not used as this is a file-based plugin
	r := csv.NewReader(strings.NewReader(contactCSV))
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	var contacts []models.Contact
	for _, rec := range records[1:] {
		if len(rec) < 3 {
			continue
		}
		// Filter by explicit IDs, if provided
		if len(params.SearchIDs) > 0 {
			if i := slices.IndexFunc(params.SearchIDs, func(n string) bool { return n == rec[0] }); i < 0 {
				continue
			}
		}
		// Filter by search text against name (case-insensitive)
		if params.Search != "" && !strings.Contains(
			strings.ToLower(rec[1]),
			strings.ToLower(params.Search),
		) {
			continue
		}
		// Add the contact
		contacts = append(contacts, models.Contact{
			ID:    rec[0],
			Name:  rec[1],
			Email: rec[2],
		})
	}

	utils.SortContacts(&contacts, params.SortDescending)

	// Cursor-based pagination (page size 20)
	pagedContacts, nextCursor := utils.PaginateCursor(contacts, params.Cursor, 20)

	return &models.Contacts{
		Items:      pagedContacts,
		Count:      len(pagedContacts),
		Total:      len(contacts),
		NextCursor: nextCursor,
	}, nil
}

// Health returns a simple constant to indicate the plugin is responsive.
func (filePlugin) Health() string { return "ok" }

// Ledger reads the embedded CSV and returns a paginated list of ledger entries.
func (filePlugin) Ledger(auth models.AuthCredentials, params models.LedgerQueryParams) (*models.Ledger, error) {
	r := csv.NewReader(strings.NewReader(ledgerCSV))
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	var entries []models.LedgerEntry
	for _, record := range records[1:] { // Skip header
		id, _ := strconv.ParseInt(record[0], 10, 64)
		entries = append(entries, models.LedgerEntry{
			ID:      id,
			Date:    record[1],
			DocType: models.DocType(record[2]),
			Amount:  record[3],
		})
	}

	// Paginate with a page size of 5
	paginatedEntries, nextCursor := utils.PaginateCursor(entries, params.Cursor, 5)

	return &models.Ledger{
		Entries:        paginatedEntries,
		CustomerName:   "File-based Customer",
		OpeningBalance: "100.00",
		NextCursor:     nextCursor,
	}, nil
}
