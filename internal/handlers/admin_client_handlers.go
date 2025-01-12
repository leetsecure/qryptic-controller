package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leetsecure/qryptic-controller/internal/services"
)

// DeleteVpnClient godoc
//
//	@Summary		DeleteVpnClient
//	@Description	DeleteVpnClient
//	@Tags			admin-client
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
//	@Router			/api/v1/admin/client/{id} [delete]
func DeleteVpnClientByAdmin(c *gin.Context) {
	clientUuid := c.Param("id")
	err := services.DeleteClientFromUserAndVpnGateway(clientUuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteExpiredClients godoc
//
//	@Summary		DeleteExpiredClients
//	@Description	DeleteExpiredClients
//	@Tags			admin-client
//	@Accept			json
//	@Produce		json
//	@Success		200				{object}	any
//	@Failure		400				{object}	any
//	@Failure		401				{object}	any
//	@Failure		500				{object}	any
//	@Param			Authorization	header		string	true	"Insert your token"	default(Bearer <token>)
//	@Router			/api/v1/admin/client/expired [delete]
func DeleteExpiredClients(c *gin.Context) {
	err := services.DeleteExpiredClientsFromUserAndVpnGateway()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
