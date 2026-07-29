package utils

import (
	"bytes"
	"encoding/base64"
	"reflect"
	"testing"
)

func TestHexToBytes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []byte
		wantErr bool
	}{
		{"empty", "", nil, true},
		{"odd length", "abc", nil, true},
		{"with 0x prefix", "0x48656c6c6f", []byte("Hello"), false},
		{"without prefix", "48656c6c6f", []byte("Hello"), false},
		{"invalid char", "4xyz", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HexToBytes(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("HexToBytes(%q) expected error, got %v", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Errorf("HexToBytes(%q) unexpected error: %v", tt.input, err)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("HexToBytes(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestBase64ToBytes(t *testing.T) {
	helloB64 := base64.StdEncoding.EncodeToString([]byte("Hello"))
	tests := []struct {
		name    string
		input   string
		want    []byte
		wantErr bool
	}{
		{"empty", "", nil, true},
		{"valid standard", helloB64, []byte("Hello"), false},
		{"valid url", base64.URLEncoding.EncodeToString([]byte("Hello")), []byte("Hello"), false},
		{"invalid", "not-base64!!!", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Base64ToBytes(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Base64ToBytes(%q) expected error, got %v", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Errorf("Base64ToBytes(%q) unexpected error: %v", tt.input, err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Base64ToBytes(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
