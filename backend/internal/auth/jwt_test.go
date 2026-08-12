package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestIssueAndParseToken_RoundTrip(t *testing.T) {
	token, err := IssueToken("secret", "u1", "Ada", "ada@example.com", RoleUser)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	claims, err := ParseToken("secret", token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.UserID != "u1" || claims.Name != "Ada" || claims.Email != "ada@example.com" || claims.Role != RoleUser {
		t.Fatalf("claims did not round-trip: %+v", claims)
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	token, err := IssueToken("secret", "u1", "Ada", "ada@example.com", RoleUser)
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}

	if _, err := ParseToken("wrong-secret", token); err == nil {
		t.Fatal("expected ParseToken to reject a token signed with a different secret")
	}
}

func TestParseToken_Expired(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour)
	claims := Claims{
		UserID: "u1",
		Role:   RoleUser,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(past.Add(-TokenTTL)),
			ExpiresAt: jwt.NewNumericDate(past),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign expired token: %v", err)
	}

	if _, err := ParseToken("secret", token); err == nil {
		t.Fatal("expected ParseToken to reject an expired token")
	}
}

func TestParseToken_RejectsNoneAlgorithm(t *testing.T) {
	claims := Claims{UserID: "admin", Role: RoleAdmin}
	token, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign none-alg token: %v", err)
	}

	if _, err := ParseToken("secret", token); err == nil {
		t.Fatal("expected ParseToken to reject a token using the none algorithm")
	}
}

func TestParseToken_Malformed(t *testing.T) {
	if _, err := ParseToken("secret", "not-a-jwt"); err == nil {
		t.Fatal("expected ParseToken to reject a malformed token string")
	}
}
