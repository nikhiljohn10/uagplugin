package main

import (
	"encoding/csv"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/nikhiljohn10/uagplugin/models"
	"github.com/nikhiljohn10/uagplugin/typing"
)

// Meta returns basic information about the plugin such as id, name, version
// and the kind of authentication it requires ("none" for this demo plugin).
type filePlugin struct{}

// Exported symbol for typed host loading
var Plugin typing.Plugin = filePlugin{}

func (filePlugin) Meta() map[string]any {
	return map[string]any{
		"platform_id":      "fileplugin",
		"platform_name":    "File Plugin",
		"version":          "1.0.0",
		"author":           "Nikhil John",
		"auth_type":        "none",
		"contract_version": typing.ContractVersion,
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
func (filePlugin) Contacts(auth models.AuthCredentials, params models.Params) (*models.Contacts, error) {
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
		// Filter by explicit IDs, if provided
		if len(params.SearchIDs) > 0 {
			if i := slices.IndexFunc(params.SearchIDs, func(n string) bool { return n == rec[0] }); i < 0 {
				continue
			}
		}
		// Filter by search text against name (case-insensitive)
		if params.SearchText != "" && !strings.Contains(
			strings.ToLower(rec[1]),
			strings.ToLower(params.SearchText),
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

	// Optional sorting by name
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

	// Cursor-based pagination (page size 20)
	pagedContacts, nextCursor := models.PaginateCursor(contacts, params.Cursor, 20)
	return &models.Contacts{
		Source:     src,
		Items:      pagedContacts,
		Count:      len(pagedContacts),
		Total:      len(contacts),
		NextCursor: nextCursor,
	}, nil
}

// Health returns a simple constant to indicate the plugin is responsive.
func (filePlugin) Health() string { return "ok" }

// Ledger implements the core interface; demo returns an empty ledger object.
func (filePlugin) Ledger(auth models.AuthCredentials, params models.Params) (models.Ledger, error) {
	r := csv.NewReader(strings.NewReader(ledgerCSV))
	rows, err := r.ReadAll()
	if err != nil {
		return models.Ledger{}, err
	}
	var entries []models.LedgerEntry
	for i, rec := range rows {
		if i == 0 { // skip header
			continue
		}
		if len(rec) < 5 {
			continue
		}
		// Optional filter by SearchIDs (match on id column as string)
		if len(params.SearchIDs) > 0 {
			if idx := slices.IndexFunc(params.SearchIDs, func(s string) bool { return s == strings.TrimSpace(rec[0]) }); idx < 0 {
				continue
			}
		}
		// Parse ID
		var id int64
		if v, err := strconv.ParseInt(strings.TrimSpace(rec[0]), 10, 64); err == nil {
			id = v
		}
		date := strings.TrimSpace(rec[1])
		docType := models.DocType(strings.TrimSpace(rec[2]))
		var ptype *models.PaymentType
		if pt := strings.TrimSpace(rec[3]); pt != "" {
			v := models.PaymentType(pt)
			ptype = &v
		}
		amount := strings.TrimSpace(rec[4])
		entries = append(entries, models.LedgerEntry{
			ID:          id,
			Date:        date,
			DocType:     docType,
			PaymentType: ptype,
			Amount:      amount,
		})
	}
	// For demo purposes keep balances zero; entries are what tests care about.
	return models.Ledger{ID: 1, Entries: entries, CreditBalance: "0", CreditLimit: "0"}, nil
}
