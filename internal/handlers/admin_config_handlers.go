package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leetsecure/qryptic-controller/internal/models"
	"github.com/leetsecure/qryptic-controller/internal/services"
)

// AddSsoConfig godoc
//
//	@Summary		Add SSO Config
//	@ID				AddSsoConfig
//	@Description	Add SSO Configuration
//	@Tags			admin-config
//	@Accept			json
//	@Produce		json
//	@Success		200					{object}	any
//	@Failure		401					{object}	any
//	@Failure		500					{object}	any
//	@Param			Authorization		header		string						true	"Insert your token"	default(Bearer <token>)
//	@Param			AddSsoConfigRequest	body		models.AddSsoConfigRequest	true	"SSO Config Details"
//	@Router			/api/v1/admin/config/sso [post]
func AddSsoConfig(c *gin.Context) {
	var addSsoConfigRequest models.AddSsoConfigRequest
	if err := c.ShouldBindJSON(&addSsoConfigRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := services.AddSsoConfig(addSsoConfigRequest.Domain,
		addSsoConfigRequest.Provider,
		addSsoConfigRequest.ClientID,
		addSsoConfigRequest.ClientSecret, addSsoConfigRequest.Platform)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteSsoConfig godoc
//
//	@Summary		DeleteSsoConfig
//	@ID				DeleteSsoConfig
//	@Description	delete sso configuration
//	@Tags			admin-config
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	any
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			Authorization	header		string	true	"Insert your token"	default(Bearer <token>)
//	@Param			id				path		string	true	"sso id"
//	@Router			/api/v1/admin/config/sso/{id} [delete]
func DeleteSsoConfig(c *gin.Context) {
	// var deleteSsoConfigRequest models.DeleteSsoConfigRequest
	// if err := c.ShouldBindJSON(&deleteSsoConfigRequest); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }
	ssoConfigUuid := c.Param("id")
	err := services.DeleteSsoConfig(ssoConfigUuid)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// UpdateAllowPasswordLogin godoc
//
//	@Summary		UpdateAllowPasswordLogin
//	@ID				UpdateAllowPasswordLogin
//	@Description	update whether password login should be allowed or not
//	@Tags			admin-config
//	@Accept			json
//	@Produce		json
//	@Success		200								{object}	any
//	@Failure		401								{object}	any
//	@Failure		500								{object}	any
//	@Param			Authorization					header		string									true	"Insert your token"	default(Bearer <token>)
//	@Param			UpdateAllowPasswordLoginRequest	body		models.UpdateAllowPasswordLoginRequest	true	"true or false"
//	@Router			/api/v1/admin/config/password-login [put]
func UpdateAllowPasswordLogin(c *gin.Context) {
	var updateAllowPasswordLoginRequest models.UpdateAllowPasswordLoginRequest
	if err := c.ShouldBindJSON(&updateAllowPasswordLoginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := services.UpdateAllowPasswordLogin(*updateAllowPasswordLoginRequest.AllowPasswordLogin)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// UpdateAllowSSOLogin godoc
//
//	@Summary		UpdateAllowSSOLogin
//	@ID				UpdateAllowSSOLogin
//	@Description	update whether sso login should be allowed or not
//	@Tags			admin-config
//	@Accept			json
//	@Produce		json
//	@Success		200							{object}	any
//	@Failure		401							{object}	any
//	@Failure		500							{object}	any
//	@Param			Authorization				header		string								true	"Insert your token"	default(Bearer <token>)
//	@Param			UpdateAllowSSOLoginRequest	body		models.UpdateAllowSSOLoginRequest	true	"true or false"
//	@Router			/api/v1/admin/config/sso-login [put]
func UpdateAllowSSOLogin(c *gin.Context) {
	var updateAllowSSOLoginRequest models.UpdateAllowSSOLoginRequest
	if err := c.ShouldBindJSON(&updateAllowSSOLoginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := services.UpdateAllowSSOLogin(*updateAllowSSOLoginRequest.AllowSsoLogin)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetAdminConfiguration godoc
//
//	@Summary		GetAdminConfiguration
//	@ID				GetAdminConfiguration
//	@Description	Get Admin Configuration
//	@Tags			admin-config
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	any
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			Authorization	header		string		true	"Insert your token"		default(Bearer <token>)
//	@Param			include			query		[]string	false	"string collections"	collectionFormat(multi)
//
//	@Router			/api/v1/admin/config [get]
func GetAdminConfiguration(c *gin.Context) {
	includeQueryParams := c.QueryArray("include")
	includeSsoConfigs := false
	for _, include := range includeQueryParams {
		if include == "sso-configs" {
			includeSsoConfigs = true
		}
	}
	adminConfiguration, err := services.GetAdminConfiguration(includeSsoConfigs)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, adminConfiguration)
}
