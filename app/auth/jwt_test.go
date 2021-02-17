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
	UserID int    `json:"userId"`
	Sub    string `json:"sub"`
	Iat    int    `json:"iat"`
}

func Test_GenerateJWT(t *testing.T) {
	secret := "asdf"
	userID := 123
	username := "myusername"

	at(time.Unix(0, 0), func() {
		token, err := auth.GenerateJWT(userID, username, secret, 60)
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

		if jwtClaims.Sub != username {
			t.Errorf("expected subject to be %q but found %q", username, jwtClaims.Sub)
		}

		if jwtClaims.UserID != userID {
			t.Errorf("expected username to be %d but found %d", userID, jwtClaims.UserID)
		}
	})
}

func Test_VerifyJWT(t *testing.T) {
	secret := "asdf"
	testJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdXRob3JpemVkIjp0cnVlLCJpYXQiOjE2MDY4NDQ5MTgsInN1YiI6Im15dXNlcm5hbWUiLCJ1c2VySWQiOjEyM30.X2CnOokXdtwaLUZmTiGZMh_i7rQhVp1Wy57W_ujBWxw"

	userID, username, err := auth.VerifyJWT(testJWT, secret)
	if err != nil {
		t.Fatalf("did not expect an error but received : %q", err.Error())
	}

	if userID != 123 {
		t.Errorf("expected jwt userId to be %d but got %d", 123, userID)
	}

	if username != "myusername" {
		t.Errorf("expected jwt subject to be %q but got %q", "myusername", username)
	}
}

func at(t time.Time, f func()) {
	jwt.TimeFunc = func() time.Time {
		return t
	}
	f()
	jwt.TimeFunc = time.Now
}
