package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leetsecure/qryptic-controller/internal/models"
	"github.com/leetsecure/qryptic-controller/internal/services"
)

// ListGateways godoc
//
//	@Summary		List gateways
//	@ID				list-gateways
//	@Description	list all the gateways
//	@Tags			admin-gateway
//	@Accept			json
//	@Produce		json
//	@Success		200				{array}		models.VpnGateway
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			Authorization	header		string		true	"Insert your token"		default(Bearer <token>)
//	@Param			include			query		[]string	false	"string collections"	collectionFormat(multi)
//	@Router			/api/v1/admin/gateway/list [get]
func ListVpnGateways(c *gin.Context) {
	includeQueryParams := c.QueryArray("include")
	includeUsers := false
	includeClients := false
	includeUsersWithClients := false
	includeGroups := false
	includeIpPool := false
	for _, include := range includeQueryParams {
		if include == "users-clients" {
			includeUsersWithClients = true
		} else if include == "users" {
			includeUsers = true
		} else if include == "groups" {
			includeGroups = true
		} else if include == "clients" {
			includeClients = true
		} else if include == "ipPool" {
			includeIpPool = true
		}
	}

	vpnGateways, err := services.ListVpnGateways(includeUsers, includeClients, includeUsersWithClients, includeGroups, includeIpPool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vpnGateways)
}

// CreateGateway godoc
//
//	@Summary		CreateGateway
//	@Description	CreateGateway
//	@Tags			admin-gateway
//	@Accept			json
//	@Produce		json
//	@Success		200						{object}	any
//	@Failure		400						{object}	any
//	@Failure		401						{object}	any
//	@Failure		500						{object}	any
//	@Param			Authorization			header		string							true	"Insert your token"	default(Bearer <token>)
//	@Param			VpnGatewayCreateRequest	body		models.VpnGatewayCreateRequest	true	"Gateway details"
//	@Router			/api/v1/admin/gateway [post]
func CreateVpnGateway(c *gin.Context) {
	var vpnGatewayCreateRequest models.VpnGatewayCreateRequest
	if err := c.ShouldBindJSON(&vpnGatewayCreateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := services.CreateVpnGateway(vpnGatewayCreateRequest.Name,
		vpnGatewayCreateRequest.Domain,
		vpnGatewayCreateRequest.IpAddress,
		vpnGatewayCreateRequest.VpnCIDR,
		vpnGatewayCreateRequest.Port,
		vpnGatewayCreateRequest.DnsServer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true})
}

// DeleteGateway godoc
//
//	@Summary		DeleteGateway
//	@Description	DeleteGateway
//	@Tags			admin-gateway
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	any
//	@Failure		400				{object}	any
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			Authorization	header		string	true	"Insert your token"	default(Bearer <token>)
//
//	@Param			id				path		string	true	"gateway id"
//	@Router			/api/v1/admin/gateway/{id} [delete]
func DeleteVpnGateway(c *gin.Context) {

	gatewayUuid := c.Param("id")
	err := services.DeleteVpnGateway(gatewayUuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})

}

// UpdateGateway godoc
//
//	@Summary		UpdateGateway
//	@Description	UpdateGateway
//	@Tags			admin-gateway
//	@Accept			json
//	@Produce		json
//	@Success		200						{object}	any
//	@Failure		400						{object}	any
//	@Failure		401						{object}	any
//	@Failure		500						{object}	any
//	@Param			Authorization			header		string							true	"Insert your token"	default(Bearer <token>)
//	@Param			VpnGatewayUpdateRequest	body		models.VpnGatewayUpdateRequest	true	"Gateway details"
//	@Param			id						path		string							true	"gateway id"
//	@Router			/api/v1/admin/gateway/{id} [put]
func UpdateVpnGateway(c *gin.Context) {
	var vpnGatewayUpdateRequest models.VpnGatewayUpdateRequest
	if err := c.ShouldBindJSON(&vpnGatewayUpdateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	gatewayUuid := c.Param("id")
	err := services.UpdateVpnGateway(gatewayUuid,
		vpnGatewayUpdateRequest.Name,
		vpnGatewayUpdateRequest.Domain,
		vpnGatewayUpdateRequest.IpAddress,
		vpnGatewayUpdateRequest.Port,
		vpnGatewayUpdateRequest.DnsServer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetVpnGatewayDeploymentConfig godoc
//
//	@Summary		GetVpnGatewayDeploymentConfig
//	@Description	GetVpnGatewayDeploymentConfig
//	@Tags			admin-gateway
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	any
//	@Failure		400				{object}	any
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			Authorization	header		string	true	"Insert your token"	default(Bearer <token>)
//	@Param			id				path		string	true	"gateway id"
//
//	@Router			/api/v1/admin/gateway/{id}/deployment-config [get]
func GetVpnGatewayDeploymentConfig(c *gin.Context) {
	gatewayUuid := c.Param("id")
	deploymentConfig, err := services.CreateVpnGatewayDeploymentConfig(gatewayUuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deploymentConfig": deploymentConfig})
}

// GetVpnGatewayDetails godoc
//
//	@Summary		GetVpnGatewayDetails
//	@Description	GetVpnGatewayDetails
//	@Tags			admin-gateway
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	models.VpnGateway
//	@Failure		400				{object}	any
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			Authorization	header		string		true	"Insert your token"	default(Bearer <token>)
//	@Param			id				path		string		true	"gateway id"
//	@Param			include			query		[]string	false	"string collections"	collectionFormat(multi)
//
//	@Router			/api/v1/admin/gateway/{id} [get]
func GetGatewayByUUID(c *gin.Context) {
	gatewayUuid := c.Param("id")
	includeQueryParams := c.QueryArray("include")
	includeUsers := false
	includeClients := false
	includeUsersWithClients := false
	includeGroups := false
	includeIpPool := false
	for _, include := range includeQueryParams {
		if include == "users-clients" {
			includeUsersWithClients = true
		} else if include == "users" {
			includeUsers = true
		} else if include == "groups" {
			includeGroups = true
		} else if include == "clients" {
			includeClients = true
		} else if include == "ipPool" {
			includeIpPool = true
		}
	}

	vpnGatewayDetail, err := services.GetVpnGatewayByUUID(gatewayUuid, includeUsers, includeClients, includeUsersWithClients, includeGroups, includeIpPool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vpnGatewayDetail)
}

// Reset Gateway godoc
//
//	@Summary		Reset Gateway
//	@Description	Clear Gateway Clients And IPPool
//	@Tags			admin-gateway
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	any
//	@Failure		400				{object}	any
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			Authorization	header		string	true	"Insert your token"	default(Bearer <token>)
//	@Param			id				path		string	true	"gateway id"
//
//	@Router			/api/v1/admin/gateway/{id}/reset [delete]
func ClearVpnGatewayClientsAndIPPool(c *gin.Context) {
	gatewayUuid := c.Param("id")
	err := services.ClearVpnGatewayClientsAndIPPool(gatewayUuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
