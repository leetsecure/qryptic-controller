package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leetsecure/qryptic-controller/internal/models"
	"github.com/leetsecure/qryptic-controller/internal/services"
	"github.com/leetsecure/qryptic-controller/internal/utils/logger"
)

// UpdateUsersInGateway godoc
//
//	@Summary		AddRemoveUsersInGateway
//	@Description	AddRemoveUsersInGateway
//	@Tags			admin-access
//	@Accept			json
//	@Produce		json
//	@Success		200							{object}	any
//	@Failure		400							{object}	any
//	@Failure		401							{object}	any
//	@Failure		500							{object}	any
//	@Param			Authorization				header		string								true	"Insert your token"	default(Bearer <token>)
//	@Param			id							path		string								true	"gateway id"
//	@Param			action						path		string								true	"action"
//	@Param			VpnGatewayUpdateUserRequest	body		models.VpnGatewayUpdateUserRequest	true	"user ids"
//	@Router			/api/v1/admin/access/gateway/{id}/{action}/users [put]
func AddRemoveUsersInVpnGateway(c *gin.Context) {
	gatewayUuid := c.Param("id")
	action := c.Param("action")
	if action != "add" && action != "remove" {
		log := logger.Default()
		log.Infof("invalid action : %s", action)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action"})
		return
	}
	var vpnGatewayUpdateUserRequest models.VpnGatewayUpdateUserRequest
	if err := c.ShouldBindJSON(&vpnGatewayUpdateUserRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := services.AddRemoveUsersInVpnGateway(action,
		gatewayUuid,
		vpnGatewayUpdateUserRequest.UserUuids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// UpdateGroupsInGateway godoc
//
//	@Summary		AddRemoveGroupsInGateway
//	@Description	AddRemoveGroupsInGateway
//	@Tags			admin-access
//	@Accept			json
//	@Produce		json
//	@Success		200							{object}	any
//	@Failure		400							{object}	any
//	@Failure		401							{object}	any
//	@Failure		500							{object}	any
//	@Param			Authorization				header		string								true	"Insert your token"	default(Bearer <token>)
//	@Param			id							path		string								true	"gateway id"
//	@Param			action						path		string								true	"action"
//	@Param			GatewayUpdateGroupRequest	body		models.GatewayUpdateGroupRequest	true	"group ids"
//
//	@Router			/api/v1/admin/access/gateway/{id}/{action}/groups [put]
func AddRemoveGroupsInVpnGateway(c *gin.Context) {
	log := logger.Default()
	gatewayUuid := c.Param("id")
	action := c.Param("action")
	if action != "add" && action != "remove" {
		log.Infof("invalid action : %s", action)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action"})
		return
	}
	var gatewayUpdateGroupRequest models.GatewayUpdateGroupRequest
	if err := c.ShouldBindJSON(&gatewayUpdateGroupRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := services.AddRemoveGroupsInVpnGateway(action,
		gatewayUuid,
		gatewayUpdateGroupRequest.GroupUuids)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ListGatewaysAccessibleByUser godoc
//
//	@Summary		ListGatewaysAccessibleByUser
//	@Description	ListGatewaysAccessibleByUser
//	@Tags			admin-access
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.VpnGatewayUserResponse
//	@Failure		400				{object}	any
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			Authorization	header		string	true	"Insert your token"	default(Bearer <token>)
//
//	@Param			id				path		string	true	"user id"
//	@Param			action			path		string	true	"action"
//
//	@Router			/api/v1/admin/access/user/{id}/gateways [get]
func ListVpnGatewaysAccessibleByUser(c *gin.Context) {
	userUuid := c.Param("id")
	vpnGateways, err := services.ListAccessibleVPNsV2(userUuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vpnGateways)
}

// ListGatewaysAccessibleByGroup godoc
//
//	@Summary		ListGatewaysAccessibleByGroup
//	@Description	ListGatewaysAccessibleByGroup
//	@Tags			admin-access
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.VpnGatewayUserResponse
//	@Failure		400				{object}	any
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			Authorization	header		string	true	"Insert your token"	default(Bearer <token>)
//
//	@Param			id				path		string	true	"group id"
//	@Param			action			path		string	true	"action"
//
//	@Router			/api/v1/admin/access/group/{id}/gateways [get]
func ListGatewaysAccessibleByGroup(c *gin.Context) {
	groupUuid := c.Param("id")
	vpnGateways, err := services.ListAccessibleGatewaysByGroup(groupUuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vpnGateways)
}
