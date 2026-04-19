package adapters

import "testing"

func TestLoad_NonEmpty(t *testing.T) {
	adapters, err := Load()
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

func TestLoad_Sorted(t *testing.T) {
	adapters, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(adapters); i++ {
		if adapters[i-1].Name > adapters[i].Name {
			t.Errorf("adapters not sorted: %q > %q", adapters[i-1].Name, adapters[i].Name)
		}
	}
}
