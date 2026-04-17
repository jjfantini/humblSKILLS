package platform

import "testing"

func TestLoadBuiltin_NonEmpty(t *testing.T) {
	adapters, err := LoadBuiltin()
	if err != nil {
		t.Fatal(err)
	}
	if len(adapters) == 0 {
		t.Fatal("expected at least one embedded adapter")
	}

	names := map[string]struct{}{}
	for _, a := range adapters {
		names[a.Name] = struct{}{}
	}
	for _, expect := range []string{"claude-code", "cursor"} {
		if _, ok := names[expect]; !ok {
			t.Errorf("expected embedded adapter %q, got %v", expect, names)
		}
	}
}
