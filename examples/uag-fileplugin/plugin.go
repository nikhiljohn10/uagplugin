// Package main implements a reference file-based plugin for UAG.
//
// It demonstrates the minimal plugin "contract" expected by the host:
//   - Meta()    -> map[string]any     : describes the plugin and auth type
//   - Contacts(auth, params)          : fetches contacts with filter/sort/pagination
//   - Health()  -> string             : quick health probe ("ok" on success)
//
// This example reads contacts from an embedded CSV and showcases typical
// behaviors a real plugin would implement (filters, sorting, pagination).
package main

import (
	"github.com/nikhiljohn10/uagplugin/models"

	_ "embed"
)

//go:embed contact.csv
var contactCSV string

//go:embed ledger.csv
var ledgerCSV string

// Back-compat: keep top-level functions delegating to the instance
func Meta() map[string]any { return Plugin.Meta() }
func Health() string       { return Plugin.Health() }
func Contacts(a models.AuthCredentials, p models.Params) (*models.Contacts, error) {
	return Plugin.Contacts(a, p)
}
func Ledger(a models.AuthCredentials, p models.Params) (models.Ledger, error) {
	return Plugin.Ledger(a, p)
}
