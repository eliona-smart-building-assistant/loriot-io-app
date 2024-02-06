package loriot

import "testing"

// TestIsValidEUI64 tests the isValidEUI64 function.
func TestIsValidEUI64(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"Valid EUI-16", "0123456789ABCDEF", true},
		{"Valid EUI-32", "0123456789ABCDEF0123456789ABCDEF", true},
		{"Valid EUI-64", "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF", true},
		{"Invalid length 1", "01234567", false},
		{"Invalid length 2", "0123456789ABCDEF0123456789ABCDEFR", false},
		{"Invalid characters", "G123456789ABCDE", false},
		{"Empty string", "", false},
		{"Nil pointer", "", false}, // Special case to test nil pointer
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var idPtr *string
			if tt.name != "Nil pointer" {
				idPtr = &tt.id
			}
			if got := IsValidEUI(idPtr); got != tt.want {
				t.Errorf("isValidEUI64() = %v, want %v", got, tt.want)
			}
		})
	}
}
