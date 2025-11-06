package typing

import "github.com/nikhiljohn10/uagplugin/models"

// Plugin defines the minimal contract a UAG plugin must implement.
// Implementations should be exported from the plugin as:
//
//	var Plugin pluginapi.Plugin = MyPlugin{}
type Plugin interface {
	// Meta returns basic info (id, name, version, etc.).
	Meta() models.MetaData
	// Health performs a quick health probe, should return "ok" when healthy.
	Health() string
	// Contacts returns a (possibly paginated) list of contacts.
	Contacts(auth models.AuthCredentials, params models.Params) (*models.Contacts, error)
	// Ledger returns ledger data for the provided parameters.
	Ledger(auth models.AuthCredentials, params models.Params) (*models.Ledger, error)
}

// Tester is an optional interface; if implemented, the host will call RunTests.
type Tester interface {
	RunTests() error
}
