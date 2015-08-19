package myxml

import (
	"testing"
	"strings"
	"regexp"
)

// TODO(amit): Use assert package.

func TestFindTag_simple(t *testing.T) {
	text := "<amit><lavon>bla</lavon></amit>"
	node, err := Read(strings.NewReader(text))
	if err != nil {
		t.Fatal(err)
	}
	find := node.FindTag(regexp.MustCompile("lavon"))
	if len(find) != 1 {
		t.Fatalf("Bad find length: %d, expected 1.", len(find))
	}
	if len(find[0].Children) != 1 {
		t.Fatalf("Bad number of children: %d, expected 1.",
				len(find[0].Children))
	}
	if find[0].Children[0].Text != "bla" {
		t.Fatalf("Bad text: '%s', expected 'bla'.", find[0].Children[0].Text)
	}
}

