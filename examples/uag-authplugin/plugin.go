package main

import (
	"errors"
	"fmt"

	"github.com/nikhiljohn10/uagplugin/models"
	"github.com/nikhiljohn10/uagplugin/typing"
)

type BasePlugin struct{}

var Plugin typing.Plugin = &BasePlugin{}
var _ typing.Plugin = (*BasePlugin)(nil)
var _ typing.Authenticator = (*BasePlugin)(nil)

// Meta returns the plugin's metadata.
func (p *BasePlugin) Meta() *models.MetaData {
	return &models.MetaData{
		ID:              "auth-plugin-example",
		Name:            "Authentication Example Plugin",
		Version:         "1.0.0",
		ContractVersion: typing.ContractVersion,
		AuthType:        "none",
		Author:          "Nikhil John",
		Description:     "An example plugin demonstrating authentication handling.",
		AuthCredentials: &models.InitialAuthCredentials{
			"api_key":         "API key for simple authentication",
			"client_id":       "Client ID for OAuth2",
			"client_secret":   "Client secret for OAuth2",
			"organization_id": "Organization ID for multi-tenancy",
		},
		ApiCredentials: &models.ApiCredentials{"token", "expires_in"},
	}
}

// Health returns the health status of the plugin.
func (p *BasePlugin) Health() string {
	return "ok"
}

// Auth is the implementation of the optional Authenticator interface.
func (p *BasePlugin) Auth(params models.AuthParams) (*models.AuthCredentials, error) {
	fmt.Println("Attempting authentication...")
	if params.Data == nil {
		return nil, errors.New("organization ID is required")
	}
	if org_id, ok := (*params.Data)["organization_id"]; !ok || org_id == "" {
		return nil, errors.New("organization ID is required")
	}

	if params.APIKey == "" {
		return nil, errors.New("API key is required")
	}
	if params.APIKey != "secret-key" {
		return nil, errors.New("invalid API key")
	}
	fmt.Println("API key authentication successful (simulated).")

	if params.ClientID == "" || params.ClientSecret == "" {
		return nil, errors.New("client ID and client secret are required for OAuth2")
	}
	fmt.Println("OAuth2 authentication successful (simulated).")
	return &models.AuthCredentials{
		"token":      "access-token",
		"expires_in": "3600",
	}, nil
}

// Contacts returns a list of contacts. This plugin requires authentication first.
func (p *BasePlugin) Contacts(auth models.AuthCredentials, params models.ContactQueryParams) (*models.Contacts, error) {
	// In a real plugin, you'd use the auth credentials to make an API call.
	// Here, we'll just return a dummy list.
	return &models.Contacts{
		Items: []models.Contact{
			{ID: "1", Name: "Authenticated User", Email: "auth@example.com"},
		},
		Count: 1,
		Total: 1,
	}, nil
}

// Ledger is not implemented in this example.
func (p *BasePlugin) Ledger(auth models.AuthCredentials, params models.LedgerQueryParams) (*models.Ledger, error) {
	entry := models.LedgerEntry{
		ID:      1,
		Date:    "2024-01-01",
		DocType: models.DocTypeInvoice,
		Amount:  "100.00",
	}
	return &models.Ledger{
		Entries:        []models.LedgerEntry{entry},
		CustomerName:   "Authenticated Customer",
		OpeningBalance: "0.00",
	}, nil
}
