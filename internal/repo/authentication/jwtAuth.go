package authentication

import (
	"fmt"
	"github.com/PereRohit/util/log"
	"github.com/dgrijalva/jwt-go"
	"time"
)

// JWTService defines the interface for JWT authentication service
type JWTService interface {
	GenerateToken(signingMethod jwt.SigningMethod, userId string, validity time.Duration) (string, error)
	ValidateToken(token string) (*jwt.Token, error)
}

type authCustomClaims struct {
	UserId string `json:"user_id"`
	jwt.StandardClaims
}

type jwtService struct {
	secretKey string
	userId    string
}

// JWTAuthService returns a new instance of JWT authentication service
func JWTAuthService(secret string) JWTService {
	return &jwtService{
		secretKey: getSecretKey(secret),
	}
}

// getSecretKey returns the default secret key if it is not provided
func getSecretKey(secret string) string {
	if secret == "" {
		secret = "DefaultSecretJwtKey"
	}
	return secret
}

// GenerateToken generates a new JWT token
func (service *jwtService) GenerateToken(signingMethod jwt.SigningMethod, userId string, validity time.Duration) (string, error) {
	var currentTime = time.Now().UTC()
	claims := &authCustomClaims{
		userId,
		jwt.StandardClaims{
			ExpiresAt: currentTime.Add(validity).Unix(),
			IssuedAt:  currentTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(signingMethod, claims)

	t, err := token.SignedString([]byte(service.secretKey))
	if err != nil {
		log.Error(err)
		return "", err
	}
	return t, nil
}

// ValidateToken validates a JWT token
func (service *jwtService) ValidateToken(encodedToken string) (*jwt.Token, error) {
	return jwt.Parse(encodedToken, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			err := fmt.Errorf("invalid token %+v", token.Header["alg"])
			return nil, err
		}
		return []byte(service.secretKey), nil
	})

}
