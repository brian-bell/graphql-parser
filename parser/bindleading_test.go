package parser

import (
	"testing"

	"github.com/brian-bell/graphql-parser/ast"
)

// TestBindLeading_PermissiveOnLatentNodes documents that bindLeading is
// intentionally permissive: now that all 39 Comments-bearing nodes implement
// ast.CommentedNode, the interface assertion succeeds even for the latent
// value nodes the parser never binds today. Behavior is preserved because the
// four call sites only pass nodes that were always bound — scope is controlled
// by *where* bindLeading is called, not by the helper rejecting node types.
func TestBindLeading_PermissiveOnLatentNodes(t *testing.T) {
	p := &parser{cfg: &config{preserveComments: true}}
	iv := &ast.IntValue{Value: "1"}
	leading := []*ast.Comment{{Text: " a comment"}}

	p.bindLeading(iv, leading)

	if iv.Comments == nil || len(iv.Comments.Leading) != 1 || iv.Comments.Leading[0].Text != " a comment" {
		t.Errorf("bindLeading should populate even a latent node's Comments; got %+v", iv.Comments)
	}
}

// TestBindLeading_NoOpWhenPreservationOff pins that bindLeading does nothing
// when comment preservation is off, even though takeLeading still runs.
func TestBindLeading_NoOpWhenPreservationOff(t *testing.T) {
	p := &parser{cfg: &config{preserveComments: false}}
	fd := &ast.FieldDefinition{Name: "x"}
	p.bindLeading(fd, []*ast.Comment{{Text: " ignored"}})
	if fd.Comments != nil {
		t.Errorf("bindLeading should be a no-op when preservation is off; got %+v", fd.Comments)
	}
}
