package accountstore_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/albertoZurini/vcx/pkg/accountstore"
)

func compactJSON(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		t.Fatalf("compactJSON: %v", err)
	}
	return buf.String()
}

func setup(t *testing.T) {
	t.Helper()
	t.Setenv("ACCOUNTSTORE_DIR", t.TempDir())
}

func TestLoad_MissingFile(t *testing.T) {
	setup(t)
	accounts, err := accountstore.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(accounts) != 0 {
		t.Fatalf("expected empty slice, got %d accounts", len(accounts))
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	setup(t)
	input := []accountstore.Account{
		{Name: "alice (alice@example.com)", Data: json.RawMessage(`{"token":"tok_alice"}`)},
		{Name: "bob (bob@example.com)", Data: json.RawMessage(`{"token":"tok_bob"}`)},
	}
	if err := accountstore.Save(input); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := accountstore.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != len(input) {
		t.Fatalf("expected %d accounts, got %d", len(input), len(got))
	}
	for i := range input {
		if got[i].Name != input[i].Name {
			t.Errorf("[%d] Name: got %q, want %q", i, got[i].Name, input[i].Name)
		}
		if compactJSON(t, got[i].Data) != compactJSON(t, input[i].Data) {
			t.Errorf("[%d] Data: got %s, want %s", i, got[i].Data, input[i].Data)
		}
	}
}

func TestAddOrUpdate_Insert(t *testing.T) {
	setup(t)
	if err := accountstore.AddOrUpdate("alice (alice@example.com)", json.RawMessage(`{"token":"tok_a"}`)); err != nil {
		t.Fatal(err)
	}
	accounts, _ := accountstore.Load()
	if len(accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(accounts))
	}
	if accounts[0].Name != "alice (alice@example.com)" {
		t.Errorf("unexpected name: %s", accounts[0].Name)
	}
}

func TestAddOrUpdate_UpdateExisting(t *testing.T) {
	setup(t)
	accountstore.AddOrUpdate("alice (alice@example.com)", json.RawMessage(`{"token":"old"}`))
	accountstore.AddOrUpdate("alice (alice@example.com)", json.RawMessage(`{"token":"new"}`))

	accounts, _ := accountstore.Load()
	if len(accounts) != 1 {
		t.Fatalf("expected 1 account after update, got %d", len(accounts))
	}
	if compactJSON(t, accounts[0].Data) != `{"token":"new"}` {
		t.Errorf("expected updated token, got %s", accounts[0].Data)
	}
}

func TestAddOrUpdate_MultipleAccounts(t *testing.T) {
	setup(t)
	accountstore.AddOrUpdate("alice (alice@example.com)", json.RawMessage(`{"token":"tok_a"}`))
	accountstore.AddOrUpdate("bob (bob@example.com)", json.RawMessage(`{"token":"tok_b"}`))

	accounts, _ := accountstore.Load()
	if len(accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(accounts))
	}
}

func TestAll_AliasForLoad(t *testing.T) {
	setup(t)
	accountstore.AddOrUpdate("alice (alice@example.com)", json.RawMessage(`{"token":"tok_a"}`))

	fromLoad, _ := accountstore.Load()
	fromAll, _ := accountstore.All()

	if len(fromLoad) != len(fromAll) {
		t.Fatalf("All() and Load() disagree: %d vs %d", len(fromAll), len(fromLoad))
	}
	if fromLoad[0].Name != fromAll[0].Name {
		t.Errorf("All() name mismatch: %q vs %q", fromAll[0].Name, fromLoad[0].Name)
	}
}

func TestClear(t *testing.T) {
	setup(t)
	accountstore.AddOrUpdate("alice (alice@example.com)", json.RawMessage(`{"token":"tok_a"}`))
	if err := accountstore.Clear(); err != nil {
		t.Fatal(err)
	}
	accounts, _ := accountstore.Load()
	if len(accounts) != 0 {
		t.Fatalf("expected empty after Clear, got %d", len(accounts))
	}
}

func TestSave_FilePermissions(t *testing.T) {
	setup(t)
	if err := accountstore.Save([]accountstore.Account{}); err != nil {
		t.Fatal(err)
	}
	path, _ := accountstore.FilePath()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected permissions 0600, got %04o", perm)
	}
}
