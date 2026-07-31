package utils

import (
	"testing"
)

func TestEscapeIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"simple name", "users", "[users]"},
		{"with spaces", "user name", "[user name]"},
		{"already bracketed", "[users]", "[users]"},
		{"bracketed with space", "[user name]", "[user name]"},
		{"special chars", "order_details", "[order_details]"},
		{"contains right bracket", "col]]", "[col]]]]]"},
		{"trim whitespace", "  users  ", "[users]"},
		{"numeric suffix", "table_1", "[table_1]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("EscapeIdentifier(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEscapeQualifiedName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"empty string", "", "", false},
		{"simple table", "users", "[users]", false},
		{"schema.table", "dbo.users", "[dbo].[users]", false},
		{"already escaped", "[dbo].[users]", "[dbo].[users]", false},
		{"three parts", "db.dbo.users", "[db].[dbo].[users]", false},
		{"four parts", "srv.db.dbo.users", "[srv].[db].[dbo].[users]", false},
		{"mixed brackets", "[dbo].users", "[dbo].[users]", false},
		{"invalid char", "user name!", "", true},
		{"unmatched bracket", "[users", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EscapeQualifiedName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("EscapeQualifiedName(%q) expected error, got %q", tt.input, result)
				}
				return
			}
			if err != nil {
				t.Errorf("EscapeQualifiedName(%q) unexpected error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("EscapeQualifiedName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSplitQualifiedName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
		wantErr  bool
	}{
		{"empty string", "", nil, false},
		{"simple table", "users", []string{"users"}, false},
		{"schema.table", "dbo.users", []string{"dbo", "users"}, false},
		{"already escaped", "[dbo].[users]", []string{"dbo", "users"}, false},
		{"three parts", "db.dbo.users", []string{"db", "dbo", "users"}, false},
		{"four parts", "srv.db.dbo.users", []string{"srv", "db", "dbo", "users"}, false},
		{"invalid char", "user name!", nil, true},
		{"unmatched bracket", "[users", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SplitQualifiedName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("SplitQualifiedName(%q) expected error, got %v", tt.input, result)
				}
				return
			}
			if err != nil {
				t.Errorf("SplitQualifiedName(%q) unexpected error: %v", tt.input, err)
				return
			}
			if len(result) != len(tt.expected) {
				t.Errorf("SplitQualifiedName(%q) = %v, want %v", tt.input, result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("SplitQualifiedName(%q) = %v, want %v", tt.input, result, tt.expected)
					break
				}
			}
		})
	}
}

func TestIsValidGUID(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"6F9619FF-8B86-D011-B42D-00C04FC964FF", true},
		{"6f9619ff-8b86-d011-b42d-00c04fc964ff", true},
	{"not-a-guid", false},
	{"too-short", false},
		{"", false},
		{"xxx-xxxx-xxxx-xxxx-xxxxxxxx", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsValidGUID(tt.input)
			if result != tt.want {
				t.Errorf("IsValidGUID(%q) = %v, want %v", tt.input, result, tt.want)
			}
		})
	}
}
