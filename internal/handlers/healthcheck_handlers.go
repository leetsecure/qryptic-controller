package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leetsecure/qryptic-controller/internal/services"
)

// Controller Health Check godoc
//
//	@Summary		Controller Health Check
//	@ID				controller-health-check
//	@Description	get health of the controller
//	@Tags			public
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	any
//	@Router			/api/v1/health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Gateway Health Check godoc
//
//	@Summary		Gateway Health Check
//	@ID				gateway-health-check
//	@Description	get health of the gateway by vpnGatewayUuid
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	any
//	@Failure		400				{object}	any
//	@Failure		401				{object}	any
//	@Failure		404				{object}	any
//	@Failure		500				{object}	any
//
//	@Param			Authorization	header		string	true	"Insert your token"	default(Bearer <token>)
//	@Param			id				path		string	true	"gateway id"
//	@Router			/api/v1/gateway/{id}/health [get]
func VpnGatewayHealthCheck(c *gin.Context) {
	gatewayUuid := c.Param("id")
	healthCheckStatus, err := services.VpnGatewayHealthCheck(gatewayUuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !healthCheckStatus {
		c.JSON(http.StatusNotFound, gin.H{"success": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
