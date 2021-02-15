package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// GenerateJWT creates a JWT with custom
var GenerateJWT = func(userID, secret string, expiresIn int) (string, error) {
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["sub"] = userID
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
var VerifyJWT = func(token, secret string) (string, error) {
	// Parse JWT
	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if ok && jwtToken.Valid {
		userID, ok := claims["sub"].(string)
		if !ok {
			return "", errors.New("Failed to verify JWT and extract the subject")
		}

		return userID, nil
	}

	return "", errors.New("Failed to verify JWT and extract the subject")
}
