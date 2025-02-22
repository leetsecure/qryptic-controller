package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leetsecure/qryptic-controller/internal/database"
	"github.com/leetsecure/qryptic-controller/internal/models"
	"github.com/leetsecure/qryptic-controller/internal/utils/auth"
	"github.com/leetsecure/qryptic-controller/internal/utils/logger"
)

func ControllerAuthCheckMiddleware(c *gin.Context) {
	log := logger.Default()
	Bearer_Schema := "Bearer "
	authorisation := c.GetHeader("Authorization")

	if len(authorisation) <= len(Bearer_Schema)+1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorised"})
		c.Abort()
		return
	}

	token := authorisation[len(Bearer_Schema):]

	userUuid, userRole, err := auth.VerifyUserAuthToken(token)
	log.Info(userUuid, userRole)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		c.Abort()
		return
	}
	c.Set("userRole", userRole)
	c.Set("userUuid", userUuid)
	c.Next()
}

func AdminRoleCheckMiddleware(c *gin.Context) {
	userRole, exists := c.Get("userRole")
	if !exists || userRole != models.AdminRole {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not Admin"})
		c.Abort()
		return
	}
	c.Next()
}

func VpnGatewayAuthCheckMiddleware(c *gin.Context) {
	vpnGatewayUuid := c.GetHeader("VPN-Gateway-UUID")
	Bearer_Schema := "Bearer "
	authorisation := c.GetHeader("Authorization")

	if len(authorisation) <= len(Bearer_Schema)+1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorised"})
		c.Abort()
		return
	}
	token := authorisation[len(Bearer_Schema):]

	if len(vpnGatewayUuid) == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorised"})
		c.Abort()
		return
	}
	var vpnGateway models.VpnGateway
	err := database.DB.Where("uuid = ?", vpnGatewayUuid).First(&vpnGateway).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		c.Abort()
		return
	}
	jwtSecret := vpnGateway.JwtSecretKey
	_, err = auth.VerifyVpnGatewayAuthToken(token, jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		c.Abort()
		return
	}
	c.Set("vpnGatewayUuid", vpnGatewayUuid)
	c.Next()
}
