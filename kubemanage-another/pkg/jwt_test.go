package pkg

import (
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

func TestJWTTokenParseTokenAcceptsOnlyHS256(t *testing.T) {
	issuer := &jwtToken{secret: "test-secret"}

	valid, err := issuer.GenerateToken(BaseClaims{ID: 7, AuthorityId: 222, TokenVersion: 3})
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	claims, err := issuer.ParseToken(valid)
	if err != nil {
		t.Fatalf("ParseToken() valid token error = %v", err)
	}
	if claims.ID != 7 || claims.AuthorityId != 222 || claims.TokenVersion != 3 {
		t.Fatalf("ParseToken() claims = %#v", claims.BaseClaims)
	}

	wrongAlgorithm := jwt.NewWithClaims(jwt.SigningMethodHS384, CustomClaims{
		BaseClaims: BaseClaims{ID: 7},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour).Unix(),
		},
	})
	wrongAlgorithmToken, err := wrongAlgorithm.SignedString([]byte(issuer.secret))
	if err != nil {
		t.Fatalf("sign HS384 token: %v", err)
	}
	if _, err := issuer.ParseToken(wrongAlgorithmToken); err == nil {
		t.Fatal("ParseToken() accepted a token signed with HS384")
	}
}

func TestJWTTokenParseTokenRejectsExpiredToken(t *testing.T) {
	issuer := &jwtToken{secret: "test-secret"}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomClaims{
		StandardClaims: jwt.StandardClaims{ExpiresAt: time.Now().Add(-time.Minute).Unix()},
	})
	signed, err := token.SignedString([]byte(issuer.secret))
	if err != nil {
		t.Fatalf("sign expired token: %v", err)
	}
	if _, err := issuer.ParseToken(signed); err == nil {
		t.Fatal("ParseToken() accepted an expired token")
	}
}
