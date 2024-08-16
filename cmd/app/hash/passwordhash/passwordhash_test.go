package passwordhash

import (
	"strings"
	"testing"
)

func TestNewHPasswordHash(t *testing.T) {
	ph := NewHPasswordHash()
	if ph == nil {
		t.Fatal("NewHPasswordHash returned nil")
	}
	if ph.memory != 64*1024 || ph.iterations != 3 || ph.parallelism != 2 || ph.saltLength != 16 || ph.keyLength != 32 {
		t.Error("NewHPasswordHash returned unexpected values")
	}
}

func TestGenerateFromPassword(t *testing.T) {
	ph := NewHPasswordHash()
	password := "testpassword"

	hash, err := ph.GenerateFromPassword(password)
	if err != nil {
		t.Fatalf("GenerateFromPassword failed: %v", err)
	}

	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		t.Errorf("Generated hash has unexpected format: %s", hash)
	}

	if parts[1] != "argon2id" {
		t.Errorf("Unexpected algorithm: %s", parts[1])
	}
}

func TestComparePasswordAndHash(t *testing.T) {
	ph := NewHPasswordHash()
	password := "testpassword"

	hash, err := ph.GenerateFromPassword(password)
	if err != nil {
		t.Fatalf("GenerateFromPassword failed: %v", err)
	}

	match, err := ph.ComparePasswordAndHash(password, hash)
	if err != nil {
		t.Fatalf("ComparePasswordAndHash failed: %v", err)
	}
	if !match {
		t.Error("ComparePasswordAndHash should return true for correct password")
	}

	wrongPassword := "wrongpassword"
	match, err = ph.ComparePasswordAndHash(wrongPassword, hash)
	if err != nil {
		t.Fatalf("ComparePasswordAndHash failed: %v", err)
	}
	if match {
		t.Error("ComparePasswordAndHash should return false for incorrect password")
	}
}

func TestComparePasswordAndHashWithInvalidHash(t *testing.T) {
	ph := NewHPasswordHash()
	password := "testpassword"
	invalidHash := "invalidhash"

	_, err := ph.ComparePasswordAndHash(password, invalidHash)
	if err == nil {
		t.Error("ComparePasswordAndHash should return an error for invalid hash")
	}
}
