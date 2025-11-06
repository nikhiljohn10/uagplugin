package testkit

import (
	"net/http"
	"os"
	"testing"

	"github.com/nikhiljohn10/uagplugin/models"
)

func TestJSONResponse(t *testing.T) {
	t.Run("should return a valid json response for contacts", func(t *testing.T) {
		contact := models.Contact{
			ID:    "1",
			Name:  "John Doe",
			Email: "john@gmail.com",
		}
		_, url := StartMockServer(map[string]http.Handler{
			"/contacts": JSONResponse(http.StatusOK, models.Contacts{
				Items: []models.Contact{contact},
				Count: 1,
			}),
		})
		defer t.Cleanup(func() {
			http.DefaultClient.CloseIdleConnections()
		})

		res, err := http.Get(url + "/contacts")
		if err != nil {
			t.Errorf("Error while making request: %v", err)
		}
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, res.StatusCode)
		}
	})

	t.Run("should return a valid json response for ledger", func(t *testing.T) {
		entry := models.LedgerEntry{
			ID:      1,
			Date:    "2025-01-01",
			DocType: models.DocTypeInvoice,
			Amount:  "100",
		}
		_, url := StartMockServer(map[string]http.Handler{
			"/ledger": JSONResponse(http.StatusOK, &models.Ledger{
				Entries: []models.LedgerEntry{entry},
			}),
		})
		defer t.Cleanup(func() {
			http.DefaultClient.CloseIdleConnections()
		})

		res, err := http.Get(url + "/ledger")
		if err != nil {
			t.Errorf("Error while making request: %v", err)
		}
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, res.StatusCode)
		}
	})
}

func TestWithEnv(t *testing.T) {
	t.Run("should set and restore env variables", func(t *testing.T) {
		WithEnv(map[string]string{"UAG_TEST": "1"}, func() {
			if os.Getenv("UAG_TEST") != "1" {
				t.Errorf("Expected UAG_TEST to be 1, got %s", os.Getenv("UAG_TEST"))
			}
		})
		if os.Getenv("UAG_TEST") != "" {
			t.Errorf("Expected UAG_TEST to be empty, got %s", os.Getenv("UAG_TEST"))
		}
	})
}
