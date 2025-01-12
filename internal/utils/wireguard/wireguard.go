package wireguard

import (
	"errors"
	"fmt"
	"io"

	"os/exec"

	"github.com/leetsecure/qryptic-controller/internal/models"
)

func GenerateWireguardPublicPrivateKeys() (string, string, error) {

	cmd := exec.Command("wg", "genkey")
	privateKey, err := cmd.Output()
	if err != nil {
		return "", "", errors.New("could not generate private key")
	}

	cmd = exec.Command("wg", "pubkey")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", "", errors.New("could not get stdin of pubkey command")
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, string(privateKey))
	}()

	publicKey, err := cmd.Output()
	if err != nil {
		return "", "", errors.New("could not generate public key")
	}
	publicKeyString := string(publicKey)
	publicKeyString = publicKeyString[:len(publicKeyString)-1] // removing \n from end

	privateKeyString := string(privateKey)
	privateKeyString = privateKeyString[:len(privateKeyString)-1] // removing \n from end

	return publicKeyString, privateKeyString, nil
}

func CreateClientConfig(vpnClientConfig models.VpnClientConfig) string {

	client_config_format := `
	[Interface]
	Address = %s
	PrivateKey = %s
	DNS = %s

	[Peer]
	PublicKey = %s
	Endpoint = %s
	AllowedIPs = %s	
	`
	return fmt.Sprintf(client_config_format, vpnClientConfig.ClientAddress, vpnClientConfig.ClientPrivateKey, vpnClientConfig.ClientDNS, vpnClientConfig.VpnGatewayPublicKey, vpnClientConfig.VpnGatewayEndpoint, vpnClientConfig.VpnGatewayAllowedIPs)
}

func CreateVpnGatewayConfig(vpnGatewayConfig models.VpnGatewayConfig) string {

	vpn_gateway_config_format := `
	[Interface]
	Address = %s
	PrivateKey = %s
	ListenPort = %d

	%s
	`
	formattedPeers := ""

	peer_format := `
	[Peer]
	# User UUID : %s #
	PublicKey = %s
	AllowedIPs = %s
	`

	for _, peer := range vpnGatewayConfig.Peers {
		formattedPeers = formattedPeers + "\n" + fmt.Sprintf(peer_format, peer.UserUUID, peer.UserPublicKey, peer.UserAllowedIPs)
	}

	return fmt.Sprintf(vpn_gateway_config_format, vpnGatewayConfig.ServerAddress, vpnGatewayConfig.ServerPrivateKey, vpnGatewayConfig.ServerListenPort, formattedPeers)
}
