package cmd

import (
	"testing"
)

func Test_TableFormatter(t *testing.T) {
	tests := []struct {
		name   string
		object interface{}
		expect string
	}{
		{
			name:   "nil object",
			object: nil,
			expect: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TableFormatter(tt.object)
			if got != tt.expect {
				t.Errorf("Expected:\n%s\nGot:\n%s\n", tt.expect, got)
			}
		})
	}
}
