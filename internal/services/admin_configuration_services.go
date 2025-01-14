package services

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/leetsecure/qryptic-controller/internal/config"
	"github.com/leetsecure/qryptic-controller/internal/database"
	"github.com/leetsecure/qryptic-controller/internal/models"
	"github.com/leetsecure/qryptic-controller/internal/utils/auth"
	"github.com/leetsecure/qryptic-controller/internal/utils/logger"
)

func InitAdminConfig() error {
	err := createInitialAdminConfiguration()
	if !config.TempUserCreated {
		createInitialTempUser()
	}
	if config.AllowSSOLogin {
		initializeSSOProviders()
	}
	return err
}

func createInitialAdminConfiguration() error {
	log := logger.Default()
	var adminConfiguration models.AdminConfiguration
	result := database.DB.First(&adminConfiguration)
	if result.RowsAffected > 0 {
		log.Info("initial admin configuration already created")
		config.AllowPasswordLogin = adminConfiguration.AllowPasswordLogin
		config.AllowSSOLogin = adminConfiguration.AllowSSOLogin
		config.UserAuthJwtSecretKey = adminConfiguration.UserAuthJwtSecretKey
		config.TempUserCreated = adminConfiguration.TempUserCreated
		config.TempUserActive = adminConfiguration.TempUserActive
		config.UserAuthSSOJwtSecretKey = adminConfiguration.UserAuthSSOJwtSecretKey

		if adminConfiguration.UserAuthSSOJwtSecretKey == "" {
			UpdateUserAuthSSOJwtSecretKey()
		}
		return nil
	}

	adminConfiguration.UUID = uuid.NewString()
	adminConfiguration.AllowPasswordLogin = config.AllowPasswordLogin
	adminConfiguration.AllowSSOLogin = config.AllowSSOLogin
	adminConfiguration.UserAuthJwtSecretKey = auth.RandomStringGenerator(32)
	adminConfiguration.UserAuthSSOJwtSecretKey = auth.RandomStringGenerator(32)
	adminConfiguration.UserJWTAlgorithm = "HS256"
	adminConfiguration.GatewayJWTAlgorithm = "HS256"
	err := createInitialTempUser()
	if err != nil {
		return errors.Join(err, errors.New("unable to create initial temp user"))
	}
	adminConfiguration.TempUserCreated = true
	adminConfiguration.TempUserActive = true
	config.UserAuthJwtSecretKey = adminConfiguration.UserAuthJwtSecretKey
	return database.DB.Save(&adminConfiguration).Error
}

func UpdateUserAuthSSOJwtSecretKey() error {
	log := logger.Default()
	var adminConfiguration models.AdminConfiguration
	err := database.DB.First(&adminConfiguration).Error
	if err != nil {
		log.Error("error in fetching admin configuration")
		return err
	}
	adminConfiguration.UserAuthSSOJwtSecretKey = auth.RandomStringGenerator(32)
	err = database.DB.Save(&adminConfiguration).Error
	if err != nil {
		log.Error("error in saving admin configuration")
		return err
	}
	return nil
}

func UpdateUserAuthJwtSecretKey() error {
	log := logger.Default()
	var adminConfiguration models.AdminConfiguration
	err := database.DB.First(&adminConfiguration).Error
	if err != nil {
		log.Error("error in fetching admin configuration")
		return err
	}
	adminConfiguration.UserAuthJwtSecretKey = auth.RandomStringGenerator(32)
	err = database.DB.Save(&adminConfiguration).Error
	if err != nil {
		log.Error("error in saving admin configuration")
		return err
	}
	return nil
}

func createInitialTempUser() error {
	log := logger.Default()
	tempEmailId := fmt.Sprintf("%s@qryptic.com", auth.RandomStringGenerator(10))
	tempPassword := fmt.Sprintf("%s@%s#%s", auth.RandomStringGenerator(5), auth.RandomStringGenerator(5), auth.RandomStringGenerator(5))
	log.Infof("Temporary Email Id : %s \n Temporary Password : %s", tempEmailId, tempPassword)
	err := RegisterUser(tempEmailId, tempPassword, string(models.AdminRole), true)
	if err != nil {
		return err
	}
	config.TempUserCreated = true
	config.TempUserActive = true
	return nil
}

func initializeSSOProviders() error {
	log := logger.Default()
	var ssoConfigs []models.SSOConfig
	if err := database.DB.Where("platform = ?", "Website").First(&ssoConfigs).Error; err != nil {
		return err
	}
	log.Infof("Number of sso configs : %d", len(ssoConfigs))
	if len(ssoConfigs) > 0 {
		config.SsoConfig = append(config.SsoConfig, map[string]string{
			"Platform":     ssoConfigs[0].Platform,
			"Provider":     ssoConfigs[0].Provider,
			"ClientID":     ssoConfigs[0].ClientID,
			"ClientSecret": ssoConfigs[0].ClientSecret,
		})
		config.GoogleClientID = ssoConfigs[0].ClientID
		config.GoogleClientSecret = ssoConfigs[0].ClientSecret
		log.Infof("GoogleClientID : %s", config.GoogleClientID)
	}
	return nil
}

func GetAdminConfiguration(includeSsoConfigs bool) (models.AdminConfiguration, error) {
	var adminConfiguration models.AdminConfiguration
	dbClient := database.DB
	if includeSsoConfigs {
		dbClient = dbClient.Preload("SSOConfigs")
	}
	err := dbClient.First(&adminConfiguration).Error
	if err != nil {
		return adminConfiguration, err
	}
	return adminConfiguration, nil
}

func GetSsoConfiguration() (bool, map[string]string, error) {

	adminConfiguration, err := GetAdminConfiguration(true)
	if err != nil {
		return false, nil, err
	}

	allowedSsoLogin := adminConfiguration.AllowSSOLogin
	if !allowedSsoLogin {
		return false, nil, errors.New("sso login not allowed")
	}
	response := map[string]string{}
	for _, ssoConfig := range adminConfiguration.SSOConfigs {
		platform := ssoConfig.Platform
		response[platform] = ssoConfig.ClientID
	}

	return true, response, nil
}

func UpdateAllowPasswordLogin(updatedValue bool) error {
	var adminConfiguration models.AdminConfiguration
	if err := database.DB.First(&adminConfiguration).Error; err != nil {
		return err
	}
	adminConfiguration.AllowPasswordLogin = updatedValue
	err := database.DB.Save(&adminConfiguration).Error
	if err != nil {
		return err
	}
	config.AllowPasswordLogin = updatedValue
	return nil
}

func UpdateAllowSSOLogin(updatedValue bool) error {
	var adminConfiguration models.AdminConfiguration
	if err := database.DB.First(&adminConfiguration).Error; err != nil {
		return err
	}
	adminConfiguration.AllowSSOLogin = updatedValue
	err := database.DB.Save(&adminConfiguration).Error
	if err != nil {
		return err
	}
	config.AllowSSOLogin = updatedValue
	return nil
}

func AddSsoConfig(domain, provider, clientId, clientSecret, platform string) error {

	adminConfiguration, err := GetAdminConfiguration(true)
	if err != nil {
		return err
	}
	var ssoConfig models.SSOConfig
	ssoConfig.UUID = uuid.NewString()
	ssoConfig.Platform = platform
	ssoConfig.Domain = domain
	ssoConfig.Provider = provider
	ssoConfig.ClientID = clientId
	ssoConfig.ClientSecret = clientSecret

	adminConfiguration.SSOConfigs = append(adminConfiguration.SSOConfigs, &ssoConfig)
	err = database.DB.Save(&adminConfiguration).Error
	if err != nil {
		return err
	}
	return nil
}

func DeleteSsoConfig(ssoConfigUuid string) error {
	var ssoConfig models.SSOConfig
	if err := database.DB.Where("uuid = ?", ssoConfigUuid).First(&ssoConfig).Error; err != nil {
		return err
	}
	err := database.DB.Delete(&ssoConfig).Error
	if err != nil {
		return err
	}

	return nil
}
