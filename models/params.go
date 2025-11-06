package models

type AuthParams struct {
	APIKey       string             `json:"api_key"`
	ClientID     string             `json:"client_id"`
	ClientSecret string             `json:"client_secret"`
	Data         *map[string]string `json:"data,omitempty"`
}

type CommonParams struct {
	// Page/Limit given priority over Cursor/Limit; Default use cursor-based pagination
	Page           int               `json:"page"`
	Cursor         string            `json:"cursor"`
	Limit          int               `json:"limit"`
	SortDescending bool              `json:"sort_descending,omitempty"`
	Extras         map[string]string `json:"extras,omitempty"`
}

type ContactQueryParams struct {
	CommonParams
	Search    string   `json:"search"`
	SearchIDs []string `json:"search_ids"`
}

// Sorted by timestamp ascending
type LedgerQueryParams struct {
	CommonParams
	CustomerID string    `json:"customer_id"`
	StartDate  string    `json:"start_date"`
	EndDate    string    `json:"end_date"`
	DocTypes   []DocType `json:"doc_types"`
}
