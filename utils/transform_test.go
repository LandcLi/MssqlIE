package utils

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestGetTransformersWrite(t *testing.T) {
	tests := []struct {
		name   string
		format string
		input  string
	}{
		{"raw", "raw", "Hello, 世界"},
		{"utf8", "utf8", "Hello, 世界"},
		{"gbk", "gbk", "Hello, 世界"},
		{"iso-8859-1", "iso-8859-1", "Hello"},
		{"case insensitivity", "GBK", "测试"},
		{"windows-1252", "windows-1252", "Hello"},
		{"cp1252", "cp1252", "Hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := GetTransformersWrite(&buf, tt.format)
			_, err := io.WriteString(w, tt.input)
			if err != nil {
				t.Fatalf("GetTransformersWrite(%q) write error: %v", tt.format, err)
			}
			if buf.Len() == 0 {
				t.Errorf("GetTransformersWrite(%q) produced empty output", tt.format)
			}
		})
	}
}

func TestGetTransformersRead(t *testing.T) {
	tests := []struct {
		name   string
		format string
		input  string
	}{
		{"raw", "raw", "Hello, 世界"},
		{"utf8", "utf8", "Hello, 世界"},
		{"gbk", "gbk", "测试"},
		{"iso-8859-1", "iso-8859-1", "Hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := GetTransformersRead(strings.NewReader(tt.input), tt.format)
			data, err := io.ReadAll(r)
			if err != nil {
				t.Fatalf("GetTransformersRead(%q) read error: %v", tt.format, err)
			}
			if len(data) == 0 {
				t.Errorf("GetTransformersRead(%q) produced empty output", tt.format)
			}
		})
	}
}

func TestGetTransformersWriteRoundTrip(t *testing.T) {
	original := "Hello, 世界"
	var encodedBuf bytes.Buffer

	w := GetTransformersWrite(&encodedBuf, "gbk")
	io.WriteString(w, original)

	r := GetTransformersRead(bytes.NewReader(encodedBuf.Bytes()), "gbk")
	decoded, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("round trip read error: %v", err)
	}

	if string(decoded) != original {
		t.Errorf("round trip = %q, want %q", string(decoded), original)
	}
}
