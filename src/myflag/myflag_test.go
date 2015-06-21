package myflag

import (
	"testing"
)

func TestHelpMessage(t *testing.T) {
	Reset()

	Int("number", "n", "integer", "yoink yoink", 5)
	Float("float", "f", "", "num num", 5)
	String("something", "", "path", "cool", "default")
	Bool("boolie", "b", "booboo", false)
	
	help := HelpString()
	expected :=
`	-n <integer>
	-number <integer>
		yoink yoink

	-f
	-float
		num num

	-something <path>
		cool

	-b
	-boolie
		booboo

`
	if help != expected {
		t.Errorf("Bad help output. Expected:\n%sActual:\n%s", expected, help)
	}
}

func TestParsing(t *testing.T) {
	Reset()
	
	i := Int("number", "n", "", "", 1)
	f := Float("float", "f", "", "", 0.1)
	s := String("something", "", "", "", "default")
	b := Bool("boolie", "b", "", false)
	
	// Check defaults.
	if *i != 1 {
		t.Fatalf("i=%v instead of 1", *i)
	}
	if *f != 0.1 {
		t.Fatalf("f=%v instead of 0.1", *f)
	}
	if *s != "default" {
		t.Fatalf("s=%v instead of 'default'", *s)
	}
	if *b != false {
		t.Fatalf("b=%v instead of false", *b)
	}
	
	// Parse with short names.
	ParseStrings([]string{ "-n", "2", "-f", "0.2", "-b" })
	
	if *i != 2 {
		t.Fatalf("i=%v instead of 2", *i)
	}
	if *f != 0.2 {
		t.Fatalf("f=%v instead of 0.2", *f)
	}
	if *s != "default" {
		t.Fatalf("s=%v instead of 'default'", *s)
	}
	if *b != true {
		t.Fatalf("b=%v instead of true", *b)
	}
	
	// Parse with long names.
	ParseStrings([]string{ "-number", "3", "-float", "0.3", "-boolie",
			"-something", "howdy" })
	
	if *i != 3 {
		t.Fatalf("i=%v instead of 3", *i)
	}
	if *f != 0.3 {
		t.Fatalf("f=%v instead of 0.3", *f)
	}
	if *s != "howdy" {
		t.Fatalf("s=%v instead of 'howdy'", *s)
	}
	if *b != true {
		t.Fatalf("b=%v instead of true", *b)
	}
}

