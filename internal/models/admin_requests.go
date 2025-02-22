package models

type RegisterUserRequest struct {
	EmailId       string       `json:"email" binding:"required"`
	Password      string       `json:"password"`
	Role          UserRoleEnum `json:"role" binding:"required"`
	IsPasswordSet *bool        `json:"isPasswordSet"`
}

type UpdateUserRequest struct {
	EmailId       string       `json:"email" `
	NewPassword   string       `json:"newPassword"`
	Role          UserRoleEnum `json:"role"`
	IsPasswordSet *bool        `json:"isPasswordSet"`
}

type VpnGatewayPeerConfig struct {
	UserUUID       string `json:"userUuid"`
	UserPublicKey  string `json:"userPublicKey"`
	UserAllowedIPs string `json:"userAllowedIPs"`
}

type VpnGatewayConfig struct {
	ServerAddress    string                 `json:"serverAddress"`
	ServerPrivateKey string                 `json:"serverPrivateKey"`
	ServerListenPort int                    `json:"serverListenPort"`
	Peers            []VpnGatewayPeerConfig `json:"peers"`
}

type VpnClientConfig struct {
	ClientAddress        string
	ClientPrivateKey     string
	ClientDNS            string
	VpnGatewayPublicKey  string
	VpnGatewayEndpoint   string
	VpnGatewayAllowedIPs string
}

type VpnGatewayUpdateUserRequest struct {
	UserUuids []string `json:"userUuids" binding:"required"`
}

type VpnGatewayCreateRequest struct {
	Name      string `json:"name"  binding:"required"`
	Domain    string `json:"domain"  binding:"required"`
	IpAddress string `json:"ipAddress"  binding:"required"`
	VpnCIDR   string `json:"vpnCIDR"  `
	Port      int    `json:"port"  binding:"required"`
	DnsServer string `json:"dnsServer"  binding:"required"`
}

type VpnGatewayUpdateRequest struct {
	Name      string `json:"name" `
	Domain    string `json:"domain"`
	IpAddress string `json:"ipAddress"`
	Port      int    `json:"port"`
	DnsServer string `json:"dnsServer"`
}

type GroupCreateRequest struct {
	Name string `json:"name"  binding:"required"`
}

type GroupUpdateRequest struct {
	Name string `json:"name" binding:"required" `
}

type GroupUpdateUserRequest struct {
	UserUuids []string `json:"userUuids" binding:"required"`
}

type GatewayUpdateGroupRequest struct {
	GroupUuids []string `json:"groupUuids" binding:"required"`
}
