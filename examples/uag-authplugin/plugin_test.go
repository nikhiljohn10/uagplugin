package main

import (
	"strings"
	"testing"

	"github.com/nikhiljohn10/uagplugin/models"
)

// TestMeta verifies that the plugin's metadata is returned correctly.
func TestMeta(t *testing.T) {
	p := &BasePlugin{}
	meta := p.Meta()

	if meta == nil {
		t.Fatal("Meta should not be nil")
	}
	if meta.ID != "auth-plugin-example" {
		t.Errorf("Expected meta ID 'auth-plugin-example', got '%s'", meta.ID)
	}
	if meta.Name != "Authentication Example Plugin" {
		t.Errorf("Expected meta name 'Authentication Example Plugin', got '%s'", meta.Name)
	}
	if meta.Version == "" {
		t.Error("Version should not be empty")
	}
	if meta.ContractVersion == "" {
		t.Error("ContractVersion should not be empty")
	}
	if meta.AuthCredentials == nil {
		t.Error("AuthCredentials should not be nil")
	}
	if meta.ApiCredentials == nil {
		t.Error("ApiCredentials should not be nil")
	}
}

// TestHealth checks that the health check returns "ok".
func TestHealth(t *testing.T) {
	p := &BasePlugin{}
	health := p.Health()
	if health != "ok" {
		t.Errorf("Expected health to be 'ok', got '%s'", health)
	}
}

// TestAuth covers success and failure scenarios for the Auth method.
func TestAuth(t *testing.T) {
	p := &BasePlugin{}

	t.Run("Successful Authentication", func(t *testing.T) {
		apiKey := "secret-key"
		clientID := "test-client"
		clientSecret := "test-secret"
		params := models.AuthParams{
			APIKey:       apiKey,
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Data: &models.AuthCredentials{
				"organization_id": "test-org",
			},
		}

		creds, err := p.Auth(params)
		if err != nil {
			t.Fatalf("Authentication should succeed, but got error: %v", err)
		}
		if creds == nil {
			t.Fatal("Credentials should not be nil on success")
		}
		if (*creds)["token"] != "access-token" {
			t.Errorf("Expected token 'access-token', got '%s'", (*creds)["token"])
		}
	})

	t.Run("Failure - Missing Organization ID", func(t *testing.T) {
		apiKey := "secret-key"
		params := models.AuthParams{APIKey: apiKey}
		_, err := p.Auth(params)
		if err == nil {
			t.Fatal("Should fail without organization ID, but got no error")
		}
		if !strings.Contains(err.Error(), "organization ID is required") {
			t.Errorf("Expected error to contain 'organization ID is required', got '%s'", err.Error())
		}
	})

	t.Run("Failure - Missing API Key", func(t *testing.T) {
		params := models.AuthParams{
			Data: &models.AuthCredentials{
				"organization_id": "test-org",
			},
		}
		_, err := p.Auth(params)
		if err == nil {
			t.Fatal("Should fail without API key, but got no error")
		}
		if !strings.Contains(err.Error(), "API key is required") {
			t.Errorf("Expected error to contain 'API key is required', got '%s'", err.Error())
		}
	})

	t.Run("Failure - Invalid API Key", func(t *testing.T) {
		apiKey := "wrong-key"
		params := models.AuthParams{
			APIKey: apiKey,
			Data: &models.AuthCredentials{
				"organization_id": "test-org",
			},
		}
		_, err := p.Auth(params)
		if err == nil {
			t.Fatal("Should fail with invalid API key, but got no error")
		}
		if !strings.Contains(err.Error(), "invalid API key") {
			t.Errorf("Expected error to contain 'invalid API key', got '%s'", err.Error())
		}
	})

	t.Run("Failure - Missing Client ID/Secret", func(t *testing.T) {
		apiKey := "secret-key"
		params := models.AuthParams{
			APIKey: apiKey,
			Data: &models.AuthCredentials{
				"organization_id": "test-org",
			},
		}
		_, err := p.Auth(params)
		if err == nil {
			t.Fatal("Should fail without client credentials, but got no error")
		}
		if !strings.Contains(err.Error(), "client ID and client secret are required") {
			t.Errorf("Expected error to contain 'client ID and client secret are required', got '%s'", err.Error())
		}
	})
}

// TestContacts verifies the Contacts method returns correct data.
func TestContacts(t *testing.T) {
	p := &BasePlugin{}
	contacts, err := p.Contacts(nil, models.ContactQueryParams{})

	if err != nil {
		t.Fatalf("Contacts should not return an error, but got: %v", err)
	}
	if contacts == nil {
		t.Fatal("Contacts result should not be nil")
	}
	if contacts.Count != 1 {
		t.Errorf("Expected count to be 1, got %d", contacts.Count)
	}
	if len(contacts.Items) != 1 {
		t.Fatalf("Expected 1 contact item, got %d", len(contacts.Items))
	}
	if contacts.Items[0].Name != "Authenticated User" {
		t.Errorf("Expected contact name 'Authenticated User', got '%s'", contacts.Items[0].Name)
	}
}

// TestLedger verifies the Ledger method returns correct data.
func TestLedger(t *testing.T) {
	p := &BasePlugin{}
	ledger, err := p.Ledger(nil, models.LedgerQueryParams{})

	if err != nil {
		t.Fatalf("Ledger should not return an error, but got: %v", err)
	}
	if ledger == nil {
		t.Fatal("Ledger result should not be nil")
	}
	if len(ledger.Entries) != 1 {
		t.Fatalf("Expected 1 ledger entry, got %d", len(ledger.Entries))
	}
	if ledger.CustomerName != "Authenticated Customer" {
		t.Errorf("Expected customer name 'Authenticated Customer', got '%s'", ledger.CustomerName)
	}
	if ledger.Entries[0].DocType != models.DocTypeInvoice {
		t.Errorf("Expected DocType to be 'invoice', got '%s'", ledger.Entries[0].DocType)
	}
}
