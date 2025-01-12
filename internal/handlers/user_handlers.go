package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/leetsecure/qryptic-controller/internal/models"

	"github.com/leetsecure/qryptic-controller/internal/services"
)

// GetVpnGatewaysAccessibleByUser godoc
//
//	@Summary		GetVpnGatewaysAccessibleByUser
//	@Description	GetVpnGatewaysAccessibleByUser
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.VpnGatewayUserResponse
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			Authorization	header		string	true	"Insert your token"	default(Bearer <token>)
//	@Router			/api/v1/gateway/list [get]
func GetVpnGatewaysAccessibleByUser(c *gin.Context) {
	userUuid, _ := c.Get("userUuid")
	vpnGateways, err := services.ListAccessibleVPNsV2(userUuid.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vpnGateways)
}

// GetVpnClientConfig godoc
//
//	@Summary		GetVpnClientConfig
//	@Description	GetVpnClientConfig
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	models.WGClientConfig
//	@Failure		400				{object}	any
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			Authorization	header		string	true	"Insert your token"	default(Bearer <token>)
//	@Param			id				path		string	true	"gateway id"
//	@Router			/api/v1/gateway/{id}/client [get]
func GetVpnClientConfig(c *gin.Context) {
	userUuid, _ := c.Get("userUuid")
	gatewayUuid := c.Param("id")
	vpnClientConfig, status, err := services.CreateVpnGatewayUserClient(userUuid.(string), gatewayUuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !status {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}
	c.JSON(http.StatusOK, vpnClientConfig)
}

// DeleteVpnClient godoc
//
//	@Summary		DeleteVpnClient
//	@Description	DeleteVpnClient
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	any
//	@Failure		400				{object}	any
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			Authorization	header		string	true	"Insert your token"	default(Bearer <token>)
//
//	@Param			id				path		string	true	"client id"
//
//	@Router			/api/v1/client/{id} [delete]
func DeleteVpnClient(c *gin.Context) {
	userUuid, _ := c.Get("userUuid")
	clientUuid := c.Param("id")
	err := services.DeleteClientFromUserAndVpnGatewayByUser(clientUuid, userUuid.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
