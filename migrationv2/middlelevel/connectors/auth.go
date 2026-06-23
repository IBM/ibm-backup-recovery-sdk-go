/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package connectors

// AuthMethod defines the authentication method type
type AuthMethod string

const (
	// Kubernetes authentication methods
	AuthMethodKubeconfig AuthMethod = "kubeconfig"
	AuthMethodToken      AuthMethod = "token"

	AuthMethodAPIKey AuthMethod = "apiKey"

	// VSI authentication methods
	AuthMethodSSHKey      AuthMethod = "ssh_key"
	AuthMethodSSHPassword AuthMethod = "ssh_password"
)

// AuthConfig holds authentication configuration for connector deployment
type AuthConfig interface {
	GetAuthMethod() AuthMethod
	Validate() error
	GetAPIKey() string
	GetIamURL() string
}

func (k *KubernetesAuthConfig) GetIamURL() string { return k.IamURL }
func (k *KubernetesAuthConfig) GetAPIKey() string { return k.ApiKey }

// KubernetesAuthConfig holds Kubernetes-specific authentication
type KubernetesAuthConfig struct {
	// AuthMethod specifies the authentication method
	AuthMethod AuthMethod `json:"authMethod"`

	// Kubeconfig-based authentication
	KubeconfigContent string `json:"kubeconfigContent,omitempty"` // Kubeconfig file content
	KubeconfigPath    string `json:"kubeconfigPath,omitempty"`    // Path to kubeconfig file
	Context           string `json:"context,omitempty"`           // Kubernetes context to use

	// Token-based authentication
	APIServer   string `json:"apiServer,omitempty"`   // Kubernetes API server URL
	BearerToken string `json:"bearerToken,omitempty"` // Bearer token for authentication
	CAData      string `json:"caData,omitempty"`

	// CA certificate data (base64 encoded)
	IamURL string `json:"url,omitempty"`

	//APIKeyBased
	ApiKey string `json:"apiKey,omitempty"` // apiKey of the cluster
	Host   string `json:"host,omitempty"`

	// Certificate-based authentication
	ClientCertData string `json:"clientCertData,omitempty"` // Client certificate (base64 encoded)
	ClientKeyData  string `json:"clientKeyData,omitempty"`  // Client key (base64 encoded)

	// Common settings
	Namespace      string `json:"namespace,omitempty"`      // Target namespace for deployment
	SkipTLSVerify  bool   `json:"skipTLSVerify,omitempty"`  // Skip TLS verification (not recommended)
	RequestTimeout int    `json:"requestTimeout,omitempty"` // Request timeout in seconds
}

// GetAuthMethod returns the authentication method
func (k *KubernetesAuthConfig) GetAuthMethod() AuthMethod {
	return k.AuthMethod
}

// Validate validates the Kubernetes authentication configuration
func (k *KubernetesAuthConfig) Validate() error {
	if k.IamURL == "" {
		return &ConnectorError{
			Code:    "INVALID_AUTH_CONFIG",
			Message: "Iam URL is required",
		}
	}
	switch k.AuthMethod {
	case AuthMethodKubeconfig:
		if k.KubeconfigContent == "" && k.KubeconfigPath == "" {
			return &ConnectorError{
				Code:    "INVALID_AUTH_CONFIG",
				Message: "either kubeconfigContent or kubeconfigPath is required for kubeconfig auth",
			}
		}
	case AuthMethodToken:
		if k.APIServer == "" {
			return &ConnectorError{
				Code:    "INVALID_AUTH_CONFIG",
				Message: "apiServer is required for token-based auth",
			}
		}
		if k.BearerToken == "" {
			return &ConnectorError{
				Code:    "INVALID_AUTH_CONFIG",
				Message: "bearerToken is required for token-based auth",
			}
		}

	case AuthMethodAPIKey:
		if k.Host == "" {
			return &ConnectorError{
				Code:    "INVALID_AUTH_CONFIG",
				Message: "host is required for api-based auth",
			}
		}

		if k.ApiKey == "" {
			return &ConnectorError{
				Code:    "INVALID_AUTH_CONFIG",
				Message: "apiKey is required for api-based auth",
			}
		}

	default:
		return &ConnectorError{
			Code:    "INVALID_AUTH_CONFIG",
			Message: "invalid authentication method for Kubernetes",
		}
	}
	return nil
}

// VSIAuthConfig holds VSI-specific authentication
type VSIAuthConfig struct {
	// AuthMethod specifies the authentication method
	AuthMethod AuthMethod `json:"authMethod"`

	// SSH connection details
	Host string `json:"host"` // VSI IP address or hostname
	Port int    `json:"port"` // SSH port (default: 22)
	User string `json:"user"` // SSH username

	// SSH Key-based authentication
	PrivateKey     string `json:"privateKey,omitempty"`     // SSH private key content
	PrivateKeyPath string `json:"privateKeyPath,omitempty"` // Path to SSH private key file
	Passphrase     string `json:"passphrase,omitempty"`     // Passphrase for encrypted private key

	// SSH Password-based authentication
	Password string `json:"password,omitempty"` // SSH password

	// SSH connection settings
	ConnectTimeout int  `json:"connectTimeout,omitempty"` // Connection timeout in seconds
	KeepAlive      int  `json:"keepAlive,omitempty"`      // Keep-alive interval in seconds
	StrictHostKey  bool `json:"strictHostKey,omitempty"`  // Strict host key checking

	// Sudo settings
	RequireSudo  bool   `json:"requireSudo,omitempty"`  // Whether sudo is required
	SudoPassword string `json:"sudoPassword,omitempty"` // Sudo password (if different from SSH password)
}

// GetAuthMethod returns the authentication method
func (v *VSIAuthConfig) GetAuthMethod() AuthMethod {
	return v.AuthMethod
}

func (v *VSIAuthConfig) GetIamURL() string { return "" }
func (v *VSIAuthConfig) GetAPIKey() string { return "" }

// Validate validates the VSI authentication configuration
func (v *VSIAuthConfig) Validate() error {
	// Validate common fields
	if v.Host == "" {
		return &ConnectorError{
			Code:    "INVALID_AUTH_CONFIG",
			Message: "host (IP address or hostname) is required",
		}
	}
	if v.User == "" {
		return &ConnectorError{
			Code:    "INVALID_AUTH_CONFIG",
			Message: "user is required",
		}
	}
	if v.Port == 0 {
		v.Port = 22 // Set default SSH port
	}

	// Validate auth method specific fields
	switch v.AuthMethod {
	case AuthMethodSSHKey:
		if v.PrivateKey == "" && v.PrivateKeyPath == "" {
			return &ConnectorError{
				Code:    "INVALID_AUTH_CONFIG",
				Message: "either privateKey or privateKeyPath is required for SSH key auth",
			}
		}
	case AuthMethodSSHPassword:
		if v.Password == "" {
			return &ConnectorError{
				Code:    "INVALID_AUTH_CONFIG",
				Message: "password is required for SSH password auth",
			}
		}
	default:
		return &ConnectorError{
			Code:    "INVALID_AUTH_CONFIG",
			Message: "invalid authentication method for VSI",
		}
	}

	return nil
}

// ConnectorError represents a connector-specific error
type ConnectorError struct {
	Code    string
	Message string
	Details map[string]interface{}
}

// Error implements the error interface
func (e *ConnectorError) Error() string {
	return e.Message
}

// NewConnectorError creates a new connector error
func NewConnectorError(code, message string) *ConnectorError {
	return &ConnectorError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// WithDetails adds details to the error
func (e *ConnectorError) WithDetails(key string, value interface{}) *ConnectorError {
	e.Details[key] = value
	return e
}
