package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/leetsecure/qryptic-controller/internal/config"
	"github.com/leetsecure/qryptic-controller/internal/database"
	"github.com/leetsecure/qryptic-controller/internal/models"
	"github.com/leetsecure/qryptic-controller/internal/utils/auth"
	"github.com/leetsecure/qryptic-controller/internal/utils/logger"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

func UserLogin(emailID string, password string) (string, error) {
	log := logger.Default()
	if !config.AllowPasswordLogin {
		return "", errors.New("login using email and password not allowed")
	}
	exists := ifUserEmailAlreadyPresent(emailID)
	if !exists {
		log.Infof("User with email id %s is not present", emailID)
		return "", errors.New("email id not present")
	}
	var user models.User
	err := database.DB.Where("email = ?", emailID).First(&user).Error
	if err != nil {
		return "", err
	}
	userRole := user.Role
	userUuid := user.UUID
	userPasswordHash := user.PasswordHash
	userIsPasswordSet := user.IsPasswordSet
	if !userIsPasswordSet {
		return "", fmt.Errorf("password is not set for %s user", userUuid)
	}
	err = auth.VerifyPassword(password, userPasswordHash)
	if err != nil {
		return "", err
	}
	userToken, err := auth.CreateUserToken(userUuid, userRole)
	if err != nil {
		return "", err
	}
	return userToken, nil
}

func UserSSOLogin(emailID string) (string, error) {
	log := logger.Default()
	if !config.AllowSSOLogin {
		return "", errors.New("login using sso not allowed")
	}
	exists := ifUserEmailAlreadyPresent(emailID)
	if !exists {
		log.Infof("User with email id %s is not present", emailID)
		return "", errors.New("email id not present")
	}
	var user models.User
	err := database.DB.Where("email = ?", emailID).First(&user).Error
	if err != nil {
		return "", err
	}
	userRole := user.Role
	userUuid := user.UUID

	userToken, err := auth.CreateUserToken(userUuid, userRole)
	if err != nil {
		return "", err
	}
	return userToken, nil
}

func AuthProviderValidate(provider string) error {
	if !config.AllowSSOLogin {
		return errors.New("sso login not allowed")
	}
	if provider != "google" {
		return errors.New("sso provider not allowed")
	}
	return nil
}

func InitiateGoogleSsoAuth(clientId, platform, codeChallenge, redirectUri, codeChallengeMethod string) (string, error) {
	log := logger.Default()
	googleOauth2Config := &oauth2.Config{
		ClientID:     config.GoogleClientID,
		ClientSecret: config.GoogleClientSecret,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	}

	if clientId != "" {
		googleOauth2Config = &oauth2.Config{
			ClientID: clientId,
			Scopes:   []string{"openid", "profile", "email"},
			Endpoint: google.Endpoint,
		}
	}

	stateJWT, err := generateStateJWT(codeChallenge, redirectUri)
	if err != nil {
		log.Error(err)
		return "", errors.New("failed to generate state jwt")
	}

	googleOauth2Config.RedirectURL = redirectUri
	authURL := googleOauth2Config.AuthCodeURL(stateJWT, oauth2.AccessTypeOffline) + "&code_challenge=" + codeChallenge + "&code_challenge_method=" + codeChallengeMethod

	return authURL, nil
}

func generateStateJWT(codeChallenge string, redirectUrl string) (string, error) {
	state := uuid.New().String() // Generate a random state string for CSRF protection
	timeNow := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"codeChallenge": codeChallenge,
		"redirectUrl":   redirectUrl,
		"state":         state,
		"exp":           timeNow.Add(config.SSOStateJwtTokenTimeout).Unix(), // Expiration time
		"iat":           timeNow.Unix(),                                     // Issued at
	})
	return token.SignedString([]byte(config.UserAuthSSOJwtSecretKey))
}

func ValidateStateJWT(stateJWT string) (string, string, error) {
	log := logger.Default()
	// Decode and validate the JWT
	token, err := jwt.Parse(stateJWT, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(config.UserAuthSSOJwtSecretKey), nil
	})

	if err != nil {
		return "", "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", errors.New("invalid state JWT")
	}
	codeChallenge, ok := claims["codeChallenge"].(string)
	if !ok {
		log.Info("missing code_challenge in state")
		return "", "", errors.New("missing code_challenge in state")
	}
	log.Info(codeChallenge)

	redirectUrl, ok := claims["redirectUrl"].(string)
	if !ok {
		log.Info("missing redirectUrl in state")
		return "", "", errors.New("missing redirectUrl in state")
	}
	log.Info(redirectUrl)
	return codeChallenge, redirectUrl, nil
}

// Verify that the code verifier matches the code challenge
func VerifyCodeVerifier(codeVerifier, codeChallenge string) bool {
	hash := sha256.New()
	hash.Write([]byte(codeVerifier))
	hashedVerifier := hash.Sum(nil)
	codeChallengeComputed := auth.Base64URLEncode(hashedVerifier)
	return codeChallenge == codeChallengeComputed
}

// Exchange authorization code for tokens from Google
func ExchangeCodeForTokens(code, codeVerifier, redirectUrl string) (*models.GoogleTokenResponse, error) {
	log := logger.Default()
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", config.GoogleClientID)
	data.Set("client_secret", config.GoogleClientSecret)
	data.Set("redirect_uri", redirectUrl) // Replace with your redirect URI
	data.Set("grant_type", "authorization_code")
	data.Set("code_verifier", codeVerifier)
	data.Set("scope", "https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile")

	log.Info(data)

	// Make the POST request to Google's token endpoint
	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for tokens: %w", err)
	}
	log.Info(resp.StatusCode)
	defer resp.Body.Close()

	// Check for a successful response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to exchange code for tokens: %s", resp.Status)
	}

	body, _ := io.ReadAll(resp.Body)
	var tokenResponse models.GoogleTokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, err
	}
	log.Info("tokenResponse: ", tokenResponse.AccessToken, tokenResponse.IDToken, tokenResponse.RefreshToken)
	return &tokenResponse, nil
}

// Function to get user information (email, name, etc.) from Google
func GetUserInfoFromGoogle(accessToken string) (models.UserInfoResponse, error) {
	log := logger.Default()
	url := "https://www.googleapis.com/oauth2/v2/userinfo"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return models.UserInfoResponse{}, err
	}

	verifyAccessToken(accessToken)

	// Set Authorization header with the access token
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return models.UserInfoResponse{}, err
	}
	log.Info(resp.StatusCode)
	defer resp.Body.Close()

	// Decode the response to extract user info (email, name, etc.)
	var userInfoResponse models.UserInfoResponse
	err = json.NewDecoder(resp.Body).Decode(&userInfoResponse)
	if err != nil {
		return models.UserInfoResponse{}, err
	}
	log.Info(userInfoResponse.Email)
	return userInfoResponse, nil
}

func verifyAccessToken(accessToken string) (bool, error) {
	log := logger.Default()
	url := fmt.Sprintf("https://www.googleapis.com/oauth2/v3/tokeninfo?access_token=%s", accessToken)
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// If the response status is 200 OK, the token is valid
	if resp.StatusCode == http.StatusOK {
		log.Info("token valid")
		return true, nil
	}
	log.Info("token invalid")
	// If we get a 401 or any other error, the token is invalid or expired
	return false, fmt.Errorf("invalid or expired token")
}

func VerifyGoogleSSOToken(token string) (string, error) {
	log := logger.Default()
	tokenValidationResp, err := idtoken.Validate(context.Background(), token, config.GoogleClientID)

	if err != nil {
		log.Error(err)
		return "", err
	}
	log.Infof("audience:%s \nclaims:%s \nexpires:%s \n issuer: %s", tokenValidationResp.Audience, tokenValidationResp.Claims, tokenValidationResp.Expires, tokenValidationResp.Issuer)

	if tokenValidationResp.Expires < time.Now().Unix() {
		log.Info("idtoken invalid")
		return "", errors.New("id token expired")
	}

	return tokenValidationResp.Claims["email"].(string), nil
}

func WebGoogleLoginInitiate(code_challenge string) (string, error) {
	log := logger.Default()
	oauthState := uuid.NewString()
	var googleOauth2Config = &oauth2.Config{
		ClientID:     config.GoogleClientID,
		ClientSecret: config.GoogleClientSecret,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
		RedirectURL:  fmt.Sprintf(config.SSOCallbackTemplate, config.ControllerDomain, "google"),
	}
	authURL := googleOauth2Config.AuthCodeURL(oauthState, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	log.Info(authURL)
	auth := models.Auth{
		UUID:          uuid.NewString(),
		Provider:      "google",
		State:         oauthState,
		CodeChallenge: code_challenge,
		ExpiryTime:    time.Now().Add(2 * time.Minute),
	}
	err := database.DB.Save(&auth).Error
	if err != nil {
		log.Error(err)
		return "", err
	}
	return authURL, nil
}

func WebGoogleLoginCallback(state, code string) error {
	log := logger.Default()
	var googleOauth2Config = &oauth2.Config{
		ClientID:     config.GoogleClientID,
		ClientSecret: config.GoogleClientSecret,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
		RedirectURL:  fmt.Sprintf(config.SSOCallbackTemplate, config.ControllerDomain, "google"),
	}
	token, err := googleOauth2Config.Exchange(context.Background(), code)
	if err != nil {
		log.Infof("code exchange wrong: %s", err.Error())
		return err
	}

	var auth models.Auth
	err = database.DB.Where("state = ?", state).First(&auth).Error
	if err != nil {
		log.Errorf("error in fetching record of state : %s from auth | error : %s", state, err)
		return err
	}

	userInfoResponse, err := GetUserInfoFromGoogle(token.AccessToken)
	if err != nil {
		log.Errorf("error in fetching user info from google | error : %s", err)
		return err
	}
	timeNow := time.Now()
	timeExpiry := auth.ExpiryTime

	if timeNow.After(timeExpiry) {
		log.Errorf("expired state %s", state)
		return errors.New("expired state")
	}

	auth.Authenticated = true
	auth.Email = userInfoResponse.Email
	err = database.DB.Save(&auth).Error
	if err != nil {
		log.Errorf("error in saving auth details for state : %s in auth | error : %s", state, err)
		return err
	}
	return nil
}

func WebGoogleLoginToken(code_verifier, code_challenge string) (bool, string, error) {
	log := logger.Default()
	isVerified := VerifyCodeVerifier(code_verifier, code_challenge)
	if !isVerified {
		log.Errorf("incorrect code_verifier: %s and code_challenge: %s pair", code_verifier, code_challenge)
		return true, "", errors.New("incorrect code_verifier and code_challenge pair")
	}
	var auth models.Auth
	err := database.DB.Order("id DESC").Where("code_challenge = ?", code_challenge).First(&auth).Error
	if err != nil {
		log.Errorf("error in fetching record of code_challenge : %s from auth | error : %s", code_challenge, err)
		return true, "", err
	}

	timeNow := time.Now()
	timeExpiry := auth.ExpiryTime

	if timeNow.After(timeExpiry) {
		log.Errorf("expired session %s", code_challenge)
		return true, "", errors.New("expired session")
	}

	if !auth.Authenticated {
		return false, "", errors.New("unauthenticated")
	}
	authToken, err := UserSSOLogin(auth.Email)
	if err != nil {
		log.Error(err)
		return true, "", err
	}
	return true, authToken, nil
}
