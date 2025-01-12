package auth

import (
	"encoding/base64"
	"math/rand"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/leetsecure/qryptic-controller/internal/config"
	"github.com/leetsecure/qryptic-controller/internal/models"
	"github.com/leetsecure/qryptic-controller/internal/utils/logger"
	"golang.org/x/crypto/bcrypt"
)

func CreateUserToken(userUuid string, role models.UserRoleEnum) (string, error) {
	log := logger.Default()
	var jwtUserAuthSecretKey = []byte(config.UserAuthJwtSecretKey)
	timeNow := time.Now()

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userUuid,                                   // Subject (user identifier)
		"iss": "qryptic-controller",                       // Issuer
		"aud": string(role),                               // Audience (user role)
		"exp": timeNow.Add(config.JwtTokenTimeout).Unix(), // Expiration time
		"iat": timeNow.Unix(),                             // Issued at
	})

	tokenString, err := claims.SignedString(jwtUserAuthSecretKey)
	if err != nil {
		log.Errorf("Error in creating signed token for user uuid : %s", userUuid)
		return "", err
	}
	return tokenString, nil
}

func CreateVpnGatewayToken(vpnGatewayUuid string, jwtVpnGatewayAuthSecretKey string) (string, error) {
	jwtVpnGatewayAuthSecretKeyBytes := []byte(jwtVpnGatewayAuthSecretKey)
	log := logger.Default()
	timeNow := time.Now()

	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": vpnGatewayUuid,                             // Subject (user identifier)
		"iss": "qryptic-controller",                       // Issuer
		"aud": "Controller",                               // Audience (user role)
		"exp": timeNow.Add(config.JwtTokenTimeout).Unix(), // Expiration time
		"iat": timeNow.Unix(),                             // Issued at
	})

	tokenString, err := claims.SignedString(jwtVpnGatewayAuthSecretKeyBytes)
	if err != nil {
		log.Errorf("Error in creating signed token for vpn gateway uuid : %s", vpnGatewayUuid)
		return "", err
	}
	return tokenString, nil
}

func VerifyVpnGatewayAuthToken(vpnGatewayAuthToken string, jwtVpnGatewayAuthSecretKey string) (string, error) {
	jwtVpnGatewayAuthSecretKeyBytes := []byte(jwtVpnGatewayAuthSecretKey)

	log := logger.Default()

	token, err := jwt.Parse(vpnGatewayAuthToken, func(token *jwt.Token) (interface{}, error) {
		return jwtVpnGatewayAuthSecretKeyBytes, nil
	})

	// Check for verification errors
	if err != nil {
		log.Error("Error in parsing and verifying vpn gateway token")
		return "", err
	}

	// Check if the token is valid
	if !token.Valid {
		log.Info("Invalid vpn gateway auth token")
		return "", err
	}

	vpnGatewayUuid, err := token.Claims.GetSubject()
	if err != nil {
		log.Error("Error in getting vpn gateway uuid from token")
		return "", err
	}

	return vpnGatewayUuid, nil
}

func VerifyUserAuthToken(userAuthToken string) (string, models.UserRoleEnum, error) {
	log := logger.Default()
	var jwtUserAuthSecretKey = []byte(config.UserAuthJwtSecretKey)
	token, err := jwt.Parse(userAuthToken, func(token *jwt.Token) (interface{}, error) {
		return jwtUserAuthSecretKey, nil
	})

	// Check for verification errors
	if err != nil {
		log.Error("Error in parsing and verifying user auth token")
		return "", models.DefaultRole, err
	}

	// Check if the token is valid
	if !token.Valid {
		log.Info("Invalid user auth token")
		return "", models.DefaultRole, err
	}

	userUuid, err := token.Claims.GetSubject()
	if err != nil {
		log.Error("Error in getting user uuid from token")
		return "", models.DefaultRole, err
	}
	userRole, err := token.Claims.GetAudience()
	if err != nil {
		log.Error("Error in getting user role from token")
		return "", models.DefaultRole, err
	}
	return userUuid, models.UserRoleEnum(userRole[0]), nil
}

func VerifyPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func CreatePasswordHash(password string) (string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	log := logger.Default()
	if err != nil {
		log.Error("Error in generating hash for password")
		return "", err
	}

	return string(passwordHash), nil
}

func RandomStringGenerator(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

// base64URLEncode performs the base64 URL-safe encoding
// without padding (i.e., no '=' characters).
func Base64URLEncode(input []byte) string {
	encoded := base64.StdEncoding.EncodeToString(input)
	encoded = strings.TrimRight(encoded, "=")       // Remove padding
	encoded = strings.ReplaceAll(encoded, "+", "-") // Replace '+' with '-'
	encoded = strings.ReplaceAll(encoded, "/", "_") // Replace '/' with '_'
	return encoded
}
