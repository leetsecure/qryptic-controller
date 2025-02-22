package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leetsecure/qryptic-controller/internal/models"
	"github.com/leetsecure/qryptic-controller/internal/services"
)

// User Management

// RegisterUser godoc
//
//	@Summary		RegisterUser
//	@Description	RegisterUser
//	@Tags			admin-user
//	@Accept			json
//	@Produce		json
//	@Success		200					{object}	any
//	@Failure		400					{object}	any
//	@Failure		401					{object}	any
//	@Failure		500					{object}	any
//	@Param			Authorization		header		string						true	"Insert your token"	default(Bearer <token>)
//	@Param			RegisterUserRequest	body		models.RegisterUserRequest	true	"User details"
//	@Router			/api/v1/admin/user [post]
func RegisterUser(c *gin.Context) {
	var registerUserRequest models.RegisterUserRequest

	if err := c.ShouldBindJSON(&registerUserRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	passwd := registerUserRequest.Password
	isPasswordSet := *(registerUserRequest.IsPasswordSet)
	if isPasswordSet && (len(passwd) < 8 || len(passwd) > 50) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password length should be between 8 and 50"})
		return
	}

	if !isPasswordSet && len(passwd) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password found in request even if isPasswordSet is false"})
		return
	}

	err := services.RegisterUser(registerUserRequest.EmailId, registerUserRequest.Password, string(registerUserRequest.Role), isPasswordSet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "User Created"})

}

// DeleteUser godoc
//
//	@Summary		DeleteUser
//	@Description	DeleteUser
//	@Tags			admin-user
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	any
//	@Failure		400				{object}	any
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			Authorization	header		string	true	"Insert your token"	default(Bearer <token>)
//
//	@Param			id				path		string	true	"user id"
//
//	@Router			/api/v1/admin/user/{id} [delete]
func DeleteUser(c *gin.Context) {
	userUuid := c.Param("id")

	err := services.DeleteUser(userUuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true})

}

// UpdateUser godoc
//
//	@Summary		UpdateUser
//	@Description	UpdateUser
//	@Tags			admin-user
//	@Accept			json
//	@Produce		json
//	@Success		200					{object}	any
//	@Failure		400					{object}	any
//	@Failure		401					{object}	any
//	@Failure		500					{object}	any
//	@Param			Authorization		header		string						true	"Insert your token"	default(Bearer <token>)
//	@Param			UpdateUserRequest	body		models.UpdateUserRequest	true	"User details"
//
//	@Param			id					path		string						true	"user id"
//
//	@Router			/api/v1/admin/user/{id} [put]
func UpdateUser(c *gin.Context) {
	var updateUserRequest models.UpdateUserRequest
	if err := c.ShouldBindJSON(&updateUserRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userUuid := c.Param("id")

	passwd := updateUserRequest.NewPassword
	isPasswordSet := *(updateUserRequest.IsPasswordSet)
	if isPasswordSet && (len(passwd) < 8 || len(passwd) > 50) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password length should be between 8 and 50"})
		return
	}

	if !isPasswordSet && len(passwd) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password found in request even if isPasswordSet is false"})
		return
	}

	err := services.UpdateUser(userUuid, updateUserRequest.EmailId, updateUserRequest.NewPassword, string(updateUserRequest.Role), isPasswordSet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true})
}

// ListUsers godoc
//
//	@Summary		ListUsers
//	@Description	ListUsers
//	@Tags			admin-user
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.User
//	@Failure		400				{object}	any
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			Authorization	header		string		true	"Insert your token"		default(Bearer <token>)
//	@Param			include			query		[]string	false	"string collections"	collectionFormat(multi)
//	@Router			/api/v1/admin/user/list [get]
func ListUsers(c *gin.Context) {
	includeQueryParams := c.QueryArray("include")
	includeGatewaysWithClients := false
	includeGateways := false
	includeGroups := false
	includeClients := false
	for _, include := range includeQueryParams {
		if include == "gateways-clients" {
			includeGatewaysWithClients = true
		} else if include == "gateways" {
			includeGateways = true
		} else if include == "groups" {
			includeGroups = true
		} else if include == "clients" {
			includeClients = true
		}
	}

	users, err := services.ListUsers(includeGateways, includeClients, includeGatewaysWithClients, includeGroups)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetUserDetails godoc
//
//	@Summary		GetUserDetails
//	@Description	GetUserDetails
//	@Tags			admin-user
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	models.User
//	@Failure		400				{object}	any
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			Authorization	header		string		true	"Insert your token"		default(Bearer <token>)
//	@Param			include			query		[]string	false	"string collections"	collectionFormat(multi)
//
//	@Param			id				path		string		true	"user id"
//
//	@Router			/api/v1/admin/user/{id} [get]
func GetUserByUUID(c *gin.Context) {
	includeQueryParams := c.QueryArray("include")
	includeGateways := false
	includeClients := false
	includeGatewaysWithClients := false
	includeGroups := false
	for _, include := range includeQueryParams {
		if include == "gateways-clients" {
			includeGatewaysWithClients = true
		} else if include == "gateways" {
			includeGateways = true
		} else if include == "groups" {
			includeGroups = true
		} else if include == "clients" {
			includeClients = true
		}
	}

	userUuid := c.Param("id")

	userDetail, err := services.GetUserByUUID(userUuid, includeGateways, includeClients, includeGatewaysWithClients, includeGroups)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, userDetail)
}
