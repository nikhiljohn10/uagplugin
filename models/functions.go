package models

type AuthCredentials = map[string]string
type MetaData = map[string]any

type Params struct {
	SearchText string   `json:"search"`
	Cursor     string   `json:"cursor"`
	Sort       bool     `json:"sort,omitempty"`
	SortOrder  string   `json:"sort_order,omitempty"`
	SearchIDs  []string `json:"search_ids,omitempty"`
	// Extra can hold any other parameters that might be needed.
	Extra map[string]string `json:"extra,omitempty"`
}

type AuthFunc func(AuthCredentials, Params) error
type ContactsFunc func(AuthCredentials, Params) (*Contacts, error)
type LedgerFunc func(AuthCredentials, Params) (*Ledger, error)
type HealthFunc func() string
