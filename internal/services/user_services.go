package services

import (
	"errors"
	"fmt"
	"time"

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

// User

func IfUserHasAccessToVpnGateway(userUuid string, vpnGatewayUuid string) (bool, error) {
	log := logger.Default()
	var user models.User

	err := database.DB.Preload("VpnGateways", "uuid = ?", vpnGatewayUuid).
		Where("uuid = ?", userUuid).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Infof("User : %s does NOT have access to VPN Gateway : %s", userUuid, vpnGatewayUuid)
			return false, nil
		} else {
			log.Errorf("Error in db query for check User : %s accessbility of VPN Gateway : %s", userUuid, vpnGatewayUuid)
			return false, err
		}
	} else if len(user.VpnGateways) > 0 {
		return true, nil
	} else {
		log.Infof("User : %s does NOT have access to VPN Gateway : %s", userUuid, vpnGatewayUuid)
		return false, nil
	}
}

func IfUsersHasAccessToGatewayV2(userUuid string, vpnGatewayUuid string) (bool, error) {
	// log := logger.Default()
	var count int64

	// Check if the user is directly associated with the VPN gateway
	err := database.DB.Table("vpn_gateways").
		Joins("JOIN user_vpngateways ON vpn_gateways.id = user_vpngateways.vpn_gateway_id").
		Joins("JOIN users ON users.id = user_vpngateways.user_id").
		Where("users.uuid = ? AND vpn_gateways.uuid = ?", userUuid, vpnGatewayUuid).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	// Check if any of the user's groups is associated with the VPN gateway
	err = database.DB.Table("vpn_gateways").
		Joins("JOIN group_vpngateways ON vpn_gateways.id = group_vpngateways.vpn_gateway_id").
		Joins("JOIN groups ON groups.id = group_vpngateways.group_id").
		Joins("JOIN group_users ON groups.id = group_users.group_id").
		Joins("JOIN users ON users.id = group_users.user_id").
		Where("users.uuid = ? AND vpn_gateways.uuid = ?", userUuid, vpnGatewayUuid).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func ListAccessibleVPNs(userUuid string) ([]models.VpnGatewayUserResponse, error) {
	var vpnGateways []models.VpnGatewayUserResponse
	var user models.User
	err := database.DB.Preload("VpnGateways").Preload("Groups.VpnGateways").
		Where("uuid = ?", userUuid).
		First(&user).Error

	if err != nil {
		return vpnGateways, err
	}
	uniqueGatewaysUuid := map[string]bool{}
	for _, vpnGateway := range user.VpnGateways {
		if !uniqueGatewaysUuid[vpnGateway.UUID] {
			vpnGateways = append(vpnGateways, models.VpnGatewayUserResponse{
				UUID:            vpnGateway.UUID,
				Name:            vpnGateway.Name,
				Domain:          vpnGateway.Domain,
				IpAddress:       vpnGateway.IpAddress,
				Port:            vpnGateway.Port,
				ServerPublicKey: vpnGateway.ServerPublicKey,
			})
		}
		uniqueGatewaysUuid[vpnGateway.UUID] = true
	}
	for _, group := range user.Groups {
		for _, vpnGateway := range group.VpnGateways {
			if _, ok := uniqueGatewaysUuid[vpnGateway.UUID]; !ok {
				vpnGateways = append(vpnGateways, models.VpnGatewayUserResponse{
					UUID:            vpnGateway.UUID,
					Name:            vpnGateway.Name,
					Domain:          vpnGateway.Domain,
					IpAddress:       vpnGateway.IpAddress,
					Port:            vpnGateway.Port,
					ServerPublicKey: vpnGateway.ServerPublicKey,
				})
			}
			uniqueGatewaysUuid[vpnGateway.UUID] = true
		}
	}
	return vpnGateways, nil
}

func ListAccessibleVPNsV2(userUuid string) ([]models.VpnGatewayUserResponse, error) {
	var vpnGateways []models.VpnGateway

	// Get the user's ID from their UUID
	var user models.User
	if err := database.DB.Where("uuid = ?", userUuid).First(&user).Error; err != nil {
		return nil, err // Return early if the user is not found
	}

	err := database.DB.
		Distinct("vpn_gateways.*"). // Ensure distinct gateways
		Model(&models.VpnGateway{}).
		Joins("LEFT JOIN user_vpngateways ON user_vpngateways.vpn_gateway_id = vpn_gateways.id").
		Joins("LEFT JOIN group_vpngateways ON group_vpngateways.vpn_gateway_id = vpn_gateways.id").
		Joins("LEFT JOIN group_users ON group_users.group_id = group_vpngateways.group_id").
		Joins("LEFT JOIN users AS group_users_user ON group_users_user.id = group_users.user_id").
		Where("user_vpngateways.user_id = ? OR group_users_user.id = ?", user.ID, user.ID).
		Find(&vpnGateways).Error

	if err != nil {
		return nil, err
	}

	var vpnGatewaysResponse []models.VpnGatewayUserResponse
	for _, gateway := range vpnGateways {
		vpnGatewaysResponse = append(vpnGatewaysResponse, models.VpnGatewayUserResponse{
			UUID:            gateway.UUID,
			Name:            gateway.Name,
			Domain:          gateway.Domain,
			IpAddress:       gateway.IpAddress,
			Port:            gateway.Port,
			ServerPublicKey: gateway.ServerPublicKey,
		})
	}
	return vpnGatewaysResponse, nil
}

func ListAccessibleGatewaysByGroup(groupUuid string) ([]models.VpnGatewayUserResponse, error) {
	group, err := GetGroupByUUID(groupUuid, false, true)
	if err != nil {
		return nil, err
	}

	var vpnGatewaysResponse []models.VpnGatewayUserResponse
	for _, gateway := range group.VpnGateways {
		vpnGatewaysResponse = append(vpnGatewaysResponse, models.VpnGatewayUserResponse{
			UUID:            gateway.UUID,
			Name:            gateway.Name,
			Domain:          gateway.Domain,
			IpAddress:       gateway.IpAddress,
			Port:            gateway.Port,
			ServerPublicKey: gateway.ServerPublicKey,
		})
	}
	return vpnGatewaysResponse, nil
}

func GetUserWithGatewaysAndClients(userUuid string) (models.User, error) {
	var user models.User
	err := database.DB.Preload("VpnGateways.Clients").Where("uuid = ?", userUuid).First(&user).Error
	if err != nil {
		return user, err
	}
	return user, nil
}

func CreateVpnGatewayUserClient(userUuid string, vpnGatewayUuid string) (models.WGClientConfig, bool, error) {

	var wgClientConfig models.WGClientConfig
	// check if user has access for given vpn gateway
	accessible, err := IfUsersHasAccessToGatewayV2(userUuid, vpnGatewayUuid)
	if err != nil {
		return wgClientConfig, false, err
	}
	if !accessible {
		return wgClientConfig, false, nil
	}
	vpnGateway, exists, err := getVpnGatewayFromUuid(vpnGatewayUuid)
	if err != nil || !exists {
		return wgClientConfig, false, err
	}

	user, exists, err := getUserFromUuid(userUuid)
	if err != nil || !exists {
		return wgClientConfig, false, err
	}

	// look for IP from IP Pool of VPN Gateway
	ipPool, err := getFirstAvailableIP(vpnGateway.ID)
	if err != nil {
		return wgClientConfig, false, err
	}

	//Create new client with expiry time
	expiryTime := time.Now().Add(config.ClientExpiry)

	publicKey, privateKey, err := wireguard.GenerateWireguardPublicPrivateKeys()
	if err != nil {
		return wgClientConfig, false, err
	}

	allocatedIP := fmt.Sprintf("%s/32", ipPool.IP)

	client := &models.Client{
		UUID:             uuid.NewString(),
		UserID:           user.ID,
		VpnGatewayID:     vpnGateway.ID,
		ExpiryTime:       expiryTime,
		IsActive:         true,
		AllocatedIP:      allocatedIP,
		AllowedIPs:       []string{"0.0.0.0/0"}[0],
		ClientPublicKey:  publicKey,
		ClientPrivateKey: privateKey,
		DnsServer:        vpnGateway.DnsServer,
		PresharedKey:     "",
	}

	if err := database.DB.Create(client).Error; err != nil {
		return wgClientConfig, false, err
	}

	if err := database.DB.Model(&models.IPPool{}).Where("id = ?", ipPool.ID).Update("assigned", true).Error; err != nil {
		return wgClientConfig, false, err
	}

	var wgServerPeerConfigs []models.WGServerPeerConfig

	wgServerPeerConfigs = append(wgServerPeerConfigs, models.WGServerPeerConfig{
		ClientAllowedIPs: ipPool.IP,
		ClientPublicKey:  client.ClientPublicKey,
		PresharedKey:     "",
	})

	//send new client to vpn gateway
	err = addNewClientsRequestToVpnGateway(vpnGateway, wgServerPeerConfigs)
	if err != nil {
		return wgClientConfig, false, err
	}

	//allocate the IP to IP allocation table

	// send the client details to user

	wgClientConfig.WGClientInterfaceConfig.ClientPrivateKey = client.ClientPrivateKey
	wgClientConfig.WGClientInterfaceConfig.AllowedIpAddress = client.AllocatedIP
	wgClientConfig.WGClientInterfaceConfig.DnsServer = client.DnsServer
	wgClientConfig.WGClientPeerConfig.AllowedIPs = []string{client.AllowedIPs}
	wgClientConfig.WGClientPeerConfig.PersistantAlive = 25
	wgClientConfig.WGClientPeerConfig.PresharedKey = ""
	wgClientConfig.WGClientPeerConfig.ServerPublicKey = vpnGateway.ServerPublicKey
	wgClientConfig.WGClientPeerConfig.VpnGatewayIP = vpnGateway.IpAddress
	wgClientConfig.WGClientPeerConfig.VpnGatewayDomain = vpnGateway.Domain
	wgClientConfig.WGClientPeerConfig.VpnGatewayPort = vpnGateway.Port
	wgClientConfig.ExpiryTime = client.ExpiryTime
	wgClientConfig.ClientUuid = client.UUID
	return wgClientConfig, true, nil
}

func getFirstAvailableIP(vpnGatewayID uint) (*models.IPPool, error) {
	var ipPool models.IPPool
	result := database.DB.Where("vpn_gateway_id = ? AND assigned = ?", vpnGatewayID, false).First(&ipPool)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("new ip not available, try clearing expired clients")
	}
	return &ipPool, nil
}

func ifUserHasAccessToClient(clientUuid string, userUuid string) (bool, error) {
	log := logger.Default()
	var user models.User

	err := database.DB.Preload("Clients", "uuid = ? and is_active = ?", clientUuid, true).
		Where("uuid = ?", userUuid).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Infof("User : %s does NOT have access to Client : %s", userUuid, clientUuid)
			return false, nil
		} else {
			log.Errorf("Error in db query for check User : %s accessbility of Client : %s", userUuid, clientUuid)
			return false, err
		}
	} else if len(user.Clients) > 0 {
		return true, nil
	} else {
		log.Infof("User : %s does NOT have access to VPN Gateway : %s", userUuid, clientUuid)
		return false, nil
	}
}

func DeleteClientFromUserAndVpnGatewayByUser(clientUuid string, userUuid string) error {
	accessible, err := ifUserHasAccessToClient(clientUuid, userUuid)
	if err != nil {
		return err
	}
	if !accessible {
		return errors.New("user doesn't have access to given client uuid")
	}

	err = DeleteClientFromUserAndVpnGateway(clientUuid)
	return err
}

// To be called by scheduler on expiryTime pass
func DeleteExpiredClientsFromUserAndVpnGateway() error {
	var expiredClients []models.Client
	currentTime := time.Now()

	err := database.DB.Preload("VpnGateway").Where("is_active = ? AND expiry_time < ?", true, currentTime).Find(&expiredClients).Error
	if err != nil {
		return err
	}

	for _, client := range expiredClients {
		var wgServerPeerConfigs []models.WGServerPeerConfig
		wgServerPeerConfigs = append(wgServerPeerConfigs, models.WGServerPeerConfig{
			ClientPublicKey: client.ClientPublicKey,
		})
		if client.VpnGateway == nil {
			continue
		}
		deleteClientsRequestFromVpnGateway(*client.VpnGateway, wgServerPeerConfigs) //delete client from vpn gateway
		//make IP available in IP pool
		err = database.DB.Model(&models.IPPool{}).Where("ip = ? AND vpn_gateway_id = ?", client.AllocatedIP, client.VpnGatewayID).Update("assigned", false).Error
		if err != nil {
			return err
		}
	}

	err = database.DB.Model(&models.Client{}).Where("expiry_time < ?", currentTime).Update("is_active", false).Error
	if err != nil {
		return err
	}

	// deallocate the IP to IP allocation table
	return nil
}

func DeleteClientFromUserAndVpnGateway(clientUuid string) error {

	//deactivate the client
	var client models.Client
	result := database.DB.Preload("VpnGateway").Where("uuid = ?", clientUuid).First(&client)
	if result.Error != nil {
		return result.Error
	}

	client.IsActive = false

	var wgServerPeerConfigs []models.WGServerPeerConfig

	wgServerPeerConfigs = append(wgServerPeerConfigs, models.WGServerPeerConfig{
		ClientPublicKey: client.ClientPublicKey,
	})

	deleteClientsRequestFromVpnGateway(*client.VpnGateway, wgServerPeerConfigs) //delete client from vpn gateway

	database.DB.Save(&client)

	//make IP available in IP pool
	result = database.DB.Model(&models.IPPool{}).Where("ip = ? AND vpn_gateway_id = ?", client.AllocatedIP, client.VpnGatewayID).Update("assigned", false)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func addNewClientsRequestToVpnGateway(vpnGateway models.VpnGateway, wgServerPeerConfigs []models.WGServerPeerConfig) error {
	log := logger.Default()
	authToken, err := auth.CreateVpnGatewayToken(vpnGateway.UUID, vpnGateway.JwtSecretKey)
	if err != nil {
		return err
	}
	responseBody, responseStatusCode, err := externalcomms.AddNewPeerInVpnGateway(vpnGateway.Domain, authToken, wgServerPeerConfigs)
	if err != nil {
		return nil
	}
	log.Infof("response body of request to add peers : %s and status code is %d", responseBody, responseStatusCode)
	if responseStatusCode != 200 {
		return errors.New("request not fulfilled")
	}
	return nil
}

func deleteClientsRequestFromVpnGateway(vpnGateway models.VpnGateway, wgServerPeerConfigs []models.WGServerPeerConfig) error {
	log := logger.Default()
	authToken, err := auth.CreateVpnGatewayToken(vpnGateway.UUID, vpnGateway.JwtSecretKey)
	if err != nil {
		return err
	}
	responseBody, responseStatusCode, err := externalcomms.DeletePeerInVpnGateway(vpnGateway.Domain, authToken, wgServerPeerConfigs)
	if err != nil {
		return nil
	}
	log.Infof("response body of request to delete peers : %s and status code is %d", responseBody, responseStatusCode)
	if responseStatusCode != 200 {
		return errors.New("request not fulfilled")
	}
	return nil
}
