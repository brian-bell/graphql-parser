package ast

import "testing"

func mkSource(body string) *Source {
	return &Source{Body: body, Name: "test.graphql"}
}

func TestPositionAt_Empty(t *testing.T) {
	s := mkSource("")
	if got := s.PositionAt(0); got != (Position{Line: 1, Column: 1}) {
		t.Errorf("PositionAt(0) of empty source = %+v; want {1, 1}", got)
	}
	if got := s.PositionAt(5); got != (Position{Line: 1, Column: 1}) {
		t.Errorf("PositionAt(5) of empty source = %+v; want {1, 1} (clamped)", got)
	}
}

func TestPositionAt_SingleLine(t *testing.T) {
	s := mkSource("hello")
	cases := []struct {
		offset int
		want   Position
	}{
		{0, Position{1, 1}},
		{1, Position{1, 2}},
		{4, Position{1, 5}},
		{5, Position{1, 6}}, // EOF
	}
	for _, c := range cases {
		if got := s.PositionAt(c.offset); got != c.want {
			t.Errorf("PositionAt(%d) = %+v; want %+v", c.offset, got, c.want)
		}
	}
}

func TestPositionAt_LF(t *testing.T) {
	s := mkSource("a\nb")
	cases := []struct {
		offset int
		want   Position
	}{
		{0, Position{1, 1}},
		{1, Position{1, 2}}, // \n
		{2, Position{2, 1}}, // b
		{3, Position{2, 2}}, // EOF
	}
	for _, c := range cases {
		if got := s.PositionAt(c.offset); got != c.want {
			t.Errorf("PositionAt(%d) = %+v; want %+v", c.offset, got, c.want)
		}
	}
}

func TestPositionAt_CRLF(t *testing.T) {
	s := mkSource("a\r\nb")
	// "a", "\r", "\n", "b" at offsets 0, 1, 2, 3.
	// Line 2 starts at offset 3 (after the \r\n).
	cases := []struct {
		offset int
		want   Position
	}{
		{0, Position{1, 1}}, // a
		{1, Position{1, 2}}, // \r — still on line 1
		{2, Position{1, 3}}, // \n inside the CRLF; binary-search reports the start-line
		{3, Position{2, 1}}, // b
		{4, Position{2, 2}}, // EOF
	}
	for _, c := range cases {
		if got := s.PositionAt(c.offset); got != c.want {
			t.Errorf("PositionAt(%d) = %+v; want %+v", c.offset, got, c.want)
		}
	}
}

func TestPositionAt_LoneCR(t *testing.T) {
	s := mkSource("a\rb")
	cases := []struct {
		offset int
		want   Position
	}{
		{0, Position{1, 1}}, // a
		{1, Position{1, 2}}, // \r
		{2, Position{2, 1}}, // b
		{3, Position{2, 2}}, // EOF
	}
	for _, c := range cases {
		if got := s.PositionAt(c.offset); got != c.want {
			t.Errorf("PositionAt(%d) = %+v; want %+v", c.offset, got, c.want)
		}
	}
}

func TestPositionAt_NoTrailingNewline(t *testing.T) {
	s := mkSource("foo\nbar")
	if got := s.PositionAt(7); got != (Position{2, 4}) {
		t.Errorf("PositionAt(7) of %q = %+v; want {2, 4}", s.Body, got)
	}
}

func TestPositionAt_TrailingNewline(t *testing.T) {
	s := mkSource("foo\n")
	if got := s.PositionAt(4); got != (Position{2, 1}) {
		t.Errorf("PositionAt(4) of %q = %+v; want {2, 1}", s.Body, got)
	}
}

func TestPositionAt_MixedTerminators(t *testing.T) {
	// "a\nb\r\nc\rd" — four lines: a, b, c, d
	// Offsets:        0 1 2 3 4 5 6 7
	s := mkSource("a\nb\r\nc\rd")
	cases := []struct {
		offset int
		want   Position
	}{
		{0, Position{1, 1}}, // a
		{2, Position{2, 1}}, // b
		{5, Position{3, 1}}, // c
		{7, Position{4, 1}}, // d
		{8, Position{4, 2}}, // EOF
	}
	for _, c := range cases {
		if got := s.PositionAt(c.offset); got != c.want {
			t.Errorf("PositionAt(%d) = %+v; want %+v", c.offset, got, c.want)
		}
	}
}

func TestPositionAt_MultiByteRunes(t *testing.T) {
	// "héllo" — é is 2 bytes (UTF-8 0xC3 0xA9). Column counts codepoints, not bytes.
	s := mkSource("héllo")
	cases := []struct {
		offset int
		want   Position
	}{
		{0, Position{1, 1}}, // h
		{1, Position{1, 2}}, // é (start byte)
		{3, Position{1, 3}}, // l (after é)
		{4, Position{1, 4}}, // l
		{5, Position{1, 5}}, // o
		{6, Position{1, 6}}, // EOF
	}
	for _, c := range cases {
		if got := s.PositionAt(c.offset); got != c.want {
			t.Errorf("PositionAt(%d) = %+v; want %+v", c.offset, got, c.want)
		}
	}
}

func TestPositionAt_NegativeOffset(t *testing.T) {
	s := mkSource("hi")
	if got := s.PositionAt(-1); got != (Position{1, 1}) {
		t.Errorf("PositionAt(-1) = %+v; want {1, 1}", got)
	}
}

func TestPositionAt_LocationOffset_Default(t *testing.T) {
	// Zero-value LocationOffset is treated as {1,1} (no offset).
	s := &Source{Body: "x"}
	if got := s.PositionAt(0); got != (Position{1, 1}) {
		t.Errorf("PositionAt(0) with zero LocationOffset = %+v; want {1, 1}", got)
	}
}

func TestPositionAt_LocationOffset_NonDefault(t *testing.T) {
	s := &Source{
		Body:           "abc\ndef",
		LocationOffset: Position{Line: 10, Column: 5},
	}
	// First line: line + 10 - 1; column + 5 - 1.
	if got := s.PositionAt(0); got != (Position{10, 5}) {
		t.Errorf("PositionAt(0) = %+v; want {10, 5}", got)
	}
	if got := s.PositionAt(1); got != (Position{10, 6}) {
		t.Errorf("PositionAt(1) = %+v; want {10, 6}", got)
	}
	// Second line: column does NOT inherit the Column offset; only Line is shifted.
	if got := s.PositionAt(4); got != (Position{11, 1}) {
		t.Errorf("PositionAt(4) = %+v; want {11, 1}", got)
	}
}

func TestLoc_StartEndPosition(t *testing.T) {
	s := mkSource("foo\nbarbaz")
	// Indices: f=0 o=1 o=2 \n=3 b=4 a=5 r=6 b=7 a=8 z=9
	loc := &Loc{Start: 4, End: 10, Source: s}
	if got := loc.StartPosition(); got != (Position{2, 1}) {
		t.Errorf("StartPosition() = %+v; want {2, 1}", got)
	}
	if got := loc.EndPosition(); got != (Position{2, 7}) {
		t.Errorf("EndPosition() = %+v; want {2, 7}", got)
	}
}

func TestPosition_String(t *testing.T) {
	cases := []struct {
		p    Position
		want string
	}{
		{Position{1, 1}, "1:1"},
		{Position{12, 34}, "12:34"},
		{Position{}, "0:0"},
	}
	for _, c := range cases {
		if got := c.p.String(); got != c.want {
			t.Errorf("Position%+v.String() = %q; want %q", c.p, got, c.want)
		}
	}
}

func TestPositionAt_ConcurrentSafe(t *testing.T) {
	// Repeated calls on the same Source must produce the same result and
	// not race on lazy line-start computation.
	s := mkSource("a\nb\r\nc\rd")
	want := Position{3, 1}
	done := make(chan struct{}, 16)
	for range 16 {
		go func() {
			for range 1000 {
				if got := s.PositionAt(5); got != want {
					t.Errorf("PositionAt(5) = %+v; want %+v", got, want)
				}
			}
			done <- struct{}{}
		}()
	}
	for range 16 {
		<-done
	}
}
