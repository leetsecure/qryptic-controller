package services

import (
	"github.com/leetsecure/qryptic-controller/internal/models"
	"github.com/leetsecure/qryptic-controller/internal/utils/logger"
)

func GetVpnGatewayWGConfig(vpnGatewayUuid string) (models.WGServerConfig, error) {
	log := logger.Default()
	var wgServerConfig models.WGServerConfig

	vpnGateway, err := GetVpnGatewayByUUID(vpnGatewayUuid, false, true, false, false, false)
	if err != nil {
		log.Errorf("Error in fetching vpn gateway details for gateway uuid : %s", vpnGatewayUuid)
		return wgServerConfig, err
	}

	wgServerConfig.WGServerInterfaceConfig.VpnGatewayUuid = vpnGateway.UUID
	wgServerConfig.WGServerInterfaceConfig.DnsServer = vpnGateway.DnsServer
	wgServerConfig.WGServerInterfaceConfig.IPAddress = vpnGateway.VpnCIDR
	wgServerConfig.WGServerInterfaceConfig.ListenPort = vpnGateway.Port
	wgServerConfig.WGServerInterfaceConfig.PrivateKey = vpnGateway.ServerPrivateKey
	wgServerConfig.WGServerInterfaceConfig.PublicKey = vpnGateway.ServerPublicKey

	for _, peer := range vpnGateway.Clients {
		var wgServerPeerConfig models.WGServerPeerConfig
		wgServerPeerConfig.ClientAllowedIPs = peer.AllocatedIP
		wgServerPeerConfig.ClientPublicKey = peer.ClientPublicKey
		wgServerPeerConfig.PresharedKey = ""
		wgServerConfig.WGServerPeerConfigs = append(wgServerConfig.WGServerPeerConfigs, wgServerPeerConfig)
	}

	return wgServerConfig, nil
}
