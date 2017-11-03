package main

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	for i, test := range tests {
		got, err := parseSchema([]byte(test.input))
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
			Doc: nil,
			Tables: []*table{
				{
					Name: []byte("amit"),
					Doc:  []byte("Hello"),
					Fields: []*field{
						{[]byte("a"), []byte("integer"), []byte("World"), false},
						{[]byte("bb"), []byte("text"), []byte("How are you?"), true},
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
			Doc: []byte("Bli Bla Blu"),
			Tables: []*table{
				{
					Name: []byte("amit"),
					Doc:  []byte("Hello"),
					Fields: []*field{
						{[]byte("a"), []byte("integer"), []byte("World"), false},
						{[]byte("bb"), []byte("text"), []byte("How are you?"), true},
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
			Doc: nil,
			Tables: []*table{
				{
					Name: []byte("stores"),
					Doc:  []byte("Identifies every store in the data."),
					Fields: []*field{
						{[]byte("store_id"), []byte("integer"), []byte(""), true},
						{[]byte("chain_id"), []byte("text"), []byte("Chain code, as provided by GS1."), false},
						{[]byte("subchain_id"), []byte("text"), []byte("Subchain number."), false},
						{[]byte("reported_store_id"), []byte("text"), []byte("Store number issued by the chain."), false},
					},
				},
			},
		},
	},
}
