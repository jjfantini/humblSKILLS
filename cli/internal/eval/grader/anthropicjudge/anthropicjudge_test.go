package anthropicjudge

import "testing"

func TestExtractJSONObject(t *testing.T) {
	tests := []struct{ in, want string }{
		{`{"a":1}`, `{"a":1}`},
		{"```json\n{\"a\":1}\n```", `{"a":1}`},
		{"some preamble {\"a\":1} trailing", `{"a":1}`},
		{"{\"outer\":{\"inner\":1}}", `{"outer":{"inner":1}}`},
	}
	for _, tc := range tests {
		if got := extractJSONObject(tc.in); got != tc.want {
			t.Errorf("extractJSONObject(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
