package auth_test

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/msanatan/go-chatroom/app/auth"
)

// TestToken is a test token struct
type TestToken struct {
	Sub string `json:"sub"`
	Iat int    `json:"iat"`
}

func Test_GenerateJWT(t *testing.T) {
	secret := "asdf"
	userID := "myuser"

	at(time.Unix(0, 0), func() {
		token, err := auth.GenerateJWT(userID, secret, 60)
		if err != nil {
			t.Fatalf("did not expect an error but received : %q", err.Error())
		}

		// Check parts of the token
		tokenParts := strings.Split(token, ".")
		sDec, _ := base64.RawStdEncoding.DecodeString(tokenParts[1])

		var jwtClaims TestToken
		err = json.Unmarshal(sDec, &jwtClaims)
		if err != nil {
			t.Fatalf("error decoding token %s", err.Error())
		}

		if jwtClaims.Sub != userID {
			t.Errorf("expected subject to be %q but found %q", userID, jwtClaims.Sub)
		}
	})
}

func Test_VerifyJWT(t *testing.T) {
	secret := "asdf"
	testJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdXRob3JpemVkIjp0cnVlLCJpYXQiOjE2MDY4NDQ5MTgsInN1YiI6Im15dXNlciJ9.ACxmz5PxK4UTEg88Arc7u3qyfyoUQ6kCGxw-Z4scaIk"

	userID, err := auth.VerifyJWT(testJWT, secret)
	if err != nil {
		t.Fatalf("did not expect an error but received : %q", err.Error())
	}

	if userID != "myuser" {
		t.Errorf("expected jwt subject to be returned but got none")
	}
}

func at(t time.Time, f func()) {
	jwt.TimeFunc = func() time.Time {
		return t
	}
	f()
	jwt.TimeFunc = time.Now
}
