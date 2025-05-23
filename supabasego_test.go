package supabasego

import (
	"os"
	"testing"
	"time"
)

type TestTenant struct {
	ID           string     `json:"id,omitempty"`
	UserID       string     `json:"user_id"`
	Name         string     `json:"name"`
	Slug         string     `json:"slug"`
	ContactEmail string     `json:"contact_email"`
	Plan         string     `json:"plan"`
	MaxUsers     int        `json:"max_users"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

func getTestClient() *Client {
	return NewClient(Config{
		BaseURL: os.Getenv("DB_URL"),
		APIKey:  os.Getenv("DB_API_KEY"),
	})
}

func TestTableCRUD(t *testing.T) {
	client := getTestClient()
	table := client.Table("test_tenants")
	userID := "test-user-123"

	tenant := TestTenant{
		UserID:       userID,
		Name:         "Test Tenant",
		Slug:         "test-tenant",
		ContactEmail: "test@example.com",
		Plan:         "free",
		MaxUsers:     10,
	}
	// --- Insert ---
	err := table.Insert(tenant, "")
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	// --- Select ---
	var tenants []TestTenant
	err = table.Eq("user_id", userID).Select(&tenants, "")
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}
	if len(tenants) == 0 {
		t.Fatalf("No tenant found after insert")
	}
	// --- Update ---
	update := map[string]interface{}{"plan": "pro"}
	err = table.Eq("user_id", userID).Update(update, "")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	// --- Select after update ---
	var updated []TestTenant
	err = table.Eq("user_id", userID).Select(&updated, "")
	if err != nil {
		t.Fatalf("Select after update failed: %v", err)
	}
	if len(updated) == 0 || updated[0].Plan != "pro" {
		t.Fatalf("Update not reflected in select")
	}
	// --- Delete ---
	err = table.Eq("user_id", userID).Delete("")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	// --- Select after delete (should be gone) ---
	var afterDelete []TestTenant
	err = table.Eq("user_id", userID).Select(&afterDelete, "")
	if err != nil {
		t.Fatalf("Select after delete failed: %v", err)
	}
	if len(afterDelete) > 0 {
		t.Fatalf("Tenant still present after delete")
	}
}

func TestTableCRUD_Scaffold(t *testing.T) {
	// Scaffold test for Table CRUD methods.
	// Real tests to be added as implementation progresses.
}
