/**
 * (C) Copyright IBM Corp. 2025.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/*
 * IBM OpenAPI SDK Code Generator Version: 3.96.1-5136e54a-20241108-203028
 */

// Package backuprecoveryv1 : Operations and models for the backuprecoveryv1 service
package backuprecoveryv1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type ManagementSreAuthenticator struct {
	authenticator AuthenticationMethods
}

type AuthenticationMethods interface {
	Authenticate(request *http.Request) error
	Validate() error
}

type ManagementSreApiKeyAuthenticator struct {
	ApiKey string
}

type ManagementSreBearerTokenAuthenticator struct {
	Username    string
	Password    string
	Url         string
	token       string
	expiryTime  int64
	issuedTime  int64
	refreshTime int64
	mu          sync.Mutex
	client      *http.Client
}

const (
	Accept              = "Accept"
	APPLICATION_JSON    = "application/json"
	CONTENT_DISPOSITION = "Content-Disposition"
	CONTENT_ENCODING    = "Content-Encoding"
	CONTENT_TYPE        = "Content-Type"
)

var (
	requestTokenMutex sync.Mutex
	needsRefreshMutex sync.Mutex
)

type ManagementSreAuthenticatorConfig struct {
	Username string
	Password string
	AuthUrl  string
	ApiKey   string
}

func NewManagementSreAuthenticator(managementSreAuthenticatorConfig *ManagementSreAuthenticatorConfig) (*ManagementSreAuthenticator, error) {

	if managementSreAuthenticatorConfig.ApiKey == "" && (managementSreAuthenticatorConfig.Username == "" || managementSreAuthenticatorConfig.Password == "" || managementSreAuthenticatorConfig.AuthUrl == "") {
		err := fmt.Errorf("The user should specify at least one of apiKey OR all of username, password and AuthUrl")
		return nil, err
	}

	if managementSreAuthenticatorConfig.ApiKey != "" {
		managementSreApiKeyAuthenticator := &ManagementSreApiKeyAuthenticator{ApiKey: managementSreAuthenticatorConfig.ApiKey}
		return &ManagementSreAuthenticator{authenticator: managementSreApiKeyAuthenticator}, nil
	}

	/*
		if ManagementSreBearerTokenAuthenticatorConfig.AuthUrl == "" {
			ManagementSreBearerTokenAuthenticatorConfig.AuthUrl = "https://manager.sre.backup-recovery.cloud.ibm.com/mcm/accessTokens"}
	*/
	managementSreBearerTokenAuthenticator := &ManagementSreBearerTokenAuthenticator{
		Username: managementSreAuthenticatorConfig.Username,
		Password: managementSreAuthenticatorConfig.Password,
		Url:      managementSreAuthenticatorConfig.AuthUrl,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
	return &ManagementSreAuthenticator{authenticator: managementSreBearerTokenAuthenticator}, nil

}

// Validate the authenticator's configuration.
func (managementSreAuthenticator *ManagementSreAuthenticator) Validate() error {

	err := managementSreAuthenticator.authenticator.Validate()
	if err != nil {
		return fmt.Errorf("Validation failed : %w", err)
	}
	return nil
}

func (authenticator *ManagementSreAuthenticator) Authenticate(request *http.Request) error {

	err := authenticator.authenticator.Authenticate(request)
	if err != nil {
		return fmt.Errorf("Authentication failed : %w", err)
	}

	return nil
}

func (authenticator *ManagementSreAuthenticator) AuthenticationType() string {
	return "managementSreAuth"
}

func (authenticator *ManagementSreApiKeyAuthenticator) Authenticate(request *http.Request) error {

	if authenticator.ApiKey != "" {
		request.Header.Set("apiKey", authenticator.ApiKey)
	}
	return nil
}

// Validate the authenticator's configuration.
// Ensures that the ApiKey and username/password properties are mutually exclusive,
func (authenticator *ManagementSreApiKeyAuthenticator) Validate() error {
	// The user should specify at least one of ApiKey or username/password.
	// Note: We'll allow both ApiKey and username/password to be specified,

	if authenticator.ApiKey != "" {
		return nil
	}
	err := fmt.Errorf("The user should specify at least one of apiKey OR username and password")
	return err
}

func (authenticator *ManagementSreBearerTokenAuthenticator) Authenticate(request *http.Request) error {
	// if token is not present or if the token is invalid
	if authenticator.getTokenData() == "" || time.Now().Unix() < authenticator.getTokenExpirationData() {
		// synchronously request the token
		err := authenticator.synchronouslyGetToken(request)
		if err != nil {
			return fmt.Errorf("failed to synchronously get token : %w", err)
		}
	} else if authenticator.needsRefresh() {
		// If refresh needed, kick off a go routine in the background to get a new token.
		//nolint: errcheck
		go authenticator.getToken()
	} else {
	}

	// return an error if the access token is not valid or was not fetched
	if authenticator.token == "" {
		err := errors.New("Error while trying to get token during Authenticate")
		return fmt.Errorf("failed to get token : %w", err)
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authenticator.token))
	return nil
}
func (authenticator *ManagementSreBearerTokenAuthenticator) synchronouslyGetToken(request *http.Request) error {
	requestTokenMutex.Lock()
	defer requestTokenMutex.Unlock()

	// Second check to prevent multiple queued quests from fetching tokens
	if authenticator.getTokenData() != "" && time.Now().Unix() < authenticator.getTokenExpirationData() {
		return nil
	}

	if err := authenticator.getToken(); err != nil {
		return fmt.Errorf("Failed to get token: %w", err)
	}

	return nil

}

func (authenticator *ManagementSreBearerTokenAuthenticator) getToken() error {

	body := map[string]interface{}{
		"username": authenticator.Username,
		"password": authenticator.Password,
		"domain":   "local",
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("Failed to marshal auth Body: %w", err)
	}

	req, err := http.NewRequest("POST", authenticator.Url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("Failed to create auth request: %w", err)
	}

	req.Header.Set(CONTENT_TYPE, APPLICATION_JSON)
	req.Header.Set(Accept, APPLICATION_JSON)

	res, err := authenticator.client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to make auth request: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("auth request failed with status: %s", res.Status)
	}

	var result struct {
		Token     string `json:"accessToken"`
		TokenType string `json:"tokenType"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}
	tokenExpiryTimeInHours := time.Duration(24)

	currentTime := time.Now()
	expiryTimeSec := currentTime.Add(tokenExpiryTimeInHours * time.Hour).Unix()
	issuedTimeSec := currentTime.Unix()

	authenticator.mu.Lock()
	defer authenticator.mu.Unlock()

	authenticator.token = result.Token
	authenticator.expiryTime = expiryTimeSec
	authenticator.issuedTime = issuedTimeSec
	authenticator.refreshTime = int64(float64(expiryTimeSec) - float64(expiryTimeSec-issuedTimeSec)*(0.2))

	return nil

}

// getTokenData returns the tokenData field from the authenticator.
func (authenticator *ManagementSreBearerTokenAuthenticator) getTokenData() string {
	authenticator.mu.Lock()
	defer authenticator.mu.Unlock()

	return authenticator.token
}

func (authenticator *ManagementSreBearerTokenAuthenticator) getTokenExpirationData() int64 {
	authenticator.mu.Lock()
	defer authenticator.mu.Unlock()

	return authenticator.expiryTime
}

func (authenticator *ManagementSreBearerTokenAuthenticator) needsRefresh() bool {

	needsRefreshMutex.Lock()
	defer needsRefreshMutex.Unlock()

	// Advance refresh by one minute
	if authenticator.refreshTime >= 0 && time.Now().Unix() > authenticator.refreshTime {
		authenticator.refreshTime = time.Now().Unix() + 90

		return true
	}
	return false
}

// Validate the authenticator's configuration.
// Ensures that the ApiKey and username/password properties are mutually exclusive,
func (authenticator *ManagementSreBearerTokenAuthenticator) Validate() error {
	// The user should specify both ApiKey and URL.
	if authenticator.Username == "" || authenticator.Password == "" || authenticator.Url == "" {
		err := fmt.Errorf("The user should specify all username, password and authentication URL.")
		return err
	}
	return nil
}
