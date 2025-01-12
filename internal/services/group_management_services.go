package services

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/leetsecure/qryptic-controller/internal/database"
	"github.com/leetsecure/qryptic-controller/internal/models"
	"github.com/leetsecure/qryptic-controller/internal/utils/logger"
)

func CreateGroup(name string) error {
	group := models.Group{
		UUID: uuid.NewString(),
		Name: name,
	}
	err := database.DB.Save(&group).Error
	if err != nil {
		return err
	}
	return nil
}

func DeleteGroup(groupUuid string) error {
	log := logger.Default()
	var group models.Group
	err := database.DB.Where("uuid = ?", groupUuid).Delete(&group).Error
	if err != nil {
		log.Errorf("Error deleting group : %s", groupUuid)
		return err
	}
	return nil
}

func UpdateGroup(groupUuid, name string) error {
	var group models.Group
	err := database.DB.Where("uuid = ?", groupUuid).First(&group).Error
	if err != nil {
		return err
	}
	group.Name = name
	err = database.DB.Save(&group).Error
	if err != nil {
		return err
	}
	return nil
}

func GetGroupByUUID(groupUuid string, includeUsers, includeGateways bool) (models.Group, error) {
	var group models.Group
	dbClient := database.DB
	if includeUsers {
		dbClient = dbClient.Preload("Users")
	}

	if includeGateways {
		dbClient = dbClient.Preload("VpnGateways")
	}

	err := dbClient.Where("uuid = ?", groupUuid).First(&group).Error

	if err != nil {
		return group, err
	}
	return group, nil
}

func ListGroups(includeUsers, includeGateways bool) ([]models.Group, error) {
	var groups []models.Group
	dbClient := database.DB
	if includeUsers {
		dbClient = dbClient.Preload("Users")
	}

	if includeGateways {
		dbClient = dbClient.Preload("VpnGateways")
	}

	if err := dbClient.Find(&groups).Error; err != nil {
		return nil, err
	}

	return groups, nil
}

func AddRemoveUsersInGroup(action string, groupUuid string, userUuids []string) error {
	log := logger.Default()
	// Start a transaction
	tx := database.DB.Begin()
	var group models.Group
	if err := tx.Preload("Users").Where("uuid = ?", groupUuid).First(&group).Error; err != nil {
		log.Errorf("error in fetching group %s with users from db: %v", groupUuid, err)
		tx.Rollback()
		return err
	}

	userUuidMap := make(map[string]bool)
	for _, userUuid := range userUuids {
		userUuidMap[userUuid] = false
	}

	if action == "remove" {
		for _, groupUser := range group.Users {
			if _, ok := userUuidMap[groupUser.UUID]; ok {
				if err := tx.Model(&group).Association("Users").Delete(&groupUser); err != nil {
					log.Errorf("error in removing user %s from group %s: %v", groupUser.UUID, groupUuid, err)
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
			if err := tx.Model(&group).Association("Users").Append(&newVpnUser); err != nil {
				log.Errorf("error in adding user %s to VPN Gateway %s: %v", userUuid, groupUuid, err)
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
		log.Errorf("error committing transaction for group %s: %v", groupUuid, err)
		return err
	}

	return nil
}
