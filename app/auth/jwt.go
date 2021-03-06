package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// GenerateJWT creates a JWT with custom
var GenerateJWT = func(userID int, username, secret string, expiresIn int) (string, error) {
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["sub"] = username
	claims["userId"] = userID
	claims["iat"] = time.Now().Unix()
	if expiresIn != 0 {
		claims["exp"] = time.Now().Add(time.Minute * time.Duration(expiresIn)).Unix()
	}

	// Create JWT
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return at.SignedString([]byte(secret))
}

// VerifyJWT checks that the JWT is well formed (i.e. it can be parsed) and returns the
// user ID encoded in the JWT.
var VerifyJWT = func(token, secret string) (int, string, error) {
	// Parse JWT
	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, "", err
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if ok && jwtToken.Valid {
		// numbers come as float64s in JSON
		userIDFromJSON, ok := claims["userId"].(float64)
		if !ok {
			return 0, "", errors.New("Failed to verify JWT and extract the subject")
		}
		userID := int(userIDFromJSON)

		username, ok := claims["sub"].(string)
		if !ok {
			return 0, "", errors.New("Failed to verify JWT and extract the username")
		}

		return userID, username, nil
	}

	return 0, "", errors.New("Failed to verify JWT and extract the subject")
}

// GetTokenFromRequest extracts the token from an HTTP request
func GetTokenFromRequest(r *http.Request) string {
	keys := r.URL.Query()
	token := keys.Get("bearer")
	if token != "" {
		return token
	}

	bearerHeader := r.Header.Get("Authorization")
	if len(strings.Split(bearerHeader, " ")) == 2 {
		return strings.Split(bearerHeader, " ")[1]
	}

	return ""
}
