package jsonparse

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestFetchers(t *testing.T) {
	tests := []struct {
		json     string
		key      string
		expected interface{}
	}{
		{
			json:     `{"ref":"ok"}`,
			key:      "ref",
			expected: "ok",
		},
		{
			json:     `[7,8,9,0]`,
			key:      "2",
			expected: float64(9),
		},
		{
			json:     `["what", "is", "this"]`,
			key:      "2",
			expected: "this",
		},
		{
			json:     `{"ref": [5,8,9]}`,
			key:      "ref.1",
			expected: float64(8),
		},
		{
			json:     `{"ref": {"joe": [1,2, {"sum": 100 } ]}}`,
			key:      "ref.joe.2.sum",
			expected: float64(100),
		},
		{
			json:     `{"ref": {"joe": [1,2, {"sum": {"100": {"dave" : "lee"}}  } ]}}`,
			key:      "ref.joe.2.sum.100.dave",
			expected: "lee",
		},
	}

	for i, tt := range tests {
		var v interface{}
		err := json.Unmarshal([]byte(tt.json), &v)
		if err != nil {
			t.Fatal(err)
			continue
		}
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			val := fetchValue(v, tt.key)
			if val != tt.expected {
				t.Errorf("want: %v, got: %v", tt.expected, val)
			}
		})
	}

}
