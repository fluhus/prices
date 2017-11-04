package main

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	for i, test := range tests {
		got, err := parseSchema(test.input)
		if err != nil {
			t.Fatalf("#%v: failed to parse test schema: %v", i+1, err)
		}
		if !reflect.DeepEqual(test.expected, got) {
			t.Errorf("#%v: parseSchema(...)=\n%v\nwant\n%v", i+1, got, test.expected)
		}
	}
}

var tests = []struct {
	input    string
	expected *schema
}{
	{
		`CREATE TABLE amit (
-- Hello
	a  integer -- World
	bb text    -- How are you? (safe)
);
`,
		&schema{
			Doc: "",
			Tables: []*table{
				{
					Name: "amit",
					Doc:  "Hello",
					Fields: []*field{
						{"a", "integer", "World", false},
						{"bb", "text", "How are you?", true},
					},
				},
			},
		},
	},
	{
		`CREATE TABLE documentation (
-- Bli
-- Bla
-- Blu

a
);

some stuffs

CREATE TABLE amit (
-- Hello
	a  integer -- World
	bb text    -- How are you? (safe)
);
`,
		&schema{
			Doc: "Bli Bla Blu",
			Tables: []*table{
				{
					Name: "amit",
					Doc:  "Hello",
					Fields: []*field{
						{"a", "integer", "World", false},
						{"bb", "text", "How are you?", true},
					},
				},
			},
		},
	},
	{
		`CREATE TABLE stores (
-- Identifies every store in the data.
	store_id           integer, -- (safe)
	chain_id           text  NOT NULL, -- Chain code, as provided by GS1.
	subchain_id        text  NOT NULL, -- Subchain number.
	reported_store_id  text  NOT NULL  -- Store number issued by the chain.
);
`,
		&schema{
			Doc: "",
			Tables: []*table{
				{
					Name: "stores",
					Doc:  "Identifies every store in the data.",
					Fields: []*field{
						{"store_id", "integer", "", true},
						{"chain_id", "text", "Chain code, as provided by GS1.", false},
						{"subchain_id", "text", "Subchain number.", false},
						{"reported_store_id", "text", "Store number issued by the chain.", false},
					},
				},
			},
		},
	},
}
