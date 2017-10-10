package scrapers

import (
	"testing"
)

func TestIsStoresFile(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"Stores1", true},
		{"Stores", false},
		{"Prices1", false},
		{"/Stores1", true},
	}
	for i, test := range tests {
		if got := isStoresFile(test.s); got != test.want {
			t.Errorf("#%v: isStoresFile(%q)=%v want %v",
				i+1, test.s, got, test.want)
		}
	}
}
