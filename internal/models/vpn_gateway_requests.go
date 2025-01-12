package models

type WGServerInterfaceConfig struct {
	VpnGatewayUuid string `json:"vpnGatewayUuid"`
	PublicKey      string `json:"publicKey"`
	PrivateKey     string `json:"privateKey"`
	IPAddress      string `json:"ipAddress"`
	ListenPort     int    `json:"listenPort"`
	PostUp         string `json:"postUp"`
	PostDown       string `json:"postDown"`
	DnsServer      string `json:"dnsServer"`
}

type WGServerPeerConfig struct {
	ClientAllowedIPs string `json:"clientAllowedIPs"`
	ClientPublicKey  string `json:"clientPublicKey"`
	PresharedKey     string `json:"presharedKey"`
}

type WGServerConfig struct {
	WGServerInterfaceConfig WGServerInterfaceConfig `json:"wgServerInterfaceConfig"`
	WGServerPeerConfigs     []WGServerPeerConfig    `json:"wgServerPeerConfigs"`
}
