package models

import (
	"strings"
)

type DocType string
type AuthType string

const (
	DocTypeInvoice      DocType = "invoice"
	DocTypePayment      DocType = "payment"
	DocTypeRefund       DocType = "refund"
	DocTypeCreditNote   DocType = "credit_note"
	DocTypeCreditRefund DocType = "credit_refund"
	DocTypeDebitNote    DocType = "debit_note"
	DocTypeJournal      DocType = "journal"

	AuthTypeAPIKey AuthType = "api_key"
	AuthTypeOAuth2 AuthType = "oauth2"
	AuthTypeNone   AuthType = "none"
)

func (dt DocType) String() string {
	return strings.ToTitle(strings.ReplaceAll(string(dt), "_", " "))
}

type MetaData struct {
	ID              string           `json:"id"`
	Name            string           `json:"name"`
	Version         string           `json:"version"`
	Description     string           `json:"description,omitempty"`
	Author          string           `json:"author"`
	AuthType        AuthType         `json:"auth_type"`
	ContractVersion string           `json:"contract_version"`
	AuthCredentials *AuthCredentials `json:"auth_credentials,omitempty"`
	ApiCredentials  *ApiCredentials  `json:"api_credentials,omitempty"`
}

type Contact struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Contacts struct {
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
