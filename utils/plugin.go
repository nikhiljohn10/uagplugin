package utils

import (
	"sort"

	"github.com/nikhiljohn10/uagplugin/models"
)

func SortContacts(contacts *[]models.Contact, desc bool) {
	if desc {
		sort.SliceStable(*contacts, func(i, j int) bool {
			return (*contacts)[i].Name > (*contacts)[j].Name
		})
	} else {
		sort.SliceStable(*contacts, func(i, j int) bool {
			return (*contacts)[i].Name < (*contacts)[j].Name
		})
	}
}
