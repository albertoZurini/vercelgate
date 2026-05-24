package jsonupdate_test

import (
	"strings"
	"testing"

	"github.com/albertoZurini/vcx/pkg/jsonupdate"
)

func TestSet_ExistingKey(t *testing.T) {
	u := jsonupdate.NewJsonUpdate(`{"token":"old"}`)
	u.Set("token", "new")
	if u.String() != `{"token":"new"}` {
		t.Errorf("got %s", u.String())
	}
}

func TestSet_NewKey(t *testing.T) {
	u := jsonupdate.NewJsonUpdate(`{"token":"x"}`)
	u.Set("currentTeam", "t_123")
	got := u.String()
	if !strings.Contains(got, `"currentTeam"`) {
		t.Errorf("new key not present in: %s", got)
	}
	if !strings.Contains(got, `"t_123"`) {
		t.Errorf("new value not present in: %s", got)
	}
}

func TestSet_Chaining(t *testing.T) {
	u := jsonupdate.NewJsonUpdate(`{}`)
	u.Set("a", "1").Set("b", "2")
	got := u.String()
	if !strings.Contains(got, `"a"`) || !strings.Contains(got, `"b"`) {
		t.Errorf("chained sets not present in: %s", got)
	}
}

func TestDelete_RemovesKey(t *testing.T) {
	u := jsonupdate.NewJsonUpdate(`{"token":"x","currentTeam":"t_123"}`)
	u.Delete("currentTeam")
	got := u.String()
	if strings.Contains(got, "currentTeam") {
		t.Errorf("key should be removed, got: %s", got)
	}
	if !strings.Contains(got, `"token"`) {
		t.Errorf("unrelated key should remain, got: %s", got)
	}
}

func TestDelete_MissingKey(t *testing.T) {
	u := jsonupdate.NewJsonUpdate(`{"token":"x"}`)
	u.Delete("nonexistent")
	if u.String() != `{"token":"x"}` {
		t.Errorf("deleting missing key should be no-op, got: %s", u.String())
	}
}

func TestPretty_IsIndented(t *testing.T) {
	u := jsonupdate.NewJsonUpdate(`{"token":"x","foo":"bar"}`)
	pretty := u.Pretty()
	if !strings.Contains(pretty, "\n") {
		t.Error("Pretty() output should be multi-line")
	}
	if !strings.Contains(pretty, "  ") {
		t.Error("Pretty() output should be indented")
	}
}

func TestString_IsMinified(t *testing.T) {
	u := jsonupdate.NewJsonUpdate(`{"token":"x"}`)
	got := u.String()
	if strings.Contains(got, "\n") {
		t.Error("String() should return compact JSON")
	}
}
