package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leetsecure/qryptic-controller/internal/models"
	"github.com/leetsecure/qryptic-controller/internal/services"
	"github.com/leetsecure/qryptic-controller/internal/utils/logger"
)

// Auth for User and Admin godoc
//
//	@Summary		Auth for User and Admin
//	@ID				auth-for-user-and-admin
//	@Description	get auth token for given username and password
//	@Tags			public
//	@Accept			json
//	@Produce		json
//	@Success		200					{object}	any
//	@Failure		401					{object}	any
//	@Failure		500					{object}	any
//	@Param			UserLoginRequest	body		models.UserLoginRequest	true	"Insert your email and password"
//	@Router			/api/v1/auth/login [post]
func UserAdminLogin(c *gin.Context) {

	var userLoginRequest models.UserLoginRequest

	if err := c.ShouldBindJSON(&userLoginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := services.UserLogin(userLoginRequest.EmailId, userLoginRequest.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"authToken": token})
}

// SSO Config godoc
//
//	@Summary		SSO Configs
//	@ID				sso-configs
//	@Description	get-client-ids-of-sso-allowed
//	@Tags			public
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	any
//	@Failure		401	{object}	any
//	@Failure		500	{object}	any
//	@Router			/api/v1/sso-config [get]
func GetSsoConfiguration(c *gin.Context) {
	status, response, err := services.GetSsoConfiguration()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !status {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sso not enabled"})
		return
	}
	c.JSON(http.StatusOK, response)
}

// InitiateSSOAuth godoc
//
//	@Summary		InitiateSSOAuth
//	@ID				InitiateSSOAuth
//	@Description	InitiateSSOAuth
//	@Tags			public
//	@Accept			json
//	@Produce		json
//	@Success		200			{object}	any
//	@Failure		401			{object}	any
//	@Failure		500			{object}	any
//	@Param			provider	path		string	true	"provider"
//
//	@Router			/api/v1/auth/{provider}/sso/initiate [get]
func InitiateSSOAuth(c *gin.Context) {
	log := logger.Default()
	provider := c.Param("provider")
	err := services.AuthProviderValidate(provider)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err})
		return
	}
	if provider == "google" {
		InitiateGoogleSSOAuth(c)
		return
	}
}

func InitiateGoogleSSOAuth(c *gin.Context) {
	log := logger.Default()
	clientId := c.DefaultQuery("client_id", "")
	platform := c.DefaultQuery("platform", "")
	if platform == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing platform"})
		return
	}
	codeChallenge := c.DefaultQuery("code_challenge", "")
	if codeChallenge == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code challenge"})
		return
	}
	redirectUri := c.DefaultQuery("redirect_uri", "")
	if redirectUri == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing redirect uri"})
		return
	}
	codeChallengeMethod := c.DefaultQuery("code_challenge_method", "S256")
	authURL, err := services.InitiateGoogleSsoAuth(clientId, platform, codeChallenge, redirectUri, codeChallengeMethod)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	c.Redirect(http.StatusFound, authURL)

}

// This is for sso signin with PKCE support
// UserAuthSSOCallback godoc
//
//	@Summary		UserAuthSSOCallback
//	@ID				UserAuthSSOCallback
//	@Description	UserAuthSSOCallback
//	@Tags			public
//	@Accept			json
//	@Produce		json
//	@Success		200			{object}	any
//	@Failure		401			{object}	any
//	@Failure		500			{object}	any
//	@Param			provider	path		string	true	"provider"
//
//	@Router			/api/v1/auth/{provider}/sso/callback [get]
func UserAuthSSOCallback(c *gin.Context) {
	log := logger.Default()
	provider := c.Param("provider")
	err := services.AuthProviderValidate(provider)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "provider not allowed"})
		return
	}
	if provider == "google" {
		GoogleSSOCallback(c)
		return
	}
}

func GoogleSSOCallback(c *gin.Context) {
	log := logger.Default()
	code := c.DefaultQuery("code", "")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code"})
		return
	}
	stateJWT := c.DefaultQuery("state", "")
	if stateJWT == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing state"})
		return
	}
	codeVerifier := c.DefaultQuery("code_verifier", "")
	if codeVerifier == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code verifier"})
		return
	}
	codeChallenge, redirectUrl, err := services.ValidateStateJWT(stateJWT)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid state JWT"})
		return
	}
	// Step 3: Verify the code verifier matches the code challenge
	if !services.VerifyCodeVerifier(codeVerifier, codeChallenge) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid code verifier"})
		return
	}

	// Step 4: Exchange the authorization code for tokens from Google
	tokenResponse, err := services.ExchangeCodeForTokens(code, codeVerifier, redirectUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for tokens"})
		return
	}
	log.Info(tokenResponse)
	userInfo, err := services.GetUserInfoFromGoogle(tokenResponse.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT"})
		return
	}
	log.Info(userInfo)

	// Step 5: Generate a custom JWT for the frontend
	userEmail := userInfo.Email
	authToken, err := services.UserSSOLogin(userEmail)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"authToken": authToken})

}

// This is for mobile devices where we only need to validate the idtoken given by the user
// UserAuthVerifySSOToken godoc
//
//	@Summary		UserAuthVerifySSOToken
//	@ID				UserAuthVerifySSOToken
//	@Description	UserAuthVerifySSOToken
//	@Tags			public
//	@Accept			json
//	@Produce		json
//	@Success		200			{object}	any
//	@Failure		401			{object}	any
//	@Failure		500			{object}	any
//	@Param			provider	path		string	true	"provider"
//
//	@Router			/api/v1/auth/{provider}/sso/token [get]
func UserAuthVerifySSOToken(c *gin.Context) {
	log := logger.Default()
	provider := c.Param("provider")
	err := services.AuthProviderValidate(provider)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err})
		return
	}
	if provider == "google" {
		UserAuthVerifyGoogleSSOToken(c)
		return
	}
}

// This is for mobile devices using google_sign_in package where we only need to validate the google idtoken given by the user
func UserAuthVerifyGoogleSSOToken(c *gin.Context) {
	log := logger.Default()
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing token"})
		return
	}
	emailId, err := services.VerifyGoogleSSOToken(token)
	if err != nil {
		log.Info(err)
		c.JSON(http.StatusUnauthorized, nil)
		return
	}

	userEmail := emailId
	authToken, err := services.UserSSOLogin(userEmail)
	if err != nil {
		c.AbortWithError(http.StatusUnauthorized, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"authToken": authToken})
}

// WebGoogleLoginInitiate godoc
//
//	@Summary		WebGoogleLoginInitiate
//	@ID				WebGoogleLoginInitiate
//	@Description	WebGoogleLoginInitiate
//	@Tags			public
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	any
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			provider		path		string	true	"provider"
//	@Param			code_challenge	query		string	true	"string"	code_challenge(string)
//	@Router			/api/v1/auth/{provider}/web/sso/initiate [get]
func WebGoogleLoginInitiate(c *gin.Context) {
	log := logger.Default()
	provider := c.Param("provider")
	err := services.AuthProviderValidate(provider)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err})
		return
	}
	codeChallenge := c.DefaultQuery("code_challenge", "")
	if codeChallenge == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code challenge"})
		return
	}
	authURL, err := services.WebGoogleLoginInitiate(codeChallenge)
	if err != nil {
		log.Info(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "internal error"})
		return
	}
	c.Redirect(http.StatusFound, authURL)
}

func WebGoogleLoginCallback(c *gin.Context) {
	log := logger.Default()
	provider := c.Param("provider")
	err := services.AuthProviderValidate(provider)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err})
		return
	}
	state := c.DefaultQuery("state", "")
	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing state"})
		return
	}
	code := c.DefaultQuery("code", "")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code"})
		return
	}
	err = services.WebGoogleLoginCallback(state, code)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Authentication Completed. You can close this tab"})
}

// WebGoogleLoginToken godoc
//
//	@Summary		WebGoogleLoginToken
//	@ID				WebGoogleLoginToken
//	@Description	WebGoogleLoginToken
//	@Tags			public
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	any
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			provider		path		string	true	"provider"
//	@Param			code_verifier	query		string	true	"string"	code_verifier(string)
//	@Param			code_challenge	query		string	true	"string"	code_challenge(string)
//	@Router			/api/v1/auth/{provider}/web/sso/token [get]
func WebGoogleLoginToken(c *gin.Context) {
	log := logger.Default()
	provider := c.Param("provider")
	err := services.AuthProviderValidate(provider)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err})
		return
	}
	code_verifier := c.DefaultQuery("code_verifier", "")
	if code_verifier == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code_verifier"})
		return
	}
	code_challenge := c.DefaultQuery("code_challenge", "")
	if code_challenge == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code_challenge"})
		return
	}

	sessionClosed, authToken, err := services.WebGoogleLoginToken(code_verifier, code_challenge)
	if err != nil {
		if sessionClosed {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorised"})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "waiting for authentication to complete"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"authToken": authToken})
}
