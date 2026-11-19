package security

import (
	"encoding/hex"
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("rahasia123")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword("rahasia123", hash) {
		t.Error("password benar harus lolos")
	}
	if CheckPassword("salah", hash) {
		t.Error("password salah harus ditolak")
	}
}

// Vektor uji resmi PBKDF2-HMAC-SHA256 (RFC 7914 bagian 11).
func TestPBKDF2KnownVector(t *testing.T) {
	got := hex.EncodeToString(pbkdf2SHA256([]byte("passwd"), []byte("salt"), 1))
	want := "55ac046e56e3089fec1691c22544b605f94185216dde0465e68b9d57c20dacbc"
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
