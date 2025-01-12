package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/leetsecure/qryptic-controller/internal/database"
	"github.com/leetsecure/qryptic-controller/internal/models"
	"github.com/leetsecure/qryptic-controller/internal/utils/auth"
	"github.com/leetsecure/qryptic-controller/internal/utils/logger"
)

func ifUserEmailAlreadyPresent(email string) bool {
	var user models.User
	result := database.DB.Where("email = ?", email).First(&user)
	return result.RowsAffected > 0
}

func RegisterUser(emailID, password, role string, isPasswordSet bool) error {
	var user models.User
	var err error
	user.UUID = uuid.NewString()
	user.Email = emailID
	user.IsPasswordSet = isPasswordSet
	if isPasswordSet && password == "" {
		return errors.New("empty password not allowed")
	}
	exists := ifUserEmailAlreadyPresent(user.Email)
	if exists {
		return errors.New("email id already present")
	}
	if isPasswordSet {
		passwordHash, err := auth.CreatePasswordHash(password)
		if err != nil {
			return err
		}
		user.PasswordHash = passwordHash
	}

	user.Role = models.UserRoleEnum(role)
	err = database.DB.Save(&user).Error
	if err != nil {
		return err
	}
	return nil
}

func BulkRegisterUser(users []models.RegisterUserRequest) error {
	var errs error
	for _, user := range users {
		err := RegisterUser(user.EmailId, user.Password, string(user.Role), *user.IsPasswordSet)
		errs = errors.Join(err)
	}
	return errs
}

func DeleteUser(userUuid string) error {
	log := logger.Default()
	var user models.User
	err := database.DB.Where("uuid = ?", userUuid).Delete(&user).Error
	if err != nil {
		log.Errorf("Error in deleting user with uuid : %s", userUuid)
		return err
	}
	return nil
}

func UpdateUser(userUuid, emailID, newPassword, role string) error {
	log := logger.Default()
	var user models.User
	user, exists, err := getUserFromUuid(userUuid)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("user with given uuid not present")
	}

	if newPassword != "" {
		passwordHash, err := auth.CreatePasswordHash(newPassword)
		if err != nil {
			return err
		}
		user.PasswordHash = passwordHash
	}

	if role != "" {
		user.Role = models.UserRoleEnum(role)
	}

	err = database.DB.Save(&user).Error
	if err != nil {
		log.Errorf("Error in updating  user : %s", userUuid)
		return err
	}
	return nil
}

func ListUsers(includeGateways, includeClients, includeGatewaysWithClients, includeGroups bool) ([]models.User, error) {
	var users []models.User

	dbClient := database.DB

	if includeGatewaysWithClients {
		dbClient = dbClient.Preload("VpnGateways.Clients", "is_active = ?", true)
	}
	if includeGateways {
		dbClient = dbClient.Preload("VpnGateways")
	}
	if includeClients {
		dbClient = dbClient.Preload("Clients", "is_active = ?", true)
	}
	if includeGroups {
		dbClient = dbClient.Preload("Groups")
	}

	if err := dbClient.Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func GetUserByUUID(userUuid string, includeGateways, includeClients, includeGatewaysWithClients, includeGroups bool) (models.User, error) {
	var user models.User
	dbClient := database.DB
	if includeGatewaysWithClients {
		dbClient = dbClient.Preload("VpnGateways.Clients", "is_active = ?", true)
	}
	if includeGateways {
		dbClient = dbClient.Preload("VpnGateways")
	}
	if includeClients {
		dbClient = dbClient.Preload("Clients", "is_active = ?", true)
	}
	if includeGroups {
		dbClient = dbClient.Preload("Groups")
	}

	err := dbClient.Where("uuid = ?", userUuid).First(&user).Error

	if err != nil {
		return user, err
	}
	return user, nil
}
