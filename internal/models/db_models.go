package models

import (
	"time"

	"gorm.io/gorm"
)

type AdminConfiguration struct {
	gorm.Model
	UUID                    string       `json:"uuid" gorm:"uniqueIndex"`
	AllowPasswordLogin      bool         `json:"allowPasswordLogin"`
	AllowSSOLogin           bool         `json:"allowSSOLogin"`
	UserAuthJwtSecretKey    string       `json:"-"`
	UserAuthSSOJwtSecretKey string       `json:"-"`
	UserJWTAlgorithm        string       `json:"-"`
	GatewayJWTAlgorithm     string       `json:"-"`
	TempUserCreated         bool         `json:"-"`
	TempUserActive          bool         `json:"tempUserActive"`
	SSOConfigs              []*SSOConfig `json:"ssoConfigs" gorm:"foreignKey:AdminConfigurationID"`
}

type SSOConfig struct {
	gorm.Model
	UUID                 string `json:"uuid" gorm:"uniqueIndex"`
	Enabled              string `json:"enabled" gorm:"default:true"`
	Domain               string `json:"domain"`   // Email domain for SSO
	Provider             string `json:"provider"` // SSO provider (e.g., "google", "microsoft")
	Platform             string `json:"platform"`
	ClientID             string `json:"clientID"`
	ClientSecret         string `json:"clientSecret"`
	AdminConfigurationID uint   `json:"adminConfigurationID"` // Foreign key to VpnGateway
}

// User DB model
type User struct {
	gorm.Model
	UUID          string        `json:"uuid" gorm:"uniqueIndex"`
	Name          string        `json:"name" gorm:"name"`
	Email         string        `json:"email" gorm:"index"`
	IsPasswordSet bool          `json:"isPasswordSet"`
	PasswordHash  string        `json:"-"` // Omit password hash from JSON output
	Role          UserRoleEnum  `json:"role"`
	Clients       []*Client     `json:"clients"`
	VpnGateways   []*VpnGateway `json:"vpnGateways" gorm:"many2many:user_vpngateways;"`
	Groups        []*Group      `json:"groups" gorm:"many2many:group_users;"`
}

// VPN Gateway DB model
type VpnGateway struct {
	gorm.Model
	UUID             string    `json:"uuid" gorm:"uniqueIndex"`
	Name             string    `json:"name" `
	JwtSecretKey     string    `json:"jwtSecretKey"`
	JwtAlgorithm     string    `json:"jwtAlgorithm"`
	ServerPublicKey  string    `json:"serverPublicKey"`
	ServerPrivateKey string    `json:"serverPrivateKey"`
	Domain           string    `json:"domain"`
	IpAddress        string    `json:"ipAddressCIDR"`
	VpnCIDR          string    `json:"vpnCIDR"`
	Port             int       `json:"port"`
	DnsServer        string    `json:"dnsServer"`
	Clients          []*Client `json:"clients"`
	Users            []*User   `json:"users" gorm:"many2many:user_vpngateways;"`
	Groups           []*Group  `json:"groups" gorm:"many2many:group_vpngateways;"`
	// IPAllocations    []*IPAllocation `json:"ipAllocations"`
	IPPool []IPPool `json:"ipPool"`
}

type UserRoleEnum string

const (
	DefaultRole UserRoleEnum = "Default"
	AdminRole   UserRoleEnum = "Admin"
	UserRole    UserRoleEnum = "User"
)

type ConnectivityThroughEnum string

const (
	Domain    ConnectivityThroughEnum = "Domain"
	IpAddress ConnectivityThroughEnum = "IpAddress"
)

type Client struct {
	gorm.Model
	UUID             string      `json:"uuid" gorm:"uniqueIndex"`
	UserID           uint        `json:"userId" gorm:"index"`
	User             *User       `json:"user" gorm:"foreignKey:UserID"`
	VpnGatewayID     uint        `json:"vpnGatewayId" gorm:"index"`
	VpnGateway       *VpnGateway `json:"vpnGateway" gorm:"foreignKey:VpnGatewayID"`
	ClientPublicKey  string      `json:"clientPublicKey"`
	ClientPrivateKey string      `json:"clientPrivateKey"`
	PresharedKey     string      `json:"preshared_key"`
	ExpiryTime       time.Time   `json:"expiryTime"`
	IsActive         bool        `json:"is_active"`
	AllocatedIP      string      `json:"allocatedIP"`
	AllowedIPs       string      `json:"allowedIPs"`
	DnsServer        string      `json:"dnsServer"`

	// IPAllocated      *IPAllocation `json:"ipAllocated" gorm:"foreignKey:ClientID"`
}

type AuditTrail struct {
	gorm.Model
	UUID         string      `json:"uuid" gorm:"uniqueIndex"`
	UserID       *uint       `json:"userId" gorm:"index"`
	User         *User       `json:"user" gorm:"foreignKey:UserID"`
	Action       string      `json:"action"`
	Description  string      `json:"description"`
	Timestamp    time.Time   `json:"timestamp"`
	VpnGatewayID *uint       `json:"vpnGatewayId"`
	VpnGateway   *VpnGateway `json:"vpnGateway" gorm:"foreignKey:VpnGatewayID"`
	ClientID     *uint       `json:"clientId" gorm:"index"`
	Client       *Client     `json:"client" gorm:"foreignKey:ClientID"`
}

type IPPool struct {
	gorm.Model
	UUID         string `json:"uuid" gorm:"uniqueIndex"`
	IP           string `json:"ip" `
	Assigned     bool   `json:"assigned" gorm:"index"`
	VpnGatewayID uint   `json:"vpngatewayID" gorm:"index"` // Foreign key to VpnGateway

}

type Group struct {
	gorm.Model
	UUID        string        `json:"uuid" gorm:"uniqueIndex"`
	Name        string        `json:"name"`
	Users       []*User       `json:"users" gorm:"many2many:group_users;"`
	VpnGateways []*VpnGateway `json:"vpnGateways" gorm:"many2many:group_vpngateways;"`
}

type Auth struct {
	gorm.Model
	UUID          string    `json:"uuid" gorm:"uniqueIndex"`
	Provider      string    `json:"provider"`
	State         string    `json:"state" gorm:"index"`
	CodeChallenge string    `json:"codeChallenge" gorm:"index"`
	ExpiryTime    time.Time `json:"expiryTime"`
	Email         string    `json:"email"`
	Authenticated bool      `json:"authenticated"`
}
