package exporter

import (
	"testing"
	"time"

	"github.com/mssql_ie/config"
)

func TestTableToCSV_EmptyTable(t *testing.T) {
	cfg := config.ExportConfig{
		Table:   "",
		CSVPath: "",
		Output:  nil,
	}
	err := TableToCSV(nil, cfg)
	if err == nil {
		t.Error("expected error for empty table name, got nil")
	}
}

func TestSQLToCSV_EmptySQL(t *testing.T) {
	cfg := config.ExportConfig{
		SQL:     "",
		CSVPath: "",
		Output:  nil,
	}
	err := SQLToCSV(nil, cfg)
	if err == nil {
		t.Error("expected error for empty SQL, got nil")
	}
}

func TestNoopWriteCloser(t *testing.T) {
	w := &noopWriteCloser{}
	if err := w.Close(); err != nil {
		t.Errorf("noopWriteCloser.Close() returned error: %v", err)
	}
}

func TestIsRealBinaryType(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"binary", "binary", true},
		{"varbinary", "varbinary", true},
		{"image", "image", true},
		{"timestamp", "timestamp", true},
		{"rowversion", "rowversion", true},
		{"case insensitive", "BINARY", true},
		{"nvarchar", "nvarchar", false},
		{"int", "int", false},
		{"decimal", "decimal", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRealBinaryType(tt.input); got != tt.want {
				t.Errorf("isRealBinaryType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestConvertBinaryToString(t *testing.T) {
	tests := []struct {
		name         string
		input        []byte
		binaryFormat string
		want         string
	}{
		{"nil", nil, "raw", ""},
		{"empty", []byte{}, "raw", ""},
		{"raw", []byte{0x48, 0x65, 0x6c}, "raw", "Hel"},
		{"hex", []byte{0x48, 0x65, 0x6c}, "hex", "48656c"},
		{"base64", []byte{0x48, 0x65, 0x6c}, "base64", "SGVs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertBinaryToString(tt.input, tt.binaryFormat); got != tt.want {
				t.Errorf("convertBinaryToString(%v, %q) = %q, want %q", tt.input, tt.binaryFormat, got, tt.want)
			}
		})
	}
}

func TestConvertValueToString_Nil(t *testing.T) {
	result := convertValueToString(nil, "nvarchar", "raw", "NULL")
	if result != "NULL" {
		t.Errorf("convertValueToString(nil) = %q, want %q", result, "NULL")
	}

	result = convertValueToString(nil, "nvarchar", "raw", "")
	if result != "" {
		t.Errorf("convertValueToString(nil, '') = %q, want empty", result)
	}
}

func TestConvertValueToString_Types(t *testing.T) {
	tests := []struct {
		name   string
		value  interface{}
		colTyp string
		bf     string
		want   string
	}{
		{"string", "hello", "nvarchar", "raw", "hello"},
		{"int64", int64(42), "int", "raw", "42"},
		{"int32", int32(42), "int", "raw", "42"},
		{"int16", int16(42), "int", "raw", "42"},
		{"int8", int8(42), "tinyint", "raw", "42"},
		{"float64", 3.14, "float", "raw", "3.14"},
		{"bool true", true, "bit", "raw", "true"},
		{"bool false", false, "bit", "raw", "false"},
		{"time datetime", time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC), "datetime", "raw", "2024-01-02 15:04:05.000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertValueToString(tt.value, tt.colTyp, tt.bf, ""); got != tt.want {
				t.Errorf("convertValueToString(%v, %q) = %q, want %q", tt.value, tt.colTyp, got, tt.want)
			}
		})
	}
}
