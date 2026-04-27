package apikey

import (
	"testing"
)

var testSecret = []byte("0123456789abcdef0123456789abcdef")
var testPrefix = "mtv_"

func TestGenerateAndValidate(t *testing.T) {
	plainKey, cipher, err := Generate(testSecret, 42, testPrefix)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if len(plainKey) == 0 {
		t.Fatal("plainKey is empty")
	}
	if len(cipher) == 0 {
		t.Fatal("cipher is empty")
	}

	userID, err := Validate(testSecret, plainKey, testPrefix)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if userID != 42 {
		t.Fatalf("expected userID 42, got %d", userID)
	}
}

func TestValidateRejectsWrongSecret(t *testing.T) {
	plainKey, _, err := Generate(testSecret, 1, testPrefix)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	wrongSecret := []byte("XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
	_, err = Validate(wrongSecret, plainKey, testPrefix)
	if err == nil {
		t.Fatal("expected error with wrong secret")
	}
}

func TestValidateRejectsWrongPrefix(t *testing.T) {
	plainKey, _, err := Generate(testSecret, 1, testPrefix)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	_, err = Validate(testSecret, plainKey, "other_")
	if err == nil {
		t.Fatal("expected error with wrong prefix")
	}
}

func TestValidateRejectsTamperedKey(t *testing.T) {
	plainKey, _, err := Generate(testSecret, 1, testPrefix)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	tampered := plainKey[:len(plainKey)-2] + "XX"
	_, err = Validate(testSecret, tampered, testPrefix)
	if err == nil {
		t.Fatal("expected error with tampered key")
	}
}

func TestValidateRejectsEmptyKey(t *testing.T) {
	_, err := Validate(testSecret, "", testPrefix)
	if err == nil {
		t.Fatal("expected error with empty key")
	}
}

func TestGenerateProducesUniqueKeys(t *testing.T) {
	keys := make(map[string]bool)
	for i := 0; i < 20; i++ {
		plainKey, _, err := Generate(testSecret, 1, testPrefix)
		if err != nil {
			t.Fatalf("Generate %d: %v", i, err)
		}
		if keys[plainKey] {
			t.Fatalf("duplicate key generated at iteration %d", i)
		}
		keys[plainKey] = true
	}
}
