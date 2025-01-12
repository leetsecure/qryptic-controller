package models

type AddSsoConfigRequest struct {
	Domain       string `json:"domain" binding:"required"`
	Platform     string `json:"platform" binding:"required"`
	Provider     string `json:"provider" binding:"required"`
	ClientID     string `json:"clientID" binding:"required"`
	ClientSecret string `json:"clientSecret" binding:"required"`
}

type UpdateAllowPasswordLoginRequest struct {
	AllowPasswordLogin *bool `json:"allowPasswordLogin" binding:"required"`
}

type UpdateAllowSSOLoginRequest struct {
	AllowSsoLogin *bool `json:"allowSsoLogin" binding:"required"`
}
