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
		out, err := Contacts(nil, models.Params{SearchText: "doe"})
		if err != nil {
			t.Fatalf("Contacts error: %v", err)
		}
		if out == nil || out.Count == 0 {
			t.Fatalf("expected some contacts")
		}
	})
}

func TestLedgerBasic(t *testing.T) {
	lg, err := Ledger(nil, models.Params{})
	if err != nil {
		t.Fatalf("Ledger error: %v", err)
	}
	if len(lg.Entries) == 0 {
		t.Fatalf("expected some ledger entries, got 0")
	}
	// Check first row mapping
	e := lg.Entries[0]
	if e.ID != 1 || e.Date != "2025-01-05" || e.DocType != models.DocTypeSaleInvoice {
		t.Fatalf("unexpected first entry: %+v", e)
	}
	if e.PaymentType != nil { // sale_invoice should not set payment type
		t.Fatalf("unexpected payment type for non-payment doc: %+v", *e.PaymentType)
	}
}

func TestLedgerFilterByIDs(t *testing.T) {
	lg, err := Ledger(nil, models.Params{SearchIDs: []string{"2", "5"}})
	if err != nil {
		t.Fatalf("Ledger error: %v", err)
	}
	if len(lg.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(lg.Entries))
	}
	// Verify payment types for the filtered entries
	for _, e := range lg.Entries {
		if e.ID == 2 {
			if e.PaymentType == nil || *e.PaymentType != models.PaymentTypeIn {
				t.Fatalf("entry 2 should be payment in: %+v", e)
			}
		}
		if e.ID == 5 {
			if e.PaymentType == nil || *e.PaymentType != models.PaymentTypeOut {
				t.Fatalf("entry 5 should be payment out: %+v", e)
			}
		}
	}
}
