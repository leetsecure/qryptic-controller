package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leetsecure/qryptic-controller/internal/models"
	"github.com/leetsecure/qryptic-controller/internal/services"
	"github.com/leetsecure/qryptic-controller/internal/utils/logger"
)

// CreateGroup godoc
//
//	@Summary		CreateGroup
//	@Description	CreateGroup
//	@Tags			admin-group
//	@Failure		400					{object}	any
//	@Failure		401					{object}	any
//	@Failure		500					{object}	any
//	@Param			Authorization		header		string						true	"Insert your token"	default(Bearer <token>)
//	@Param			GroupCreateRequest	body		models.GroupCreateRequest	true	"Group details"
//	@Router			/api/v1/admin/group [post]
func CreateGroup(c *gin.Context) {
	var groupCreateRequest models.GroupCreateRequest
	if err := c.ShouldBindJSON(&groupCreateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := services.CreateGroup(groupCreateRequest.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true})
}

// DeleteGroup godoc
//
//	@Summary		DeleteGroup
//	@Description	DeleteGroup
//	@Tags			admin-group
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	any
//	@Failure		400				{object}	any
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			Authorization	header		string	true	"Insert your token"	default(Bearer <token>)
//	@Param			id				path		string	true	"group id"
//	@Router			/api/v1/admin/group/{id} [delete]
func DeleteGroup(c *gin.Context) {
	groupUuid := c.Param("id")
	err := services.DeleteVpnGateway(groupUuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// UpdateGroup godoc
//
//	@Summary		UpdateGroup
//	@Description	UpdateGroup
//	@Tags			admin-group
//	@Accept			json
//	@Produce		json
//	@Success		200					{object}	any
//	@Failure		400					{object}	any
//	@Failure		401					{object}	any
//	@Failure		500					{object}	any
//	@Param			Authorization		header		string						true	"Insert your token"	default(Bearer <token>)
//	@Param			GroupUpdateRequest	body		models.GroupUpdateRequest	true	"Group details"
//	@Param			id					path		string						true	"group id"
//
//	@Router			/api/v1/admin/group/{id} [put]
func UpdateGroup(c *gin.Context) {
	var groupUpdateRequest models.GroupUpdateRequest
	if err := c.ShouldBindJSON(&groupUpdateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	groupUuid := c.Param("id")
	err := services.UpdateGroup(groupUuid,
		groupUpdateRequest.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AddRemoveUsersInGroup godoc
//
//	@Summary		AddRemoveUsersInGroup
//	@Description	AddRemoveUsersInGroup
//	@Tags			admin-group
//	@Accept			json
//	@Produce		json
//	@Success		200						{object}	any
//	@Failure		400						{object}	any
//	@Failure		401						{object}	any
//	@Failure		500						{object}	any
//	@Param			Authorization			header		string							true	"Insert your token"	default(Bearer <token>)
//	@Param			GroupUpdateUserRequest	body		models.GroupUpdateUserRequest	true	"Provide details"
//	@Param			id						path		string							true	"group id"
//	@Param			action					path		string							true	"action"
//
//	@Router			/api/v1/admin/group/{id}/{action}/users [put]
func AddRemoveUsersInGroup(c *gin.Context) {
	log := logger.Default()
	var groupUpdateUserRequest models.GroupUpdateUserRequest
	if err := c.ShouldBindJSON(&groupUpdateUserRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	groupUuid := c.Param("id")
	action := c.Param("action")
	if action != "add" && action != "remove" {
		log.Infof("invalid action : %s", action)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action"})
		return
	}
	err := services.AddRemoveUsersInGroup(action,
		groupUuid,
		groupUpdateUserRequest.UserUuids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ListGroups godoc
//
//	@Summary		ListGroups
//	@ID				ListGroups
//	@Description	list all the groups
//	@Tags			admin-group
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.Group
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			Authorization	header		string		true	"Insert your token"		default(Bearer <token>)
//	@Param			include			query		[]string	false	"string collections"	collectionFormat(multi)
//
//	@Router			/api/v1/admin/group/list [get]
func ListGroups(c *gin.Context) {
	includeQueryParams := c.QueryArray("include")
	includeUsers := false
	includeGateways := false
	for _, include := range includeQueryParams {
		if include == "users" {
			includeUsers = true
		} else if include == "gateways" {
			includeGateways = true
		}
	}
	groups, err := services.ListGroups(includeUsers, includeGateways)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// GetGroupDetails godoc
//
//	@Summary		Get Group
//	@ID				GetGroup
//	@Description	Get Group By UUID
//	@Tags			admin-group
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.Group
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			Authorization	header		string		true	"Insert your token"		default(Bearer <token>)
//	@Param			include			query		[]string	false	"string collections"	collectionFormat(multi)
//	@Param			id				path		string		true	"group id"
//
//	@Router			/api/v1/admin/group/{id} [get]
func GetGroupByUUID(c *gin.Context) {
	groupUuid := c.Param("id")
	includeQueryParams := c.QueryArray("include")
	includeUsers := false
	includeGateways := false
	for _, include := range includeQueryParams {
		if include == "users" {
			includeUsers = true
		} else if include == "gateways" {
			includeGateways = true
		}
	}

	group, err := services.GetGroupByUUID(groupUuid, includeUsers, includeGateways)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, group)
}
