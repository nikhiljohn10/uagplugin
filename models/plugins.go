package models

import "strings"

type DocType string

const (
	DocTypeInvoice      DocType = "invoice"
	DocTypePayment      DocType = "payment"
	DocTypeRefund       DocType = "refund"
	DocTypeCreditNote   DocType = "credit_note"
	DocTypeCreditRefund DocType = "credit_refund"
	DocTypeDebitNote    DocType = "debit_note"
	DocTypeJournal      DocType = "journal"
)

func (dt DocType) String() string {
	return strings.ToTitle(strings.ReplaceAll(string(dt), "_", " "))
}

type Contact struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Contacts struct {
	Source     string    `json:"source"`
	Items      []Contact `json:"contacts"`
	Count      int       `json:"count"`
	Total      int       `json:"total"`
	NextCursor *string   `json:"next_cursor,omitempty"`
}

type LedgerEntry struct {
	ID      int64   `json:"id"`
	Date    string  `json:"date"`
	DocType DocType `json:"doc_type"`
	Amount  string  `json:"amount"`
}

type Ledger struct {
	Entries        []LedgerEntry `json:"entries"`
	CustomerName   string        `json:"customer_name"`
	OpeningBalance string        `json:"opening_balance"`
	NextCursor     *string       `json:"next_cursor,omitempty"`
}
