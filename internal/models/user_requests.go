package models

import "time"

type VpnGatewayUserResponse struct {
	UUID            string `json:"uuid"`
	Name            string `json:"name" `
	Domain          string `json:"domain"`
	IpAddress       string `json:"ipAddress"`
	Port            int    `json:"port"`
	ServerPublicKey string `json:"serverPublicKey"`
}

type UserLoginRequest struct {
	EmailId  string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type WGClientInterfaceConfig struct {
	ClientPrivateKey string `json:"privateKey"`
	AllowedIpAddress string `json:"ipAddress"`
	DnsServer        string `json:"dnsServer"`
}

type WGClientPeerConfig struct {
	AllowedIPs       []string `json:"allowedIPs"`
	ServerPublicKey  string   `json:"publicKey"`
	PresharedKey     string   `json:"presharedKey"`
	PersistantAlive  int      `json:"persistantAlive"`
	VpnGatewayDomain string   `json:"vpnGatewayDomain"`
	VpnGatewayIP     string   `json:"vpnGatewayIP"`
	VpnGatewayPort   int      `json:"vpnGatewayPort"`
}

type WGClientConfig struct {
	ClientUuid              string                  `json:"clientUuid"`
	WGClientInterfaceConfig WGClientInterfaceConfig `json:"clientInterfaceConfig"`
	WGClientPeerConfig      WGClientPeerConfig      `json:"clientPeerConfig"`
	ExpiryTime              time.Time               `json:"expiryTime"`
}
