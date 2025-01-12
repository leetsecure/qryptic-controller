package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/leetsecure/qryptic-controller/internal/models"
	"github.com/leetsecure/qryptic-controller/internal/services"
)

// GetVpnGatewayWGConfigByGW godoc
//
//	@Summary		GetVpnGatewayWGConfigByGW
//	@Description	GetVpnGatewayWGConfigByGW
//	@Tags			gateway
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	models.WGServerConfig
//	@Failure		400				{object}	any
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			Authorization	header		string	true	"Insert your token"	default(Bearer <token>)
//	@Router			/api/v1/gateway/get-gateway-config [get]
func GetVpnGatewayWGConfigByGW(c *gin.Context) {
	vpnGatewayUuid, _ := c.Get("vpnGatewayUuid")
	vpnGatewayConfig, err := services.GetVpnGatewayWGConfig(vpnGatewayUuid.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, vpnGatewayConfig)
}
