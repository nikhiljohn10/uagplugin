package main

import (
	"net/http"
	"testing"

	"github.com/nikhiljohn10/uagplugin/models"
	tk "github.com/nikhiljohn10/uagplugin/testkit"
)

func TestHealth(t *testing.T) {
	if Health() != "ok" {
		t.Fatalf("health not ok")
	}
}

func TestContactsWithMock(t *testing.T) {
	server, url := tk.StartMockServer(map[string]http.Handler{
		"/users": tk.JSONResponse(200, []map[string]any{
			{"id": 1, "name": "Alice Wonderland", "email": "alice@example.com"},
			{"id": 2, "name": "Bob Builder", "email": "bob@example.com"},
		}),
	})
	defer server.Close()

	tk.WithEnv(tk.TestVars{"API_BASE_URL": url}, func() {
		out, err := Contacts(nil, models.ContactQueryParams{Search: "alice", CommonParams: models.CommonParams{SortDescending: true}})
		if err != nil {
			t.Fatalf("Contacts error: %v", err)
		}
		if out == nil || out.Count != 1 {
			t.Fatalf("expected 1 contact, got %+v", out)
		}
		if out.Items[0].Name != "Alice Wonderland" {
			t.Fatalf("unexpected contact: %+v", out.Items[0])
		}
	})
}
