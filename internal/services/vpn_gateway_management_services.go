package services

import (
	"errors"
	"fmt"
	"net"

	"github.com/google/uuid"
	"github.com/leetsecure/qryptic-controller/internal/config"
	"github.com/leetsecure/qryptic-controller/internal/database"
	"github.com/leetsecure/qryptic-controller/internal/externalcomms"
	"github.com/leetsecure/qryptic-controller/internal/models"
	"github.com/leetsecure/qryptic-controller/internal/utils/auth"
	"github.com/leetsecure/qryptic-controller/internal/utils/logger"
	"github.com/leetsecure/qryptic-controller/internal/utils/wireguard"
	"gorm.io/gorm"
)

// Admin User
func AddRemoveUsersInVpnGateway(action string, vpnGatewayUuid string, userUuids []string) error {
	log := logger.Default()

	// Start a transaction
	tx := database.DB.Begin()

	var vpnGateway models.VpnGateway
	if err := tx.Preload("Users").Where("uuid = ?", vpnGatewayUuid).First(&vpnGateway).Error; err != nil {
		log.Errorf("error in fetching vpn gateway %s with users from db: %v", vpnGatewayUuid, err)
		tx.Rollback()
		return err
	}

	userUuidMap := make(map[string]bool)
	for _, userUuid := range userUuids {
		userUuidMap[userUuid] = false
	}

	if action == "remove" {
		for _, vpnUser := range vpnGateway.Users {
			if _, ok := userUuidMap[vpnUser.UUID]; ok {
				if err := tx.Model(&vpnGateway).Association("Users").Delete(&vpnUser); err != nil {
					log.Errorf("error in removing user %s from VPN Gateway %s: %v", vpnUser.UUID, vpnGatewayUuid, err)
					tx.Rollback()
					return err
				}
			}
		}
	} else if action == "add" {
		for userUuid := range userUuidMap {
			newVpnUser, exists, err := getUserFromUuid(userUuid)
			if err != nil {
				tx.Rollback()
				return err
			}
			if !exists {
				continue
			}
			if err := tx.Model(&vpnGateway).Association("Users").Append(&newVpnUser); err != nil {
				log.Errorf("error in adding user %s to VPN Gateway %s: %v", userUuid, vpnGatewayUuid, err)
				tx.Rollback()
				return err
			}
		}
	} else {
		tx.Rollback()
		return fmt.Errorf("invalid action: %s", action)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Errorf("error committing transaction for VPN Gateway %s: %v", vpnGatewayUuid, err)
		return err
	}

	return nil
}
func AddRemoveGroupsInVpnGateway(action string, vpnGatewayUuid string, groupUuids []string) error {
	log := logger.Default()

	// Start a transaction
	tx := database.DB.Begin()

	var vpnGateway models.VpnGateway
	if err := tx.Preload("Groups").Where("uuid = ?", vpnGatewayUuid).First(&vpnGateway).Error; err != nil {
		log.Errorf("error in fetching vpn gateway %s with groups from db: %v", vpnGatewayUuid, err)
		tx.Rollback()
		return err
	}

	groupUuidMap := make(map[string]bool)
	for _, groupUuid := range groupUuids {
		groupUuidMap[groupUuid] = false
	}

	if action == "remove" {
		for _, vpnGroup := range vpnGateway.Groups {
			if _, ok := groupUuidMap[vpnGroup.UUID]; ok {
				if err := tx.Model(&vpnGateway).Association("Groups").Delete(&vpnGroup); err != nil {
					log.Errorf("error in removing group %s from VPN Gateway %s: %v", vpnGroup.UUID, vpnGatewayUuid, err)
					tx.Rollback()
					return err
				}
			}
		}
	} else if action == "add" {
		for groupUuid := range groupUuidMap {
			newVpnGroup, exists, err := getGroupFromUuid(groupUuid)
			if err != nil {
				tx.Rollback()
				return err
			}
			if !exists {
				continue
			}
			if err := tx.Model(&vpnGateway).Association("Groups").Append(&newVpnGroup); err != nil {
				log.Errorf("error in adding group %s to VPN Gateway %s: %v", groupUuid, vpnGatewayUuid, err)
				tx.Rollback()
				return err
			}
		}
	} else {
		tx.Rollback()
		return fmt.Errorf("invalid action: %s", action)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Errorf("error committing transaction for VPN Gateway %s: %v", vpnGatewayUuid, err)
		return err
	}

	return nil
}

func CreateVpnGateway(name, domain, ipAddress, vpnCidr string, port int, dnsServer string) error {
	log := logger.Default()
	log.Info("start creating vpn gateway")
	publicKey, privateKey, err := wireguard.GenerateWireguardPublicPrivateKeys()
	if err != nil {
		return err
	}

	secretKey := auth.RandomStringGenerator(32)
	vpnGateway := models.VpnGateway{
		UUID:             uuid.NewString(),
		Name:             name,
		JwtSecretKey:     secretKey,
		JwtAlgorithm:     "HS256",
		ServerPublicKey:  publicKey,
		ServerPrivateKey: privateKey,
		Domain:           domain,
		VpnCIDR:          vpnCidr,
		IpAddress:        ipAddress,
		Port:             port,
		DnsServer:        dnsServer,
	}
	tx := database.DB.Begin()

	err = tx.Save(&vpnGateway).Error
	if err != nil {
		log.Info("issue saving vpn gateway")
		tx.Rollback()
		return err
	}

	err = createIPPoolForGateway(tx, &vpnGateway)
	if err != nil {
		log.Info("issue creating vpn gateway ippool")
		tx.Rollback()
		return err
	}
	if err = tx.Commit().Error; err != nil {
		log.Errorf("error committing transaction for gateway creation")
		return err
	}
	return nil
}

func createIPPoolForGateway(tx *gorm.DB, vpnGateway *models.VpnGateway) error {
	log := logger.Default()
	_, ipNet, err := net.ParseCIDR(vpnGateway.VpnCIDR)
	if err != nil {
		log.Errorf("invalid CIDR: %v", err)
		return err
	}
	firstIP := dupIP(ipNet.IP)
	incIP(firstIP)
	if !ipNet.Contains(firstIP) {
		log.Errorf("No IP available for clients for this CIDR : %s", vpnGateway.VpnCIDR)
		return errors.New("no ip available for clients for this small cidr")
	}
	lastIP := lastIP(ipNet)
	var ipPools []models.IPPool

	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); incIP(ip) {
		// Skip the network and broadcast addresses
		if ip.Equal(ipNet.IP) || ip.Equal(lastIP) || ip.Equal(firstIP) {
			continue
		}
		// Create an IPPool entry for each IP
		ipPool := models.IPPool{
			UUID:         uuid.NewString(),
			IP:           ip.String(),
			Assigned:     false,
			VpnGatewayID: vpnGateway.ID,
		}
		ipPools = append(ipPools, ipPool)

	}
	// Insert the IPPool entry into the database
	if err := tx.Create(&ipPools).Error; err != nil {
		log.Errorf("failed to create IPPool entry for gateway %s with CIDR %s", vpnGateway.UUID, vpnGateway.VpnCIDR)
		return err
	}
	return nil
}

// Function to increment an IP address
func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] != 0 {
			break
		}
	}
}

// Function to get the last IP address in the subnet (broadcast address)
func lastIP(n *net.IPNet) net.IP {
	ip := make(net.IP, len(n.IP))
	for i := range ip {
		ip[i] = n.IP[i] | ^n.Mask[i]
	}
	return ip
}

// Function to create duplicate copy of ipNet
func dupIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}

func DeleteVpnGateway(vpnGatewayUuid string) error {
	log := logger.Default()
	var vpnGateway models.VpnGateway

	err := ClearVpnGatewayClientsAndIPPool(vpnGatewayUuid)
	if err != nil {
		log.Errorf("Error in clearing Clients And IPPool : %s", vpnGatewayUuid)
		return err
	}
	err = database.DB.Where("uuid = ?", vpnGatewayUuid).Delete(&vpnGateway).Error
	if err != nil {
		log.Errorf("Error deleting vpn gateway : %s", vpnGatewayUuid)
		return err
	}
	return nil
}

func UpdateVpnGateway(vpnGatewayUuid, name, domain, ipAddress string, port int, dnsServer string) error {
	var vpnGateway models.VpnGateway
	err := database.DB.Where("uuid = ?", vpnGatewayUuid).First(&vpnGateway).Error
	if err != nil {
		return err
	}
	if name != "" {
		vpnGateway.Name = name
	}
	if domain != "" {
		vpnGateway.Domain = domain
	}
	if ipAddress != "" {
		vpnGateway.IpAddress = ipAddress
	}
	if port != 0 {
		vpnGateway.Port = port
	}

	if dnsServer != "" {
		vpnGateway.DnsServer = dnsServer
	}

	err = database.DB.Save(&vpnGateway).Error
	if err != nil {
		return err
	}
	return nil
}

func ClearVpnGatewayClientsAndIPPool(vpnGatewayUuid string) error {

	tx := database.DB.Begin()

	var vpnGateway models.VpnGateway
	if err := tx.Where("uuid = ?", vpnGatewayUuid).First(&vpnGateway).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to find VPN Gateway: %w", err)
	}

	if err := tx.Model(&models.Client{}).Where("vpn_gateway_id = ? AND is_active = true", vpnGateway.ID).
		Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to deactivate clients: %w", err)
	}

	if err := tx.Model(&models.IPPool{}).Where("vpn_gateway_id = ?", vpnGateway.ID).
		Update("assigned", false).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to unassign IPs in IPPool: %w", err)
	}

	if tx.Commit().Error != nil {
		return tx.Commit().Error
	}
	authToken, err := auth.CreateVpnGatewayToken(vpnGateway.UUID, vpnGateway.JwtSecretKey)
	if err != nil {
		return err
	}
	_, _, err = externalcomms.RestartVpnGateway(vpnGateway.Domain, authToken)
	if err != nil {
		return err
	}
	return nil
}

func ListVpnGateways(includeUsers, includeClients, includeUsersWithClients, includeGroups, includeIpPool bool) ([]models.VpnGateway, error) {
	var vpnGateways []models.VpnGateway
	dbClient := database.DB

	if includeUsersWithClients {
		dbClient = dbClient.Preload("Users.Clients", "is_active = ?", true)
	}
	if includeUsers {
		dbClient = dbClient.Preload("Users")
	}
	if includeClients {
		dbClient = dbClient.Preload("Clients", "is_active = ?", true)
	}
	if includeIpPool {
		dbClient = dbClient.Preload("IPPool")
	}
	if includeGroups {
		dbClient = dbClient.Preload("Groups")
	}
	if err := dbClient.Find(&vpnGateways).Error; err != nil {
		return nil, err
	}

	return vpnGateways, nil
}

func GetVpnGatewayByUUID(vpnGatewayUuid string, includeUsers, includeClients, includeUsersWithClients, includeGroups, includeIpPool bool) (models.VpnGateway, error) {
	var vpnGateway models.VpnGateway

	dbClient := database.DB

	if includeUsersWithClients {
		dbClient = dbClient.Preload("Users.Clients", "is_active = ?", true)
	}
	if includeUsers {
		dbClient = dbClient.Preload("Users")
	}
	if includeClients {
		dbClient = dbClient.Preload("Clients", "is_active = ?", true)
	}
	if includeIpPool {
		dbClient = dbClient.Preload("IPPool")
	}
	if includeGroups {
		dbClient = dbClient.Preload("Groups")
	}

	if err := dbClient.Where("uuid = ?", vpnGatewayUuid).First(&vpnGateway).Error; err != nil {
		return models.VpnGateway{}, err
	}

	return vpnGateway, nil
}

func CreateVpnGatewayDeploymentConfig(vpnGatewayUuid string) (string, error) {
	log := logger.Default()
	var vpnGateway models.VpnGateway
	err := database.DB.Where("uuid = ?", vpnGatewayUuid).First(&vpnGateway).Error
	if err != nil {
		log.Errorf("Error in fetching vpn gateway : %s details from database", vpnGatewayUuid)
		return "", err
	}
	deploymentFormat := `docker run -d --cap-add=NET_ADMIN --cap-add=SYS_MODULE --sysctl='net.ipv4.conf.all.src_valid_mark=1' --sysctl='net.ipv4.ip_forward=1' --sysctl='net.ipv6.conf.all.forwarding=1' -p %s:51820/udp  -p %s:%s  -e VpnGatewayUuid='%s' -e VpnGatewayControllerJWTSecretKey='%s' -e VpnGatewayControllerJWTAlgorithm='%s' -e ControllerVGWConfigUrlEndpoint='%s' -e ApplicationPort='%s' %s`
	vpnGatewayJwtSecretKey := vpnGateway.JwtSecretKey
	vpnGatewayJwtAlgorithm := vpnGateway.JwtAlgorithm
	controllerConfigUrl := fmt.Sprintf(config.GatewayCallbackForConfigTemplate, config.ControllerDomain)
	vpnGatewayApplicationPort := config.GatewayPort
	wireguardPort := config.WireguardPort
	imageName := config.VpnGatewayApplicationImageName

	deployment := fmt.Sprintf(deploymentFormat, wireguardPort, vpnGatewayApplicationPort, vpnGatewayApplicationPort, vpnGatewayUuid, vpnGatewayJwtSecretKey, vpnGatewayJwtAlgorithm, controllerConfigUrl, vpnGatewayApplicationPort, imageName)

	return deployment, nil
}
