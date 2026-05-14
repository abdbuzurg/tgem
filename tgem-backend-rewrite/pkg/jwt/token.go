package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/spf13/viper"
)

type Payload struct {
	jwt.StandardClaims
	UserID    uint `json:"userID"`
	WorkerID  uint `json:"workerID"`
	RoleID    uint `json:"roleID"`
	ProjectID uint `json:"projectID"`
}

// secretKey is read lazily so viper has time to load the config file
// (config.GetConfig runs in main.init, after this package is initialized).
func secretKey() []byte {
	return []byte(viper.GetString("Jwt.Secret"))
}

func CreateToken(userID, workerID, roleID, projectID uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Payload{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(10 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		userID,
		workerID,
		roleID,
		projectID,
	})

	tokenString, err := token.SignedString(secretKey())

	if err != nil {
		return "", fmt.Errorf("could not sign the token: %v", err)
	}

	return tokenString, err
}

func VerifyToken(token string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errors.New("invalid token signature")
		}
		return secretKey(), nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		return nil, errors.New("invalid token format")
	}

	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, errors.New("invalid token payload")
	}

	return payload, nil
}
