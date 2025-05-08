//go:build linux

package internal

import "testing"

func TestIsLinuxModelPlaceholder(t *testing.T) {
	cases := map[string]bool{
		"ThinkPad T14":             false,
		"To Be Filled By O.E.M.":   true,
		"System Product Name":      true,
		"Unknown":                  true,
		"":                         false,
	}
	for in, want := range cases {
		if got := isLinuxModelPlaceholder(in); got != want {
			t.Fatalf("%q: got %v want %v", in, got, want)
		}
	}
}
