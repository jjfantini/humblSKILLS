package adapters

import "testing"

func TestNameSet(t *testing.T) {
	set := NameSet([]Adapter{{Name: "a"}, {Name: "b"}})
	if _, ok := set["a"]; !ok {
		t.Error("missing a")
	}
	if _, ok := set["b"]; !ok {
		t.Error("missing b")
	}
	if _, ok := set["c"]; ok {
		t.Error("unexpected c")
	}
}
