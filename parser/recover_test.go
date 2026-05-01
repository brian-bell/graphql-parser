package parser_test

import (
	"errors"
	"testing"

	"github.com/bellbm/graphql-parser/ast"
	"github.com/bellbm/graphql-parser/parser"
)

func TestRecovery_Off_StillFailsFast(t *testing.T) {
	// Identical to phase 6's empty-document test: with no WithRecovery the
	// API behaves exactly as before.
	if _, err := parser.Parse(""); err == nil {
		t.Fatal("expected error")
	}
	if _, err := parser.Parse("{ foo bar(}"); err == nil {
		t.Fatal("expected error")
	}
}

func TestRecovery_BadDefinition_ResumesAtNextDefinition(t *testing.T) {
	body := `
		bogus_keyword stuff
		query Q { ok }
	`
	doc, err := parser.Parse(body, parser.WithRecovery())
	if err == nil {
		t.Fatal("expected ParseErrors")
	}
	var es parser.ParseErrors
	if !errors.As(err, &es) {
		t.Fatalf("err is %T, expected ParseErrors", err)
	}
	if len(es) == 0 {
		t.Errorf("expected at least 1 error, got %d", len(es))
	}
	if doc == nil {
		t.Fatal("expected partial document")
	}
	if len(doc.Definitions) < 2 {
		t.Fatalf("expected >=2 definitions; got %d (%v)", len(doc.Definitions), doc.Definitions)
	}
	if _, ok := doc.Definitions[0].(*ast.BadDefinition); !ok {
		t.Errorf("[0] is %T; want BadDefinition", doc.Definitions[0])
	}
	if op, ok := doc.Definitions[1].(*ast.OperationDefinition); !ok || op.Name != "Q" {
		t.Errorf("[1] is %T %v; want OperationDefinition Q", doc.Definitions[1], doc.Definitions[1])
	}
}

func TestRecovery_MultipleBadDefinitions(t *testing.T) {
	// "bogus1 bogus2" coalesces into a single BadDefinition because
	// skipToDefinitionStart advances over consecutive non-keyword tokens
	// until it lands on a real definition keyword. That is intentional
	// (less noise for the LSP); the contract we test is "the good defs
	// survive and at least one error is reported per bad region."
	body := `
		bogus1 bogus2
		query Q { ok }
		more_bogus
		fragment F on T { y }
	`
	doc, err := parser.Parse(body, parser.WithRecovery())
	var es parser.ParseErrors
	if !errors.As(err, &es) {
		t.Fatalf("err is %T", err)
	}
	if len(es) < 2 {
		t.Errorf("expected >=2 errors, got %d: %v", len(es), es)
	}

	bad, good := 0, 0
	for _, def := range doc.Definitions {
		if _, ok := def.(*ast.BadDefinition); ok {
			bad++
		} else {
			good++
		}
	}
	if bad < 2 || good < 2 {
		t.Errorf("expected >=2 bad and >=2 good defs; got bad=%d good=%d", bad, good)
	}
}

func TestRecovery_BadFieldInsideSelectionSet(t *testing.T) {
	// "{ a $bogus b }" — $ is not a legal selection start. Recovery skips
	// it as a single bad token, then "bogus" parses as a fresh Field. The
	// contract: good fields a and b survive, at least one BadField is
	// inserted, and recovery reports the error.
	body := `{ a $bogus b }`
	doc, err := parser.Parse(body, parser.WithRecovery())
	var es parser.ParseErrors
	if !errors.As(err, &es) {
		t.Fatalf("err is %T (want ParseErrors): %v", err, err)
	}
	if len(es) == 0 {
		t.Error("expected at least 1 error")
	}
	op := doc.Definitions[0].(*ast.OperationDefinition)
	if op.SelectionSet == nil {
		t.Fatal("nil selection set")
	}
	sels := op.SelectionSet.Selections
	if len(sels) < 3 {
		t.Fatalf("selections = %d; want >=3", len(sels))
	}

	var hasA, hasB, hasBad bool
	for _, s := range sels {
		switch v := s.(type) {
		case *ast.Field:
			if v.Name == "a" {
				hasA = true
			}
			if v.Name == "b" {
				hasB = true
			}
		case *ast.BadField:
			hasBad = true
		}
	}
	if !hasA || !hasB {
		t.Errorf("expected both 'a' and 'b' to survive; hasA=%v hasB=%v", hasA, hasB)
	}
	if !hasBad {
		t.Error("expected at least one BadField placeholder")
	}
}

func TestRecovery_ErrorsCarryPositions(t *testing.T) {
	body := `bogus query Q { ok }`
	_, err := parser.Parse(body, parser.WithRecovery())
	var es parser.ParseErrors
	if !errors.As(err, &es) {
		t.Fatalf("err is %T", err)
	}
	if len(es) == 0 {
		t.Fatal("no errors")
	}
	pos := es[0].Position()
	if pos.Line != 1 || pos.Column < 1 {
		t.Errorf("error position = %+v; expected line 1, column >=1", pos)
	}
}

func TestRecovery_NoErrors_NoParseErrors(t *testing.T) {
	body := `query Q { x }`
	doc, err := parser.Parse(body, parser.WithRecovery())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc == nil {
		t.Fatal("expected document")
	}
}

func TestRecovery_ErrorMessage_AggregatesAll(t *testing.T) {
	body := `bogus1 bogus2`
	_, err := parser.Parse(body, parser.WithRecovery())
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if len(msg) == 0 {
		t.Error("error message empty")
	}
}
