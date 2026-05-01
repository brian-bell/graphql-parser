package ast

import "testing"

func TestBlockStringValue(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"empty", "", ""},
		{"single line", "hello", "hello"},
		{"first line indent preserved", "  hello", "  hello"},
		{
			"strips common indent of non-first lines",
			"\n    Hello,\n      World!\n\n    Yours,\n      GraphQL.\n  ",
			"Hello,\n  World!\n\nYours,\n  GraphQL.",
		},
		{
			"removes blank leading and trailing lines",
			"\n\n  Hello,\n    World!\n\n\n",
			"Hello,\n  World!",
		},
		{
			"keeps blank lines in the middle",
			"\n  Hello,\n\n  World!\n",
			"Hello,\n\nWorld!",
		},
		{
			"first-line indent excluded from common-indent calculation",
			"      Hello,\n    World!",
			"      Hello,\nWorld!",
		},
		{
			"lone CR is a line terminator",
			"\r  Hello,\r    World!\r",
			"Hello,\n  World!",
		},
		{
			"CRLF normalized to LF",
			"\r\n  Hello,\r\n    World!\r\n",
			"Hello,\n  World!",
		},
		{
			"mixed terminators",
			"\n  a\r\n  b\r  c\n",
			"a\nb\nc",
		},
		{
			"tabs count as one indent character each",
			"\n\thello\n\tworld",
			"hello\nworld",
		},
		{
			"only whitespace lines are blank",
			"\n   \n  hello\n   \n",
			"hello",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := BlockStringValue(c.raw)
			if got != c.want {
				t.Errorf("BlockStringValue(%q):\ngot  %q\nwant %q", c.raw, got, c.want)
			}
		})
	}
}

func TestBlockStringValue_FixedPoints(t *testing.T) {
	// Inputs that are already in their canonical block-string-value form
	// should pass through unchanged. Note that the spec algorithm is NOT
	// idempotent in general — re-running on a result like "a\n  b" would
	// strip the 2-space continuation indent — so we only test inputs whose
	// non-first lines have no leading whitespace.
	inputs := []string{
		"",
		"hello",
		"a\nb\nc",
		"first line indented\nsecond not",
	}
	for _, in := range inputs {
		got := BlockStringValue(in)
		if got != in {
			t.Errorf("BlockStringValue(%q) = %q; want unchanged", in, got)
		}
	}
}
