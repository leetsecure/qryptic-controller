package config

import (
	"errors"
	"os"
	"time"

	"github.com/leetsecure/qryptic-controller/internal/utils/logger"
)

var Environment = "production"
var ClientExpiry = 60 * 4 * time.Minute
var JwtTokenTimeout = 60 * time.Minute
var SSOStateJwtTokenTimeout = 5 * time.Minute
var SSOCallbackTemplate = "https://%s/api/v1/auth/%s/web/sso/callback"
var VpnGatewayApplicationImageName = "940482412786.dkr.ecr.ap-south-1.amazonaws.com/qryptic/gateway:<version>"
var GatewayHealthCheckUrlTemplate = "https://%s/health"
var GatewayCallbackForConfigTemplate = "https://%s/api/v1/gateway/get-gateway-config"
var GatewayPort = "8080"
var WireguardPort = "51820"

var AllowPasswordLogin = true
var AllowSSOLogin = false

var SsoConfig = []map[string]string{}
var GoogleClientID = ""
var GoogleClientSecret = ""

var (
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSslMode  string
)

var (
	TempUserCreated bool
	TempUserActive  bool
)

var (
	UserAuthJwtSecretKey    string
	UserAuthSSOJwtSecretKey string
)

var (
	ControllerDomain string
)

func UpdateEnvConfig() error {
	log := logger.Default()
	var err error
	dBHost, exists := os.LookupEnv("DBHost")
	if !exists {
		err = errors.Join(err, errors.New("required environment variables not present:DBHost"))
	}

	dBPort, exists := os.LookupEnv("DBPort")
	if !exists {
		err = errors.Join(err, errors.New("required environment variables not present:DBPort"))
	}

	dBUser, exists := os.LookupEnv("DBUser")
	if !exists {
		err = errors.Join(err, errors.New("required environment variables not present:DBUser"))
	}

	dBPassword, exists := os.LookupEnv("DBPassword")
	if !exists {
		err = errors.Join(err, errors.New("required environment variables not present:DBPassword"))
	}

	dBName, exists := os.LookupEnv("DBName")
	if !exists {
		err = errors.Join(err, errors.New("required environment variables not present:DBName"))
	}

	dBSslMode, exists := os.LookupEnv("DBSslMode")
	if !exists {
		err = errors.Join(err, errors.New("required environment variables not present:DBSslMode"))
	}

	controllerDomain, exists := os.LookupEnv("ControllerDomain")
	if !exists {
		err = errors.Join(err, errors.New("required environment variables not present:ControllerDomain"))
	}

	environment, exists := os.LookupEnv("Environment")
	if exists {
		if !((environment == "production") || (environment == "development") || (environment == "local")) {
			err = errors.Join(err, errors.New("required environment variables is incorrect:Enviroment"))
		}
		log.Infof("Environment : %s activated", environment)
		Environment = environment
	} else {
		log.Info("Environment : production activated")
	}

	if err != nil {
		return err
	}

	DBHost = dBHost
	DBPort = dBPort
	DBUser = dBUser
	DBPassword = dBPassword
	DBName = dBName
	DBSslMode = dBSslMode
	ControllerDomain = controllerDomain

	if Environment == "local" {
		ClientExpiry = 10 * time.Minute
		JwtTokenTimeout = 2 * 60 * time.Minute
		SSOCallbackTemplate = "http://%s/api/v1/auth/%s/web/sso/callback"
	}

	if Environment == "development" {
		ClientExpiry = 60 * time.Minute
		JwtTokenTimeout = 60 * time.Minute
		SSOCallbackTemplate = "https://%s/api/v1/auth/%s/web/sso/callback"
	}

	return nil
}
