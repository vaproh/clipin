package config

import "testing"

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "valid", raw: "15m", want: true},
		{name: "zero", raw: "0s", want: false},
		{name: "invalid", raw: "later", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseDuration(tt.raw)
			if (err == nil) != tt.want {
				t.Fatalf("parseDuration(%q) error=%v, want valid=%v", tt.raw, err, tt.want)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	n, err := parseInt("20")
	if err != nil || n != 20 {
		t.Fatalf("parseInt(20) = %d, %v", n, err)
	}
	_, err = parseInt("many")
	if err == nil {
		t.Fatal("expected invalid integer error")
	}
}
