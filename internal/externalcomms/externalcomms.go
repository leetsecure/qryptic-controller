package externalcomms

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/leetsecure/qryptic-controller/internal/config"
	"github.com/leetsecure/qryptic-controller/internal/models"
	"github.com/leetsecure/qryptic-controller/internal/utils/logger"
)

func VpnGatewayHealthCheck(vpnGatewayDomain string) (string, error) {

	log := logger.Default()

	spaceClient := http.Client{
		Timeout: time.Second * 5,
	}
	vpnGatewayHealthCheckUrl := fmt.Sprintf(config.GatewayHealthCheckUrlTemplate, vpnGatewayDomain)
	req, err := http.NewRequest(http.MethodGet, vpnGatewayHealthCheckUrl, nil)
	if err != nil {
		log.Errorf("Error in creating request for %s", vpnGatewayHealthCheckUrl)
		return "", err
	}
	res, err := spaceClient.Do(req)
	if err != nil {
		log.Errorf("Error in executing request for %s", vpnGatewayHealthCheckUrl)
		return "", err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}
	if res.StatusCode != http.StatusOK {
		return "", errors.New("gateway is not healthy")
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Errorf("Error in reading response body from %s", vpnGatewayHealthCheckUrl)
		return "", err
	}

	return string(body), nil

}

func AddNewPeerInVpnGateway(vpnGatewayDomain string, authToken string, wgServerPeerConfigs []models.WGServerPeerConfig) (string, int, error) {
	log := logger.Default()

	spaceClient := http.Client{
		Timeout: time.Second * 5,
	}
	vpnGatewayAddPeersUrl := fmt.Sprintf("https://%s/controller/add-peers", vpnGatewayDomain)

	jsonWGServerPeerConfigs, err := json.Marshal(wgServerPeerConfigs)
	if err != nil {
		return "", 0, err
	}
	req, err := http.NewRequest(http.MethodPost, vpnGatewayAddPeersUrl, bytes.NewBuffer(jsonWGServerPeerConfigs))
	if err != nil {
		log.Errorf("Error in creating request for %s", vpnGatewayAddPeersUrl)
		return "", 0, err
	}

	authTokenWithBearer := fmt.Sprintf("Bearer %s", authToken)
	req.Header.Set("Authorization", authTokenWithBearer)
	req.Header.Set("Content-Type", "application/json")

	res, err := spaceClient.Do(req)
	if err != nil {
		log.Errorf("Error in executing request for %s", vpnGatewayAddPeersUrl)
		return "", 0, err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Errorf("Error in reading response body from %s", vpnGatewayAddPeersUrl)
		return "", 0, err
	}

	return string(body), res.StatusCode, nil
}

func DeletePeerInVpnGateway(vpnGatewayDomain string, authToken string, wgServerPeerConfigs []models.WGServerPeerConfig) (string, int, error) {
	log := logger.Default()

	spaceClient := http.Client{
		Timeout: time.Second * 5,
	}
	vpnGatewayAddPeersUrl := fmt.Sprintf("https://%s/controller/delete-peers", vpnGatewayDomain)

	jsonWGServerPeerConfigs, err := json.Marshal(wgServerPeerConfigs)
	if err != nil {
		return "", 0, err
	}
	req, err := http.NewRequest(http.MethodPost, vpnGatewayAddPeersUrl, bytes.NewBuffer(jsonWGServerPeerConfigs))
	if err != nil {
		log.Errorf("Error in creating request for %s", vpnGatewayAddPeersUrl)
		return "", 0, err
	}

	authTokenWithBearer := fmt.Sprintf("Bearer %s", authToken)
	req.Header.Set("Authorization", authTokenWithBearer)
	req.Header.Set("Content-Type", "application/json")

	res, err := spaceClient.Do(req)
	if err != nil {
		log.Errorf("Error in executing request for %s", vpnGatewayAddPeersUrl)
		return "", 0, err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Errorf("Error in reading response body from %s", vpnGatewayAddPeersUrl)
		return "", 0, err
	}

	return string(body), res.StatusCode, nil
}

func RestartVpnGateway(vpnGatewayDomain string, authToken string) (string, int, error) {
	log := logger.Default()

	spaceClient := http.Client{
		Timeout: time.Second * 5,
	}
	vpnGatewayRestartUrl := fmt.Sprintf("https://%s/controller/restart", vpnGatewayDomain)

	req, err := http.NewRequest(http.MethodPost, vpnGatewayRestartUrl, nil)
	if err != nil {
		log.Errorf("Error in creating request for %s", vpnGatewayRestartUrl)
		return "", 0, err
	}

	authTokenWithBearer := fmt.Sprintf("Bearer %s", authToken)
	req.Header.Set("Authorization", authTokenWithBearer)
	req.Header.Set("Content-Type", "application/json")

	res, err := spaceClient.Do(req)
	if err != nil {
		log.Errorf("Error in executing request for %s", vpnGatewayRestartUrl)
		return "", 0, err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Errorf("Error in reading response body from %s", vpnGatewayRestartUrl)
		return "", 0, err
	}

	return string(body), res.StatusCode, nil
}
