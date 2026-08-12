package auth

import "testing"

func TestHashPassword_Deterministic(t *testing.T) {
	got1 := HashPassword("secret", "salt", "password")
	got2 := HashPassword("secret", "salt", "password")
	if got1 != got2 {
		t.Fatalf("HashPassword is not deterministic: %q != %q", got1, got2)
	}
}

func TestHashPassword_InputsAffectOutput(t *testing.T) {
	base := HashPassword("secret", "salt", "password")

	tests := map[string]string{
		"different secret":   HashPassword("other-secret", "salt", "password"),
		"different salt":     HashPassword("secret", "other-salt", "password"),
		"different password": HashPassword("secret", "salt", "other-password"),
	}

	for name, got := range tests {
		t.Run(name, func(t *testing.T) {
			if got == base {
				t.Fatalf("expected hash to change when %s, both were %q", name, got)
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	hash := HashPassword("secret", "salt", "correct-password")

	tests := []struct {
		name     string
		secret   string
		salt     string
		password string
		want     bool
	}{
		{"correct credentials", "secret", "salt", "correct-password", true},
		{"wrong password", "secret", "salt", "wrong-password", false},
		{"wrong salt", "secret", "other-salt", "correct-password", false},
		{"wrong secret", "other-secret", "salt", "correct-password", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VerifyPassword(tt.secret, tt.salt, tt.password, hash)
			if got != tt.want {
				t.Errorf("VerifyPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerifyAdminPassword(t *testing.T) {
	tests := []struct {
		name       string
		submitted  string
		configured string
		want       bool
	}{
		{"correct password", "admin-pass", "admin-pass", true},
		{"wrong password", "wrong-pass", "admin-pass", false},
		{"empty submitted", "", "admin-pass", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VerifyAdminPassword("secret", tt.submitted, tt.configured)
			if got != tt.want {
				t.Errorf("VerifyAdminPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateSalt_Unique(t *testing.T) {
	a, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() error = %v", err)
	}
	b, err := GenerateSalt()
	if err != nil {
		t.Fatalf("GenerateSalt() error = %v", err)
	}
	if a == b {
		t.Fatalf("expected two calls to GenerateSalt to differ, both were %q", a)
	}
	if len(a) != saltBytes*2 { // hex-encoded
		t.Fatalf("expected salt of length %d, got %d", saltBytes*2, len(a))
	}
}
