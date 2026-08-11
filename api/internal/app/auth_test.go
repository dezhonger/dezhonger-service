package app

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestUsernameValidation(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"admin", "wei.long", "user_01", "a-b"} {
		if !validateUsername(value) {
			t.Fatalf("expected %q to be valid", value)
		}
	}
	for _, value := range []string{"ab", "Admin", "-admin", "user space"} {
		if validateUsername(value) {
			t.Fatalf("expected %q to be invalid", value)
		}
	}
}

func TestPasswordHash(t *testing.T) {
	t.Parallel()
	password := "a-secure-password"
	hash, err := hashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		t.Fatalf("password did not match generated hash: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte("anything")); err == bcrypt.ErrHashTooShort {
		t.Fatal("dummy password hash is malformed")
	}
}
