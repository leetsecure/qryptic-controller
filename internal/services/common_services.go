package services

import (
	"github.com/leetsecure/qryptic-controller/internal/database"
	"github.com/leetsecure/qryptic-controller/internal/externalcomms"
	"github.com/leetsecure/qryptic-controller/internal/models"
	"github.com/leetsecure/qryptic-controller/internal/utils/logger"
)

func getUserFromUuid(userUuid string) (models.User, bool, error) {
	log := logger.Default()
	var user models.User
	result := database.DB.Where("uuid = ?", userUuid).First(&user)
	exists := result.RowsAffected > 0
	if result.Error != nil {
		log.Errorf("Error in fetching user with uuid : %s", userUuid)
		return user, false, result.Error
	}
	if !exists {
		log.Errorf("user : %s not present", userUuid)
		return user, false, nil
	}
	return user, true, nil
}

func getGroupFromUuid(groupUuid string) (models.Group, bool, error) {
	log := logger.Default()
	var group models.Group
	result := database.DB.Where("uuid = ?", groupUuid).First(&group)
	exists := result.RowsAffected > 0
	if result.Error != nil {
		log.Errorf("Error in fetching group with uuid : %s", groupUuid)
		return group, false, result.Error
	}

	if !exists {
		log.Errorf("group : %s not present", groupUuid)
		return group, false, nil
	}
	return group, true, nil
}

func getVpnGatewayFromUuid(vpnGatewayUuid string) (models.VpnGateway, bool, error) {
	log := logger.Default()
	var vpnGateway models.VpnGateway
	result := database.DB.Where("uuid = ?", vpnGatewayUuid).First(&vpnGateway)
	exists := result.RowsAffected > 0
	if result.Error != nil {
		log.Errorf("Error in fetching user with uuid : %s", vpnGatewayUuid)
		return vpnGateway, false, result.Error
	}

	if !exists {
		log.Errorf("vpn gateway : %s not present", vpnGatewayUuid)
		return vpnGateway, false, nil
	}
	return vpnGateway, true, nil
}

func VpnGatewayHealthCheck(vpnGatewayUuid string) (bool, error) {
	log := logger.Default()
	var vpnGateway models.VpnGateway
	err := database.DB.Where("uuid = ?", vpnGatewayUuid).First(&vpnGateway).Error
	if err != nil {
		log.Errorf("error in fetching vpn gateway detail : %s", vpnGatewayUuid)
		return false, err
	}

	healthCheckResponse, err := externalcomms.VpnGatewayHealthCheck(vpnGateway.Domain)
	if err != nil {
		return false, err
	}
	if healthCheckResponse != "" {
		return true, nil
	}
	return false, nil
}
