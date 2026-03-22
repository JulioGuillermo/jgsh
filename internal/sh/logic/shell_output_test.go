package logic

import (
	"testing"
)

func TestFoldCarriageReturnsWithBackspace(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"Hello\rWorld", "World"},
		{"Loading [|]  \b\b\b\b\b\b[/]  ", "Loading [/]  "},
		{"abc\b\bde", "ade"},
		{"First line\nSecond\rOverwritten", "First line\nOverwritten"},
	}

	for _, c := range cases {
		result := FoldCarriageReturns(c.input)
		if result != c.expected {
			t.Errorf("For %q, expected %q, got %q", c.input, c.expected, result)
		}
	}
}
