package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrNoAuthHeaderIncluded = errors.New("not auth header included in request")

func HashPassword(pwd string) (string, error) {
	dat, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(dat), nil
}

func CheckPasswordHash(pwd, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd))
}

func MakeJWT(userID int, tokenSecret string, expiresIn time.Duration, issuer string) (string, error) {
	signingKey := []byte(tokenSecret)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   fmt.Sprintf("%d", userID),
	})
	return token.SignedString(signingKey)
}

func ValidateJWT(tokenString, tokenSecret string) (string, string, error) {
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claimsStruct,
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return "", "", err
	}
	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return "", "", err
	}
	tokenIssuer, err := token.Claims.GetIssuer()
	if err != nil {
		return "", "", err
	}

	return userIDString, tokenIssuer, nil
}

func GetBearerToken(headers http.Header, tokenName string) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", ErrNoAuthHeaderIncluded
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != tokenName {
		return "", errors.New("malformed authorization header")
	}

	return splitAuth[1], nil
}
