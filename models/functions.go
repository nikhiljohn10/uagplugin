package models

type InitialAuthCredentials = map[string]string
type AuthCredentials = map[string]string
type ApiCredentials = []string

type AuthFunc func(InitialAuthCredentials, AuthParams) (*AuthCredentials, error)
type ContactsFunc func(AuthCredentials, ContactQueryParams) (*Contacts, error)
type LedgerFunc func(AuthCredentials, LedgerQueryParams) (*Ledger, error)
type MetaFunc func() *MetaData
type HealthFunc func() string
