package importer

import (
	"strings"
	"testing"

	"github.com/mssql_ie/config"
)

func TestValidateImportConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.ImportConfig
		wantErr bool
	}{
		{"empty table", config.ImportConfig{Table: "", CSVPath: "test.csv", Batch: 1000}, true},
		{"empty csv and input", config.ImportConfig{Table: "t", CSVPath: "", Batch: 1000}, true},
		{"zero batch", config.ImportConfig{Table: "t", CSVPath: "test.csv", Batch: 0}, true},
		{"valid", config.ImportConfig{Table: "t", CSVPath: "test.csv", Batch: 1000}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateImportConfig(tt.cfg)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestBuildInsertSQL(t *testing.T) {
	sql, err := buildInsertSQL("users", []string{"[id]", "[name]"})
	if err != nil {
		t.Fatalf("buildInsertSQL error: %v", err)
	}
	expected := "INSERT INTO [users] ([id],[name]) VALUES (?,?)"
	if sql != expected {
		t.Errorf("buildInsertSQL = %q, want %q", sql, expected)
	}
}

func TestBuildInsertSQL_EmptyCols(t *testing.T) {
	_, err := buildInsertSQL("users", []string{})
	if err == nil {
		t.Error("expected error for empty columns")
	}
}

func TestIsBinaryType(t *testing.T) {
	tests := []struct {
		name     string
		dataType string
		want     bool
	}{
		{"binary", "binary", true},
		{"varbinary", "varbinary", true},
		{"image", "image", true},
		{"nvarchar", "nvarchar", false},
		{"int", "int", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBinaryType(tt.dataType); got != tt.want {
				t.Errorf("isBinaryType(%q) = %v, want %v", tt.dataType, got, tt.want)
			}
		})
	}
}

func TestGetDefaultValue(t *testing.T) {
	tests := []struct {
		name     string
		dataType string
		want     interface{}
		checkFn  func(t *testing.T, got interface{})
	}{
		{"int", "int", 0, func(t *testing.T, got interface{}) {
			if v, ok := got.(int); !ok || v != 0 {
				t.Errorf("got %v (%T), want 0 (int)", got, got)
			}
		}},
		{"bit", "bit", false, func(t *testing.T, got interface{}) {
			if v, ok := got.(bool); !ok || v != false {
				t.Errorf("got %v (%T), want false (bool)", got, got)
			}
		}},
		{"binary", "binary", []byte{}, func(t *testing.T, got interface{}) {
			v, ok := got.([]byte)
			if !ok {
				t.Errorf("got %T, want []byte", got)
				return
			}
			if len(v) != 0 {
				t.Errorf("got len=%d, want empty slice", len(v))
			}
		}},
		{"nvarchar", "nvarchar", "", func(t *testing.T, got interface{}) {
			if v, ok := got.(string); !ok || v != "" {
				t.Errorf("got %v (%T), want empty string", got, got)
			}
		}},
		{"varchar", "varchar", "", func(t *testing.T, got interface{}) {
			if v, ok := got.(string); !ok || v != "" {
				t.Errorf("got %v (%T), want empty string", got, got)
			}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDefaultValue(tt.dataType)
			tt.checkFn(t, got)
		})
	}
}

func TestIsValidWKT(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"point", "POINT(1 2)", true},
		{"linestring", "LINESTRING(1 2, 3 4)", true},
		{"polygon", "POLYGON((1 2, 3 4, 5 6, 1 2))", true},
		{"multipoint", "MULTIPOINT(1 2, 3 4)", true},
		{"multi linestring", "MULTILINESTRING((1 2, 3 4))", true},
		{"multipolygon", "MULTIPOLYGON(((1 2, 3 4, 5 6, 1 2)))", true},
		{"geometry collection", "GEOMETRYCOLLECTION(POINT(1 2))", true},
		{"invalid", "NOT_A_WKT", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidWKT(tt.input); got != tt.want {
				t.Errorf("isValidWKT(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNoopReadCloser(t *testing.T) {
	r := &noopReadCloser{Reader: strings.NewReader("test")}
	if err := r.Close(); err != nil {
		t.Errorf("noopReadCloser.Close() error: %v", err)
	}
}

func TestConvertBinary(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		binaryFormat string
		want         string
	}{
		{"raw", "hello", "raw", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertBinary(tt.value, tt.binaryFormat)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(result) != tt.want {
				t.Errorf("convertBinary(%q, %q) = %v, want %v", tt.value, tt.binaryFormat, result, tt.want)
			}
		})
	}
}

func TestSetArgToNull(t *testing.T) {
	// non-binary type
	col := ColumnInfo{Name: "name", DataType: "nvarchar"}
	var arg interface{}
	setArgToNull(&arg, col)
	if arg != nil {
		t.Errorf("expected nil for nvarchar, got %v", arg)
	}

	// binary type
	col = ColumnInfo{Name: "data", DataType: "varbinary"}
	setArgToNull(&arg, col)
	if arg == nil {
		t.Error("expected non-nil []byte for varbinary, got nil")
	}
	if b, ok := arg.([]byte); !ok || b != nil {
		t.Errorf("expected nil []byte for varbinary, got %T=%v", arg, arg)
	}
}

func TestConvertValue(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		col          ColumnInfo
		binaryFormat string
		want         interface{}
	}{
		{"string", "hello", ColumnInfo{DataType: "nvarchar"}, "raw", "hello"},
		{"bit true", "true", ColumnInfo{DataType: "bit"}, "raw", true},
		{"bit false", "0", ColumnInfo{DataType: "bit"}, "raw", false},
		{"bit yes", "yes", ColumnInfo{DataType: "bit"}, "raw", true},
		{"empty string", "", ColumnInfo{DataType: "nvarchar"}, "raw", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertValue(tt.value, tt.col, tt.binaryFormat)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("convertValue(%q, %+v) = %v (%T), want %v (%T)",
					tt.value, tt.col, got, got, tt.want, tt.want)
			}
		})
	}
}

// TestBatchInsertConfig ensures the config struct is used correctly
func TestBatchInsertConfig(t *testing.T) {
	cfg := BatchInsertConfig{
		BatchSize:      1000,
		SkipErrors:     true,
		IdentityInsert: false,
		BinaryFormat:   "raw",
	}
	if cfg.BatchSize != 1000 {
		t.Errorf("BatchSize = %d, want 1000", cfg.BatchSize)
	}
	if !cfg.SkipErrors {
		t.Error("SkipErrors should be true")
	}
}

// TestCSVToTable_NoDB tests CSVToTable with nil DB (should fail at db operations)
func TestCSVToTable_Validation(t *testing.T) {
	cfg := config.ImportConfig{
		Table:   "",
		CSVPath: "test.csv",
		Batch:   1000,
	}
	err := CSVToTable(nil, cfg)
	if err == nil {
		t.Error("expected error for empty table, got nil")
	}
}
