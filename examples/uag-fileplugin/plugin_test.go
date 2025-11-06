package main

import (
	"testing"

	"github.com/nikhiljohn10/uagplugin/models"
	tk "github.com/nikhiljohn10/uagplugin/testkit"
)

func TestHealth(t *testing.T) {
	if Health() != "ok" {
		t.Fatalf("health not ok")
	}
}

func TestContactsBasic(t *testing.T) {
	tk.WithEnv(tk.TestVars{"UAG_TEST": "1"}, func() {
		out, err := Contacts(nil, models.ContactQueryParams{Search: "doe"})
		if err != nil {
			t.Fatalf("Contacts error: %v", err)
		}
		if out == nil || out.Count == 0 {
			t.Fatalf("expected some contacts")
		}
	})
}

func TestLedgerBasic(t *testing.T) {
	lg, err := Ledger(nil, models.LedgerQueryParams{})
	if err != nil {
		t.Fatalf("Ledger error: %v", err)
	}
	if len(lg.Entries) != 5 {
		t.Errorf("Expected 5 ledger entries, got %d", len(lg.Entries))
	}
}

func TestLedgerFilterByIDs(t *testing.T) {
	lg, err := Ledger(nil, models.LedgerQueryParams{CustomerID: "CUST-002"})
	if err != nil {
		t.Fatalf("Ledger error: %v", err)
	}
	if len(lg.Entries) != 2 {
		t.Errorf("Expected 2 ledger entries, got %d", len(lg.Entries))
	}
}
