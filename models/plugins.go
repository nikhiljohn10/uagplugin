package models

type PaymentType string
type DocType string

const (
	PaymentTypeIn  PaymentType = "in"
	PaymentTypeOut PaymentType = "out"

	DocTypeSaleInvoice     DocType = "sale_invoice"
	DocTypePurchaseInvoice DocType = "purchase_invoice"
	DocTypeDebitNote       DocType = "debit_note"
	DocTypeCreditNote      DocType = "credit_note"
	DocTypePayment         DocType = "payment"
)

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
	ID          int64        `json:"id"`
	Date        string       `json:"date"`
	DocType     DocType      `json:"doc_type"`
	PaymentType *PaymentType `json:"payment_type"`
	Amount      string       `json:"amount"`
}

type Ledger struct {
	ID            int64         `json:"id"`
	Entries       []LedgerEntry `json:"entries"`
	CreditBalance string        `json:"credit_balance"`
	CreditLimit   string        `json:"credit_limit"`
}
