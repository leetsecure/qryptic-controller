package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/leetsecure/qryptic-controller/internal/handlers"
	"github.com/leetsecure/qryptic-controller/internal/middlewares"
)

func SetupControllerRoutes(r *gin.Engine) {

	publicGroup := r.Group("/api/v1")
	{
		publicGroup.GET("/health", handlers.HealthCheck)
		publicGroup.GET("/sso-config", handlers.GetSsoConfiguration)
	}

	authGroup := r.Group("/api/v1/auth")
	{
		authGroup.POST("/login", handlers.UserAdminLogin)
		authGroup.GET("/:provider/sso/initiate", handlers.InitiateSSOAuth)
		authGroup.GET("/:provider/sso/callback", handlers.UserAuthSSOCallback)
		authGroup.GET("/:provider/sso/token", handlers.UserAuthVerifySSOToken)
		authGroup.GET("/:provider/web/sso/initiate", handlers.WebGoogleLoginInitiate)
		authGroup.GET("/:provider/web/sso/callback", handlers.WebGoogleLoginCallback)
		authGroup.GET("/:provider/web/sso/token", handlers.WebGoogleLoginToken)
	}

	gatewayGroup := r.Group("/api/v1/gateway")
	{
		gatewayGroup.GET("/get-gateway-config", middlewares.VpnGatewayAuthCheckMiddleware, handlers.GetVpnGatewayWGConfigByGW)
	}

	adminAccessGroup := r.Group("/api/v1/admin/access")
	{
		adminAccessGroup.PUT("/gateway/:id/:action/users", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.AddRemoveUsersInVpnGateway)
		adminAccessGroup.PUT("/gateway/:id/:action/groups", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.AddRemoveGroupsInVpnGateway)
		adminAccessGroup.GET("/user/:id/gateways", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.ListVpnGatewaysAccessibleByUser)
		adminAccessGroup.GET("/group/:id/gateways", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.ListGatewaysAccessibleByGroup)

	}

	adminClientGroup := r.Group("/api/v1/admin/client")
	{
		adminClientGroup.DELETE("/:id", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.DeleteVpnClientByAdmin)
		adminClientGroup.DELETE("/expired", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.DeleteExpiredClients)
	}

	adminConfigGroup := r.Group("/api/v1/admin/config")
	{
		adminConfigGroup.POST("/sso", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.AddSsoConfig)
		adminConfigGroup.DELETE("/sso/:id", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.DeleteSsoConfig)
		adminConfigGroup.PUT("/password-login", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.UpdateAllowPasswordLogin)
		adminConfigGroup.PUT("/sso-login", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.UpdateAllowSSOLogin)
		adminConfigGroup.GET("/", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.GetAdminConfiguration)

	}

	adminGatewayGroup := r.Group("/api/v1/admin/gateway")
	{
		adminGatewayGroup.POST("/", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.CreateVpnGateway)
		adminGatewayGroup.DELETE("/:id", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.DeleteVpnGateway)
		adminGatewayGroup.PUT("/:id", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.UpdateVpnGateway)
		adminGatewayGroup.GET("/list", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.ListVpnGateways)
		adminGatewayGroup.GET("/:id/deployment-config", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.GetVpnGatewayDeploymentConfig)
		adminGatewayGroup.GET("/:id", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.GetGatewayByUUID)
		adminGatewayGroup.DELETE("/:id/reset", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.ClearVpnGatewayClientsAndIPPool)

	}

	adminGroupGroup := r.Group("/api/v1/admin/group")
	{
		adminGroupGroup.GET("/list", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.ListGroups)
		adminGroupGroup.POST("/", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.CreateGroup)
		adminGroupGroup.DELETE("/:id", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.DeleteGroup)
		adminGroupGroup.PUT("/:id", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.UpdateGroup)
		adminGroupGroup.PUT("/:id/:action/users", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.AddRemoveUsersInGroup)
		adminGroupGroup.GET("/:id", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.GetGroupByUUID)

	}

	adminUserGroup := r.Group("/api/v1/admin/user")
	{
		adminUserGroup.POST("/", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.RegisterUser)
		adminUserGroup.PUT("/:id", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.UpdateUser)
		adminUserGroup.DELETE("/:id", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.DeleteUser)
		adminUserGroup.GET("/list", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.ListUsers)
		adminUserGroup.GET("/:id", middlewares.ControllerAuthCheckMiddleware, middlewares.AdminRoleCheckMiddleware, handlers.GetUserByUUID)
	}

	userGroup := r.Group("/api/v1/")
	{
		userGroup.GET("/gateway/:id/health", middlewares.ControllerAuthCheckMiddleware, handlers.VpnGatewayHealthCheck)
		userGroup.GET("/gateway/list", middlewares.ControllerAuthCheckMiddleware, handlers.GetVpnGatewaysAccessibleByUser)
		userGroup.GET("/gateway/:id/client", middlewares.ControllerAuthCheckMiddleware, handlers.GetVpnClientConfig)
		userGroup.DELETE("/client/:id", middlewares.ControllerAuthCheckMiddleware, handlers.DeleteVpnClient)
	}

}
