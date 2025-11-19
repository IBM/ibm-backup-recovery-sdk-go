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
 * IBM OpenAPI SDK Code Generator Version: 3.105.1-067d600b-20250616-154447
 */

// Package backuprecovery : Operations and models for the backuprecovery service
package backuprecoveryv1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
	common "github.com/IBM/ibm-backup-recovery-sdk-go/common"
)

// DefaultServiceURL is the default URL to make service requests to.
const DefaulManagementSreApiServiceURL = "https://management-sre-api.cloud.ibm.com/v2"

// DefaultServiceName is the default key used to find external configuration information.
const DefaultManagementSreServiceName = "backup_recovery_management_sre_api"

// BackupRecoveryManagementSreApiV1 : IBM API provides a RESTful interface to access the various data management operations on IBM
// cluster and Management Console.
//
// API Version: 2.0
type BackupRecoveryManagementSreApiV1 struct {
	Service *core.BaseService
}

// BackupRecoveryManagementSreApiV1Options : Service options
type BackupRecoveryManagementSreApiV1Options struct {
	ServiceName   string
	URL           string
	Authenticator core.Authenticator
}

// NewBackupRecoveryManagementSreApiV1UsingExternalConfig : constructs an instance of BackupRecoveryManagementSreApiV1 with passed in options and external configuration.
func NewBackupRecoveryManagementSreApiV1UsingExternalConfig(options *BackupRecoveryManagementSreApiV1Options) (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1, err error) {
	if options.ServiceName == "" {
		options.ServiceName = DefaultManagementSreServiceName
	}

	if options.Authenticator == nil {
		options.Authenticator, err = core.GetAuthenticatorFromEnvironment(options.ServiceName)
		if err != nil {
			err = core.SDKErrorf(err, "", "env-auth-error", common.GetComponentInfo())
			return
		}
	}

	backupRecoveryManagementSreApi, err = NewBackupRecoveryManagementSreApiV1(options)
	err = core.RepurposeSDKProblem(err, "new-client-error")
	if err != nil {
		return
	}

	err = backupRecoveryManagementSreApi.Service.ConfigureService(options.ServiceName)
	if err != nil {
		err = core.SDKErrorf(err, "", "client-config-error", common.GetComponentInfo())
		return
	}

	if options.URL != "" {
		err = backupRecoveryManagementSreApi.Service.SetServiceURL(options.URL)
		err = core.RepurposeSDKProblem(err, "url-set-error")
	}
	return
}

// NewBackupRecoveryManagementSreApiV1 : constructs an instance of BackupRecoveryManagementSreApiV1 with passed in options.
func NewBackupRecoveryManagementSreApiV1(options *BackupRecoveryManagementSreApiV1Options) (service *BackupRecoveryManagementSreApiV1, err error) {
	serviceOptions := &core.ServiceOptions{
		URL:           DefaulManagementSreApiServiceURL,
		Authenticator: options.Authenticator,
	}

	baseService, err := core.NewBaseService(serviceOptions)
	if err != nil {
		err = core.SDKErrorf(err, "", "new-base-error", common.GetComponentInfo())
		return
	}

	if options.URL != "" {
		err = baseService.SetServiceURL(options.URL)
		if err != nil {
			err = core.SDKErrorf(err, "", "set-url-error", common.GetComponentInfo())
			return
		}
	}

	service = &BackupRecoveryManagementSreApiV1{
		Service: baseService,
	}

	return
}

// GetServiceURLForRegion returns the service URL to be used for the specified region
func GetManagementSreApiServiceURLForRegion(region string) (string, error) {
	return "", core.SDKErrorf(nil, "service does not support regional URLs", "no-regional-support", common.GetComponentInfo())
}

// Clone makes a copy of "backupRecoveryManagementSreApi" suitable for processing requests.
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) Clone() *BackupRecoveryManagementSreApiV1 {
	if core.IsNil(backupRecoveryManagementSreApi) {
		return nil
	}
	clone := *backupRecoveryManagementSreApi
	clone.Service = backupRecoveryManagementSreApi.Service.Clone()
	return &clone
}

// SetServiceURL sets the service URL
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) SetServiceURL(url string) error {
	err := backupRecoveryManagementSreApi.Service.SetServiceURL(url)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-set-error", common.GetComponentInfo())
	}
	return err
}

// GetServiceURL returns the service URL
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) GetServiceURL() string {
	return backupRecoveryManagementSreApi.Service.GetServiceURL()
}

// SetDefaultHeaders sets HTTP headers to be sent in every request
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) SetDefaultHeaders(headers http.Header) {
	backupRecoveryManagementSreApi.Service.SetDefaultHeaders(headers)
}

// SetEnableGzipCompression sets the service's EnableGzipCompression field
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) SetEnableGzipCompression(enableGzip bool) {
	backupRecoveryManagementSreApi.Service.SetEnableGzipCompression(enableGzip)
}

// GetEnableGzipCompression returns the service's EnableGzipCompression field
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) GetEnableGzipCompression() bool {
	return backupRecoveryManagementSreApi.Service.GetEnableGzipCompression()
}

// EnableRetries enables automatic retries for requests invoked for this service instance.
// If either parameter is specified as 0, then a default value is used instead.
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) EnableRetries(maxRetries int, maxRetryInterval time.Duration) {
	backupRecoveryManagementSreApi.Service.EnableRetries(maxRetries, maxRetryInterval)
}

// DisableRetries disables automatic retries for requests invoked for this service instance.
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) DisableRetries() {
	backupRecoveryManagementSreApi.Service.DisableRetries()
}

// GetAlerts : Get alerts
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) GetAlerts(getAlertsOptions *GetAlertsOptions) (result *AlertList, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementSreApi.GetAlertsWithContext(context.Background(), getAlertsOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// GetAlertsWithContext is an alternate form of the GetAlerts method which supports a Context parameter
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) GetAlertsWithContext(ctx context.Context, getAlertsOptions *GetAlertsOptions) (result *AlertList, response *core.DetailedResponse, err error) {
	err = core.ValidateStruct(getAlertsOptions, "getAlertsOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	builder := core.NewRequestBuilder(core.GET)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementSreApi.GetEnableGzipCompression()
	_, err = builder.ResolveRequestURL(backupRecoveryManagementSreApi.Service.Options.URL, `/alerts`, nil)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_sre_api", "V2", "GetAlerts")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range getAlertsOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")
	if getAlertsOptions.XScopeIdentifier != nil {
		builder.AddHeader("X-Scope-Identifier", fmt.Sprint(*getAlertsOptions.XScopeIdentifier))
	}

	if getAlertsOptions.AlertIds != nil {
		builder.AddQuery("alertIds", strings.Join(getAlertsOptions.AlertIds, ","))
	}
	if getAlertsOptions.AlertTypes != nil {
		err = builder.AddQuerySlice("alertTypes", getAlertsOptions.AlertTypes)
		if err != nil {
			err = core.SDKErrorf(err, "", "add-query-slice-error", common.GetComponentInfo())
			return
		}
	}
	if getAlertsOptions.AlertCategories != nil {
		builder.AddQuery("alertCategories", strings.Join(getAlertsOptions.AlertCategories, ","))
	}
	if getAlertsOptions.AlertStates != nil {
		builder.AddQuery("alertStates", strings.Join(getAlertsOptions.AlertStates, ","))
	}
	if getAlertsOptions.AlertSeverities != nil {
		builder.AddQuery("alertSeverities", strings.Join(getAlertsOptions.AlertSeverities, ","))
	}
	if getAlertsOptions.AlertTypeBuckets != nil {
		builder.AddQuery("alertTypeBuckets", strings.Join(getAlertsOptions.AlertTypeBuckets, ","))
	}
	if getAlertsOptions.StartTimeUsecs != nil {
		builder.AddQuery("startTimeUsecs", fmt.Sprint(*getAlertsOptions.StartTimeUsecs))
	}
	if getAlertsOptions.EndTimeUsecs != nil {
		builder.AddQuery("endTimeUsecs", fmt.Sprint(*getAlertsOptions.EndTimeUsecs))
	}
	if getAlertsOptions.MaxAlerts != nil {
		builder.AddQuery("maxAlerts", fmt.Sprint(*getAlertsOptions.MaxAlerts))
	}
	if getAlertsOptions.PropertyKey != nil {
		builder.AddQuery("propertyKey", fmt.Sprint(*getAlertsOptions.PropertyKey))
	}
	if getAlertsOptions.PropertyValue != nil {
		builder.AddQuery("propertyValue", fmt.Sprint(*getAlertsOptions.PropertyValue))
	}
	if getAlertsOptions.AlertName != nil {
		builder.AddQuery("alertName", fmt.Sprint(*getAlertsOptions.AlertName))
	}
	if getAlertsOptions.ResolutionIds != nil {
		err = builder.AddQuerySlice("resolutionIds", getAlertsOptions.ResolutionIds)
		if err != nil {
			err = core.SDKErrorf(err, "", "add-query-slice-error", common.GetComponentInfo())
			return
		}
	}
	if getAlertsOptions.TenantIds != nil {
		builder.AddQuery("tenantIds", strings.Join(getAlertsOptions.TenantIds, ","))
	}
	if getAlertsOptions.AllUnderHierarchy != nil {
		builder.AddQuery("allUnderHierarchy", fmt.Sprint(*getAlertsOptions.AllUnderHierarchy))
	}

	request, err := builder.Build()
	if err != nil {
		err = core.SDKErrorf(err, "", "build-error", common.GetComponentInfo())
		return
	}

	var rawResponse map[string]json.RawMessage
	response, err = backupRecoveryManagementSreApi.Service.Request(request, &rawResponse)
	if err != nil {
		core.EnrichHTTPProblem(err, "GetAlerts", getManagementSreServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}
	if rawResponse != nil {
		err = core.UnmarshalModel(rawResponse, "", &result, UnmarshalAlertList)
		if err != nil {
			err = core.SDKErrorf(err, "", "unmarshal-resp-error", common.GetComponentInfo())
			return
		}
		response.Result = result
	}

	return
}

// GetAlertSummary : Get alerts summary
// Get alerts summary grouped by category.
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) GetAlertSummary(getAlertSummaryOptions *GetAlertSummaryOptions) (result *AlertsSummaryResponse, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementSreApi.GetAlertSummaryWithContext(context.Background(), getAlertSummaryOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// GetAlertSummaryWithContext is an alternate form of the GetAlertSummary method which supports a Context parameter
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) GetAlertSummaryWithContext(ctx context.Context, getAlertSummaryOptions *GetAlertSummaryOptions) (result *AlertsSummaryResponse, response *core.DetailedResponse, err error) {
	err = core.ValidateStruct(getAlertSummaryOptions, "getAlertSummaryOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	builder := core.NewRequestBuilder(core.GET)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementSreApi.GetEnableGzipCompression()
	_, err = builder.ResolveRequestURL(backupRecoveryManagementSreApi.Service.Options.URL, `/alerts-summary`, nil)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_sre_api", "V2", "GetAlertSummary")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range getAlertSummaryOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")
	if getAlertSummaryOptions.XScopeIdentifier != nil {
		builder.AddHeader("X-Scope-Identifier", fmt.Sprint(*getAlertSummaryOptions.XScopeIdentifier))
	}

	if getAlertSummaryOptions.StartTimeUsecs != nil {
		builder.AddQuery("startTimeUsecs", fmt.Sprint(*getAlertSummaryOptions.StartTimeUsecs))
	}
	if getAlertSummaryOptions.EndTimeUsecs != nil {
		builder.AddQuery("endTimeUsecs", fmt.Sprint(*getAlertSummaryOptions.EndTimeUsecs))
	}
	if getAlertSummaryOptions.IncludeTenants != nil {
		builder.AddQuery("includeTenants", fmt.Sprint(*getAlertSummaryOptions.IncludeTenants))
	}
	if getAlertSummaryOptions.TenantIds != nil {
		builder.AddQuery("tenantIds", strings.Join(getAlertSummaryOptions.TenantIds, ","))
	}
	if getAlertSummaryOptions.StatesList != nil {
		builder.AddQuery("statesList", strings.Join(getAlertSummaryOptions.StatesList, ","))
	}

	request, err := builder.Build()
	if err != nil {
		err = core.SDKErrorf(err, "", "build-error", common.GetComponentInfo())
		return
	}

	var rawResponse map[string]json.RawMessage
	response, err = backupRecoveryManagementSreApi.Service.Request(request, &rawResponse)
	if err != nil {
		core.EnrichHTTPProblem(err, "GetAlertSummary", getManagementSreServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}
	if rawResponse != nil {
		err = core.UnmarshalModel(rawResponse, "", &result, UnmarshalAlertsSummaryResponse)
		if err != nil {
			err = core.SDKErrorf(err, "", "unmarshal-resp-error", common.GetComponentInfo())
			return
		}
		response.Result = result
	}

	return
}

// GetManagementAlertsSummary : Get alerts summary on Management Console
// Get alerts summary grouped by category.
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) GetManagementAlertsSummary(getManagementAlertsSummaryOptions *GetManagementAlertsSummaryOptions) (result *AlertsSummaryResponse, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementSreApi.GetManagementAlertsSummaryWithContext(context.Background(), getManagementAlertsSummaryOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// GetManagementAlertsSummaryWithContext is an alternate form of the GetManagementAlertsSummary method which supports a Context parameter
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) GetManagementAlertsSummaryWithContext(ctx context.Context, getManagementAlertsSummaryOptions *GetManagementAlertsSummaryOptions) (result *AlertsSummaryResponse, response *core.DetailedResponse, err error) {
	err = core.ValidateStruct(getManagementAlertsSummaryOptions, "getManagementAlertsSummaryOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	builder := core.NewRequestBuilder(core.GET)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementSreApi.GetEnableGzipCompression()
	_, err = builder.ResolveRequestURL(backupRecoveryManagementSreApi.Service.Options.URL, `/mcm/stats/alerts-summary`, nil)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_sre_api", "V2", "GetManagementAlertsSummary")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range getManagementAlertsSummaryOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")

	if getManagementAlertsSummaryOptions.ClusterIdentifiers != nil {
		builder.AddQuery("clusterIdentifiers", strings.Join(getManagementAlertsSummaryOptions.ClusterIdentifiers, ","))
	}
	if getManagementAlertsSummaryOptions.StartTimeUsecs != nil {
		builder.AddQuery("startTimeUsecs", fmt.Sprint(*getManagementAlertsSummaryOptions.StartTimeUsecs))
	}
	if getManagementAlertsSummaryOptions.EndTimeUsecs != nil {
		builder.AddQuery("endTimeUsecs", fmt.Sprint(*getManagementAlertsSummaryOptions.EndTimeUsecs))
	}
	if getManagementAlertsSummaryOptions.StatesList != nil {
		builder.AddQuery("statesList", strings.Join(getManagementAlertsSummaryOptions.StatesList, ","))
	}

	request, err := builder.Build()
	if err != nil {
		err = core.SDKErrorf(err, "", "build-error", common.GetComponentInfo())
		return
	}

	var rawResponse map[string]json.RawMessage
	response, err = backupRecoveryManagementSreApi.Service.Request(request, &rawResponse)
	if err != nil {
		core.EnrichHTTPProblem(err, "GetManagementAlertsSummary", getManagementSreServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}
	if rawResponse != nil {
		err = core.UnmarshalModel(rawResponse, "", &result, UnmarshalAlertsSummaryResponse)
		if err != nil {
			err = core.SDKErrorf(err, "", "unmarshal-resp-error", common.GetComponentInfo())
			return
		}
		response.Result = result
	}

	return
}

// GetManagementAlerts : Get list of management console alerts
// Get the list of management console alerts.
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) GetManagementAlerts(getManagementAlertsOptions *GetManagementAlertsOptions) (result *AlertsList, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementSreApi.GetManagementAlertsWithContext(context.Background(), getManagementAlertsOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// GetManagementAlertsWithContext is an alternate form of the GetManagementAlerts method which supports a Context parameter
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) GetManagementAlertsWithContext(ctx context.Context, getManagementAlertsOptions *GetManagementAlertsOptions) (result *AlertsList, response *core.DetailedResponse, err error) {
	err = core.ValidateStruct(getManagementAlertsOptions, "getManagementAlertsOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	builder := core.NewRequestBuilder(core.GET)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementSreApi.GetEnableGzipCompression()
	_, err = builder.ResolveRequestURL(backupRecoveryManagementSreApi.Service.Options.URL, `/mcm/alerts`, nil)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_sre_api", "V2", "GetManagementAlerts")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range getManagementAlertsOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")

	if getManagementAlertsOptions.AlertIdList != nil {
		builder.AddQuery("alertIdList", strings.Join(getManagementAlertsOptions.AlertIdList, ","))
	}
	if getManagementAlertsOptions.AlertStateList != nil {
		builder.AddQuery("alertStateList", strings.Join(getManagementAlertsOptions.AlertStateList, ","))
	}
	if getManagementAlertsOptions.AlertTypeList != nil {
		err = builder.AddQuerySlice("alertTypeList", getManagementAlertsOptions.AlertTypeList)
		if err != nil {
			err = core.SDKErrorf(err, "", "add-query-slice-error", common.GetComponentInfo())
			return
		}
	}
	if getManagementAlertsOptions.AlertSeverityList != nil {
		builder.AddQuery("alertSeverityList", strings.Join(getManagementAlertsOptions.AlertSeverityList, ","))
	}
	if getManagementAlertsOptions.RegionIds != nil {
		builder.AddQuery("regionIds", strings.Join(getManagementAlertsOptions.RegionIds, ","))
	}
	if getManagementAlertsOptions.ClusterIdentifiers != nil {
		builder.AddQuery("clusterIdentifiers", strings.Join(getManagementAlertsOptions.ClusterIdentifiers, ","))
	}
	if getManagementAlertsOptions.StartDateUsecs != nil {
		builder.AddQuery("startDateUsecs", fmt.Sprint(*getManagementAlertsOptions.StartDateUsecs))
	}
	if getManagementAlertsOptions.EndDateUsecs != nil {
		builder.AddQuery("endDateUsecs", fmt.Sprint(*getManagementAlertsOptions.EndDateUsecs))
	}
	if getManagementAlertsOptions.MaxAlerts != nil {
		builder.AddQuery("maxAlerts", fmt.Sprint(*getManagementAlertsOptions.MaxAlerts))
	}
	if getManagementAlertsOptions.AlertCategoryList != nil {
		builder.AddQuery("alertCategoryList", strings.Join(getManagementAlertsOptions.AlertCategoryList, ","))
	}
	if getManagementAlertsOptions.TenantIds != nil {
		builder.AddQuery("tenantIds", strings.Join(getManagementAlertsOptions.TenantIds, ","))
	}
	if getManagementAlertsOptions.AlertTypeBucketList != nil {
		builder.AddQuery("alertTypeBucketList", strings.Join(getManagementAlertsOptions.AlertTypeBucketList, ","))
	}
	if getManagementAlertsOptions.AlertPropertyKeyList != nil {
		builder.AddQuery("alertPropertyKeyList", strings.Join(getManagementAlertsOptions.AlertPropertyKeyList, ","))
	}
	if getManagementAlertsOptions.AlertPropertyValueList != nil {
		builder.AddQuery("alertPropertyValueList", strings.Join(getManagementAlertsOptions.AlertPropertyValueList, ","))
	}
	if getManagementAlertsOptions.AlertName != nil {
		builder.AddQuery("alertName", fmt.Sprint(*getManagementAlertsOptions.AlertName))
	}
	if getManagementAlertsOptions.ServiceInstanceIds != nil {
		builder.AddQuery("serviceInstanceIds", strings.Join(getManagementAlertsOptions.ServiceInstanceIds, ","))
	}

	request, err := builder.Build()
	if err != nil {
		err = core.SDKErrorf(err, "", "build-error", common.GetComponentInfo())
		return
	}

	var rawResponse map[string]json.RawMessage
	response, err = backupRecoveryManagementSreApi.Service.Request(request, &rawResponse)
	if err != nil {
		core.EnrichHTTPProblem(err, "GetManagementAlerts", getManagementSreServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}
	if rawResponse != nil {
		err = core.UnmarshalModel(rawResponse, "", &result, UnmarshalAlertsList)
		if err != nil {
			err = core.SDKErrorf(err, "", "unmarshal-resp-error", common.GetComponentInfo())
			return
		}
		response.Result = result
	}

	return
}

// GetManagementAlertResolution : List the Alert Resolutions in BRS system
// Returns the Management Console Alert Resolution objects found in IBM system that match the filter criteria from given
// parameters. If no filter parameters are specified, MaxResolutions Alert Resolution objects are returned. Each object
// provides details about the Alert Resolution info and the resolved alert info.
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) GetManagementAlertResolution(getManagementAlertResolutionOptions *GetManagementAlertResolutionOptions) (result *AlertResolutionsList, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementSreApi.GetManagementAlertResolutionWithContext(context.Background(), getManagementAlertResolutionOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// GetManagementAlertResolutionWithContext is an alternate form of the GetManagementAlertResolution method which supports a Context parameter
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) GetManagementAlertResolutionWithContext(ctx context.Context, getManagementAlertResolutionOptions *GetManagementAlertResolutionOptions) (result *AlertResolutionsList, response *core.DetailedResponse, err error) {
	err = core.ValidateNotNil(getManagementAlertResolutionOptions, "getManagementAlertResolutionOptions cannot be nil")
	if err != nil {
		err = core.SDKErrorf(err, "", "unexpected-nil-param", common.GetComponentInfo())
		return
	}
	err = core.ValidateStruct(getManagementAlertResolutionOptions, "getManagementAlertResolutionOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	builder := core.NewRequestBuilder(core.GET)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementSreApi.GetEnableGzipCompression()
	_, err = builder.ResolveRequestURL(backupRecoveryManagementSreApi.Service.Options.URL, `/mcm/alerts/resolutions`, nil)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_sre_api", "V2", "GetManagementAlertResolution")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range getManagementAlertResolutionOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")

	builder.AddQuery("maxResolutions", fmt.Sprint(*getManagementAlertResolutionOptions.MaxResolutions))
	if getManagementAlertResolutionOptions.ResolutionName != nil {
		builder.AddQuery("resolutionName", fmt.Sprint(*getManagementAlertResolutionOptions.ResolutionName))
	}
	if getManagementAlertResolutionOptions.ResolutionID != nil {
		builder.AddQuery("resolutionId", fmt.Sprint(*getManagementAlertResolutionOptions.ResolutionID))
	}

	request, err := builder.Build()
	if err != nil {
		err = core.SDKErrorf(err, "", "build-error", common.GetComponentInfo())
		return
	}

	var rawResponse map[string]json.RawMessage
	response, err = backupRecoveryManagementSreApi.Service.Request(request, &rawResponse)
	if err != nil {
		core.EnrichHTTPProblem(err, "GetManagementAlertResolution", getManagementSreServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}
	if rawResponse != nil {
		err = core.UnmarshalModel(rawResponse, "", &result, UnmarshalAlertResolutionsList)
		if err != nil {
			err = core.SDKErrorf(err, "", "unmarshal-resp-error", common.GetComponentInfo())
			return
		}
		response.Result = result
	}

	return
}

// GetManagementAlertsStats : Compute the stats on active alerts
// Compute the stats on active alerts.
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) GetManagementAlertsStats(getManagementAlertsStatsOptions *GetManagementAlertsStatsOptions) (result *McmActiveAlertsStats, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementSreApi.GetManagementAlertsStatsWithContext(context.Background(), getManagementAlertsStatsOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// GetManagementAlertsStatsWithContext is an alternate form of the GetManagementAlertsStats method which supports a Context parameter
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) GetManagementAlertsStatsWithContext(ctx context.Context, getManagementAlertsStatsOptions *GetManagementAlertsStatsOptions) (result *McmActiveAlertsStats, response *core.DetailedResponse, err error) {
	err = core.ValidateNotNil(getManagementAlertsStatsOptions, "getManagementAlertsStatsOptions cannot be nil")
	if err != nil {
		err = core.SDKErrorf(err, "", "unexpected-nil-param", common.GetComponentInfo())
		return
	}
	err = core.ValidateStruct(getManagementAlertsStatsOptions, "getManagementAlertsStatsOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	builder := core.NewRequestBuilder(core.GET)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementSreApi.GetEnableGzipCompression()
	_, err = builder.ResolveRequestURL(backupRecoveryManagementSreApi.Service.Options.URL, `/mcm/stats/alerts`, nil)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_sre_api", "V2", "GetManagementAlertsStats")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range getManagementAlertsStatsOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")

	builder.AddQuery("startTimeUsecs", fmt.Sprint(*getManagementAlertsStatsOptions.StartTimeUsecs))
	builder.AddQuery("endTimeUsecs", fmt.Sprint(*getManagementAlertsStatsOptions.EndTimeUsecs))
	if getManagementAlertsStatsOptions.ClusterIds != nil {
		err = builder.AddQuerySlice("clusterIds", getManagementAlertsStatsOptions.ClusterIds)
		if err != nil {
			err = core.SDKErrorf(err, "", "add-query-slice-error", common.GetComponentInfo())
			return
		}
	}
	if getManagementAlertsStatsOptions.ServiceInstanceIds != nil {
		builder.AddQuery("serviceInstanceIds", strings.Join(getManagementAlertsStatsOptions.ServiceInstanceIds, ","))
	}
	if getManagementAlertsStatsOptions.RegionIds != nil {
		builder.AddQuery("regionIds", strings.Join(getManagementAlertsStatsOptions.RegionIds, ","))
	}
	if getManagementAlertsStatsOptions.ExcludeStatsByCluster != nil {
		builder.AddQuery("excludeStatsByCluster", fmt.Sprint(*getManagementAlertsStatsOptions.ExcludeStatsByCluster))
	}
	if getManagementAlertsStatsOptions.AlertSource != nil {
		builder.AddQuery("alertSource", fmt.Sprint(*getManagementAlertsStatsOptions.AlertSource))
	}
	if getManagementAlertsStatsOptions.TenantIds != nil {
		builder.AddQuery("tenantIds", strings.Join(getManagementAlertsStatsOptions.TenantIds, ","))
	}

	request, err := builder.Build()
	if err != nil {
		err = core.SDKErrorf(err, "", "build-error", common.GetComponentInfo())
		return
	}

	var rawResponse map[string]json.RawMessage
	response, err = backupRecoveryManagementSreApi.Service.Request(request, &rawResponse)
	if err != nil {
		core.EnrichHTTPProblem(err, "GetManagementAlertsStats", getManagementSreServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}
	if rawResponse != nil {
		err = core.UnmarshalModel(rawResponse, "", &result, UnmarshalMcmActiveAlertsStats)
		if err != nil {
			err = core.SDKErrorf(err, "", "unmarshal-resp-error", common.GetComponentInfo())
			return
		}
		response.Result = result
	}

	return
}

// ClustersUpgradesInfo : Fetch upgrade info
// Get progress details and logs for a cluster upgrade. Logs will in json string format.
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) ClustersUpgradesInfo(clustersUpgradesInfoOptions *ClustersUpgradesInfoOptions) (result []UpgradeInfo, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementSreApi.ClustersUpgradesInfoWithContext(context.Background(), clustersUpgradesInfoOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// ClustersUpgradesInfoWithContext is an alternate form of the ClustersUpgradesInfo method which supports a Context parameter
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) ClustersUpgradesInfoWithContext(ctx context.Context, clustersUpgradesInfoOptions *ClustersUpgradesInfoOptions) (result []UpgradeInfo, response *core.DetailedResponse, err error) {
	err = core.ValidateStruct(clustersUpgradesInfoOptions, "clustersUpgradesInfoOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	builder := core.NewRequestBuilder(core.GET)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementSreApi.GetEnableGzipCompression()
	_, err = builder.ResolveRequestURL(backupRecoveryManagementSreApi.Service.Options.URL, `/mcm/cluster-mgmt/upgrades/info`, nil)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_sre_api", "V2", "ClustersUpgradesInfo")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range clustersUpgradesInfoOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")

	if clustersUpgradesInfoOptions.ClusterIdentifiers != nil {
		builder.AddQuery("clusterIdentifiers", strings.Join(clustersUpgradesInfoOptions.ClusterIdentifiers, ","))
	}

	request, err := builder.Build()
	if err != nil {
		err = core.SDKErrorf(err, "", "build-error", common.GetComponentInfo())
		return
	}

	var rawResponse []json.RawMessage
	response, err = backupRecoveryManagementSreApi.Service.Request(request, &rawResponse)
	if err != nil {
		core.EnrichHTTPProblem(err, "ClustersUpgradesInfo", getManagementSreServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}
	if rawResponse != nil {
		err = core.UnmarshalModel(rawResponse, "", &result, UnmarshalUpgradeInfo)
		if err != nil {
			err = core.SDKErrorf(err, "", "unmarshal-resp-error", common.GetComponentInfo())
			return
		}
		response.Result = result
	}

	return
}

// UpdateClustersUpgrades : Updates scheduled cluster upgrades
// Updates scheduled cluster upgrades.
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) UpdateClustersUpgrades(updateClustersUpgradesOptions *UpdateClustersUpgradesOptions) (result []UpgradeResponse, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementSreApi.UpdateClustersUpgradesWithContext(context.Background(), updateClustersUpgradesOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// UpdateClustersUpgradesWithContext is an alternate form of the UpdateClustersUpgrades method which supports a Context parameter
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) UpdateClustersUpgradesWithContext(ctx context.Context, updateClustersUpgradesOptions *UpdateClustersUpgradesOptions) (result []UpgradeResponse, response *core.DetailedResponse, err error) {
	err = core.ValidateStruct(updateClustersUpgradesOptions, "updateClustersUpgradesOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	builder := core.NewRequestBuilder(core.PUT)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementSreApi.GetEnableGzipCompression()
	_, err = builder.ResolveRequestURL(backupRecoveryManagementSreApi.Service.Options.URL, `/mcm/cluster-mgmt/upgrades`, nil)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_sre_api", "V2", "UpdateClustersUpgrades")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range updateClustersUpgradesOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")
	builder.AddHeader("Content-Type", "application/json")

	body := make(map[string]interface{})
	if updateClustersUpgradesOptions.AuthHeaders != nil {
		body["authHeaders"] = updateClustersUpgradesOptions.AuthHeaders
	}
	if updateClustersUpgradesOptions.Clusters != nil {
		body["clusters"] = updateClustersUpgradesOptions.Clusters
	}
	if updateClustersUpgradesOptions.IntervalForRollingUpgradeInHours != nil {
		body["intervalForRollingUpgradeInHours"] = updateClustersUpgradesOptions.IntervalForRollingUpgradeInHours
	}
	if updateClustersUpgradesOptions.PackageURL != nil {
		body["packageUrl"] = updateClustersUpgradesOptions.PackageURL
	}
	if updateClustersUpgradesOptions.PatchUpgradeParams != nil {
		body["patchUpgradeParams"] = updateClustersUpgradesOptions.PatchUpgradeParams
	}
	if updateClustersUpgradesOptions.TargetVersion != nil {
		body["targetVersion"] = updateClustersUpgradesOptions.TargetVersion
	}
	if updateClustersUpgradesOptions.TimeStampToUpgradeAtMsecs != nil {
		body["timeStampToUpgradeAtMsecs"] = updateClustersUpgradesOptions.TimeStampToUpgradeAtMsecs
	}
	if updateClustersUpgradesOptions.Type != nil {
		body["type"] = updateClustersUpgradesOptions.Type
	}
	_, err = builder.SetBodyContentJSON(body)
	if err != nil {
		err = core.SDKErrorf(err, "", "set-json-body-error", common.GetComponentInfo())
		return
	}

	request, err := builder.Build()
	if err != nil {
		err = core.SDKErrorf(err, "", "build-error", common.GetComponentInfo())
		return
	}

	var rawResponse []json.RawMessage
	response, err = backupRecoveryManagementSreApi.Service.Request(request, &rawResponse)
	if err != nil {
		core.EnrichHTTPProblem(err, "UpdateClustersUpgrades", getManagementSreServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}
	if rawResponse != nil {
		err = core.UnmarshalModel(rawResponse, "", &result, UnmarshalUpgradeResponse)
		if err != nil {
			err = core.SDKErrorf(err, "", "unmarshal-resp-error", common.GetComponentInfo())
			return
		}
		response.Result = result
	}

	return
}

// CreateClustersUpgrades : Initiates instant and scheduled cluster upgrade
// Initiates instant and scheduled cluster upgrade.
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) CreateClustersUpgrades(createClustersUpgradesOptions *CreateClustersUpgradesOptions) (result []UpgradeResponse, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementSreApi.CreateClustersUpgradesWithContext(context.Background(), createClustersUpgradesOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// CreateClustersUpgradesWithContext is an alternate form of the CreateClustersUpgrades method which supports a Context parameter
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) CreateClustersUpgradesWithContext(ctx context.Context, createClustersUpgradesOptions *CreateClustersUpgradesOptions) (result []UpgradeResponse, response *core.DetailedResponse, err error) {
	err = core.ValidateStruct(createClustersUpgradesOptions, "createClustersUpgradesOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	builder := core.NewRequestBuilder(core.POST)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementSreApi.GetEnableGzipCompression()
	_, err = builder.ResolveRequestURL(backupRecoveryManagementSreApi.Service.Options.URL, `/mcm/cluster-mgmt/upgrades`, nil)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_sre_api", "V2", "CreateClustersUpgrades")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range createClustersUpgradesOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")
	builder.AddHeader("Content-Type", "application/json")

	body := make(map[string]interface{})
	if createClustersUpgradesOptions.AuthHeaders != nil {
		body["authHeaders"] = createClustersUpgradesOptions.AuthHeaders
	}
	if createClustersUpgradesOptions.Clusters != nil {
		body["clusters"] = createClustersUpgradesOptions.Clusters
	}
	if createClustersUpgradesOptions.IntervalForRollingUpgradeInHours != nil {
		body["intervalForRollingUpgradeInHours"] = createClustersUpgradesOptions.IntervalForRollingUpgradeInHours
	}
	if createClustersUpgradesOptions.PackageURL != nil {
		body["packageUrl"] = createClustersUpgradesOptions.PackageURL
	}
	if createClustersUpgradesOptions.PatchUpgradeParams != nil {
		body["patchUpgradeParams"] = createClustersUpgradesOptions.PatchUpgradeParams
	}
	if createClustersUpgradesOptions.TargetVersion != nil {
		body["targetVersion"] = createClustersUpgradesOptions.TargetVersion
	}
	if createClustersUpgradesOptions.TimeStampToUpgradeAtMsecs != nil {
		body["timeStampToUpgradeAtMsecs"] = createClustersUpgradesOptions.TimeStampToUpgradeAtMsecs
	}
	if createClustersUpgradesOptions.Type != nil {
		body["type"] = createClustersUpgradesOptions.Type
	}
	_, err = builder.SetBodyContentJSON(body)
	if err != nil {
		err = core.SDKErrorf(err, "", "set-json-body-error", common.GetComponentInfo())
		return
	}

	request, err := builder.Build()
	if err != nil {
		err = core.SDKErrorf(err, "", "build-error", common.GetComponentInfo())
		return
	}

	var rawResponse []json.RawMessage
	response, err = backupRecoveryManagementSreApi.Service.Request(request, &rawResponse)
	if err != nil {
		core.EnrichHTTPProblem(err, "CreateClustersUpgrades", getManagementSreServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}
	if rawResponse != nil {
		err = core.UnmarshalModel(rawResponse, "", &result, UnmarshalUpgradeResponse)
		if err != nil {
			err = core.SDKErrorf(err, "", "unmarshal-resp-error", common.GetComponentInfo())
			return
		}
		response.Result = result
	}

	return
}

// DeleteClustersUpgrades : Cancels scheduled cluster upgrades
// Cancels scheduled cluster upgrades.
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) DeleteClustersUpgrades(deleteClustersUpgradesOptions *DeleteClustersUpgradesOptions) (result []UpgradeCancelResponse, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementSreApi.DeleteClustersUpgradesWithContext(context.Background(), deleteClustersUpgradesOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// DeleteClustersUpgradesWithContext is an alternate form of the DeleteClustersUpgrades method which supports a Context parameter
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) DeleteClustersUpgradesWithContext(ctx context.Context, deleteClustersUpgradesOptions *DeleteClustersUpgradesOptions) (result []UpgradeCancelResponse, response *core.DetailedResponse, err error) {
	err = core.ValidateStruct(deleteClustersUpgradesOptions, "deleteClustersUpgradesOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	builder := core.NewRequestBuilder(core.DELETE)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementSreApi.GetEnableGzipCompression()
	_, err = builder.ResolveRequestURL(backupRecoveryManagementSreApi.Service.Options.URL, `/mcm/cluster-mgmt/upgrades`, nil)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_sre_api", "V2", "DeleteClustersUpgrades")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range deleteClustersUpgradesOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")

	if deleteClustersUpgradesOptions.ClusterIdentifiers != nil {
		builder.AddQuery("clusterIdentifiers", strings.Join(deleteClustersUpgradesOptions.ClusterIdentifiers, ","))
	}

	request, err := builder.Build()
	if err != nil {
		err = core.SDKErrorf(err, "", "build-error", common.GetComponentInfo())
		return
	}

	var rawResponse []json.RawMessage
	response, err = backupRecoveryManagementSreApi.Service.Request(request, &rawResponse)
	if err != nil {
		core.EnrichHTTPProblem(err, "DeleteClustersUpgrades", getManagementSreServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}
	if rawResponse != nil {
		err = core.UnmarshalModel(rawResponse, "", &result, UnmarshalUpgradeCancelResponse)
		if err != nil {
			err = core.SDKErrorf(err, "", "unmarshal-resp-error", common.GetComponentInfo())
			return
		}
		response.Result = result
	}

	return
}

// CompatibleClustersForRelease : Fetch compatible clusters for release version
// Get list of clusters that are compatible for an upgrade to the specified release version.
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) CompatibleClustersForRelease(compatibleClustersForReleaseOptions *CompatibleClustersForReleaseOptions) (result []CompatibleCluster, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementSreApi.CompatibleClustersForReleaseWithContext(context.Background(), compatibleClustersForReleaseOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// CompatibleClustersForReleaseWithContext is an alternate form of the CompatibleClustersForRelease method which supports a Context parameter
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) CompatibleClustersForReleaseWithContext(ctx context.Context, compatibleClustersForReleaseOptions *CompatibleClustersForReleaseOptions) (result []CompatibleCluster, response *core.DetailedResponse, err error) {
	err = core.ValidateStruct(compatibleClustersForReleaseOptions, "compatibleClustersForReleaseOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	builder := core.NewRequestBuilder(core.GET)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementSreApi.GetEnableGzipCompression()
	_, err = builder.ResolveRequestURL(backupRecoveryManagementSreApi.Service.Options.URL, `/mcm/cluster-mgmt/compatible-clusters`, nil)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_sre_api", "V2", "CompatibleClustersForRelease")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range compatibleClustersForReleaseOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")

	if compatibleClustersForReleaseOptions.ReleaseVersion != nil {
		builder.AddQuery("releaseVersion", fmt.Sprint(*compatibleClustersForReleaseOptions.ReleaseVersion))
	}

	request, err := builder.Build()
	if err != nil {
		err = core.SDKErrorf(err, "", "build-error", common.GetComponentInfo())
		return
	}

	var rawResponse []json.RawMessage
	response, err = backupRecoveryManagementSreApi.Service.Request(request, &rawResponse)
	if err != nil {
		core.EnrichHTTPProblem(err, "CompatibleClustersForRelease", getManagementSreServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}
	if rawResponse != nil {
		err = core.UnmarshalModel(rawResponse, "", &result, UnmarshalCompatibleCluster)
		if err != nil {
			err = core.SDKErrorf(err, "", "unmarshal-resp-error", common.GetComponentInfo())
			return
		}
		response.Result = result
	}

	return
}

// GetClustersInfo : Clusters information with upgrade details
// Get clusters information and their upgrade details.
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) GetClustersInfo(getClustersInfoOptions *GetClustersInfoOptions) (result *ClusterDetails, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementSreApi.GetClustersInfoWithContext(context.Background(), getClustersInfoOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// GetClustersInfoWithContext is an alternate form of the GetClustersInfo method which supports a Context parameter
func (backupRecoveryManagementSreApi *BackupRecoveryManagementSreApiV1) GetClustersInfoWithContext(ctx context.Context, getClustersInfoOptions *GetClustersInfoOptions) (result *ClusterDetails, response *core.DetailedResponse, err error) {
	err = core.ValidateStruct(getClustersInfoOptions, "getClustersInfoOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	builder := core.NewRequestBuilder(core.GET)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementSreApi.GetEnableGzipCompression()
	_, err = builder.ResolveRequestURL(backupRecoveryManagementSreApi.Service.Options.URL, `/mcm/cluster-mgmt/info`, nil)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_sre_api", "V2", "GetClustersInfo")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range getClustersInfoOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")

	request, err := builder.Build()
	if err != nil {
		err = core.SDKErrorf(err, "", "build-error", common.GetComponentInfo())
		return
	}

	var rawResponse map[string]json.RawMessage
	response, err = backupRecoveryManagementSreApi.Service.Request(request, &rawResponse)
	if err != nil {
		core.EnrichHTTPProblem(err, "GetClustersInfo", getManagementSreServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}
	if rawResponse != nil {
		err = core.UnmarshalModel(rawResponse, "", &result, UnmarshalClusterDetails)
		if err != nil {
			err = core.SDKErrorf(err, "", "unmarshal-resp-error", common.GetComponentInfo())
			return
		}
		response.Result = result
	}

	return
}
func getManagementSreServiceComponentInfo() *core.ProblemComponent {
	return core.NewProblemComponent(DefaultServiceName, "2.0")
}

// ActiveAlertsStats : Specifies the active alert statistics details.
type ActiveAlertsStats struct {
	// Specifies the count of active critical Alerts excluding alerts that belong to other bucket.
	NumCriticalAlerts *int64 `json:"numCriticalAlerts,omitempty"`

	// Specifies the count of active critical alerts categories.
	NumCriticalAlertsCategories *int64 `json:"numCriticalAlertsCategories,omitempty"`

	// Specifies the count of active service Alerts.
	NumDataServiceAlerts *int64 `json:"numDataServiceAlerts,omitempty"`

	// Specifies the count of active service critical Alerts.
	NumDataServiceCriticalAlerts *int64 `json:"numDataServiceCriticalAlerts,omitempty"`

	// Specifies the count of active service info Alerts.
	NumDataServiceInfoAlerts *int64 `json:"numDataServiceInfoAlerts,omitempty"`

	// Specifies the count of active service warning Alerts.
	NumDataServiceWarningAlerts *int64 `json:"numDataServiceWarningAlerts,omitempty"`

	// Specifies the count of active hardware Alerts.
	NumHardwareAlerts *int64 `json:"numHardwareAlerts,omitempty"`

	// Specifies the count of active hardware critical Alerts.
	NumHardwareCriticalAlerts *int64 `json:"numHardwareCriticalAlerts,omitempty"`

	// Specifies the count of active hardware info Alerts.
	NumHardwareInfoAlerts *int64 `json:"numHardwareInfoAlerts,omitempty"`

	// Specifies the count of active hardware warning Alerts.
	NumHardwareWarningAlerts *int64 `json:"numHardwareWarningAlerts,omitempty"`

	// Specifies the count of active info Alerts excluding alerts that belong to other bucket.
	NumInfoAlerts *int64 `json:"numInfoAlerts,omitempty"`

	// Specifies the count of active info alerts categories.
	NumInfoAlertsCategories *int64 `json:"numInfoAlertsCategories,omitempty"`

	// Specifies the count of active Alerts of maintenance bucket.
	NumMaintenanceAlerts *int64 `json:"numMaintenanceAlerts,omitempty"`

	// Specifies the count of active other critical Alerts.
	NumMaintenanceCriticalAlerts *int64 `json:"numMaintenanceCriticalAlerts,omitempty"`

	// Specifies the count of active other info Alerts.
	NumMaintenanceInfoAlerts *int64 `json:"numMaintenanceInfoAlerts,omitempty"`

	// Specifies the count of active other warning Alerts.
	NumMaintenanceWarningAlerts *int64 `json:"numMaintenanceWarningAlerts,omitempty"`

	// Specifies the count of active software Alerts.
	NumSoftwareAlerts *int64 `json:"numSoftwareAlerts,omitempty"`

	// Specifies the count of active software critical Alerts.
	NumSoftwareCriticalAlerts *int64 `json:"numSoftwareCriticalAlerts,omitempty"`

	// Specifies the count of active software info Alerts.
	NumSoftwareInfoAlerts *int64 `json:"numSoftwareInfoAlerts,omitempty"`

	// Specifies the count of active software warning Alerts.
	NumSoftwareWarningAlerts *int64 `json:"numSoftwareWarningAlerts,omitempty"`

	// Specifies the count of active warning Alerts excluding alerts that belong to other bucket.
	NumWarningAlerts *int64 `json:"numWarningAlerts,omitempty"`

	// Specifies the count of active warning alerts categories.
	NumWarningAlertsCategories *int64 `json:"numWarningAlertsCategories,omitempty"`
}

// UnmarshalActiveAlertsStats unmarshals an instance of ActiveAlertsStats from the specified map of raw messages.
func UnmarshalActiveAlertsStats(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(ActiveAlertsStats)
	err = core.UnmarshalPrimitive(m, "numCriticalAlerts", &obj.NumCriticalAlerts)
	if err != nil {
		err = core.SDKErrorf(err, "", "numCriticalAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numCriticalAlertsCategories", &obj.NumCriticalAlertsCategories)
	if err != nil {
		err = core.SDKErrorf(err, "", "numCriticalAlertsCategories-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numDataServiceAlerts", &obj.NumDataServiceAlerts)
	if err != nil {
		err = core.SDKErrorf(err, "", "numDataServiceAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numDataServiceCriticalAlerts", &obj.NumDataServiceCriticalAlerts)
	if err != nil {
		err = core.SDKErrorf(err, "", "numDataServiceCriticalAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numDataServiceInfoAlerts", &obj.NumDataServiceInfoAlerts)
	if err != nil {
		err = core.SDKErrorf(err, "", "numDataServiceInfoAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numDataServiceWarningAlerts", &obj.NumDataServiceWarningAlerts)
	if err != nil {
		err = core.SDKErrorf(err, "", "numDataServiceWarningAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numHardwareAlerts", &obj.NumHardwareAlerts)
	if err != nil {
		err = core.SDKErrorf(err, "", "numHardwareAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numHardwareCriticalAlerts", &obj.NumHardwareCriticalAlerts)
	if err != nil {
		err = core.SDKErrorf(err, "", "numHardwareCriticalAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numHardwareInfoAlerts", &obj.NumHardwareInfoAlerts)
	if err != nil {
		err = core.SDKErrorf(err, "", "numHardwareInfoAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numHardwareWarningAlerts", &obj.NumHardwareWarningAlerts)
	if err != nil {
		err = core.SDKErrorf(err, "", "numHardwareWarningAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numInfoAlerts", &obj.NumInfoAlerts)
	if err != nil {
		err = core.SDKErrorf(err, "", "numInfoAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numInfoAlertsCategories", &obj.NumInfoAlertsCategories)
	if err != nil {
		err = core.SDKErrorf(err, "", "numInfoAlertsCategories-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numMaintenanceAlerts", &obj.NumMaintenanceAlerts)
	if err != nil {
		err = core.SDKErrorf(err, "", "numMaintenanceAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numMaintenanceCriticalAlerts", &obj.NumMaintenanceCriticalAlerts)
	if err != nil {
		err = core.SDKErrorf(err, "", "numMaintenanceCriticalAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numMaintenanceInfoAlerts", &obj.NumMaintenanceInfoAlerts)
	if err != nil {
		err = core.SDKErrorf(err, "", "numMaintenanceInfoAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numMaintenanceWarningAlerts", &obj.NumMaintenanceWarningAlerts)
	if err != nil {
		err = core.SDKErrorf(err, "", "numMaintenanceWarningAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numSoftwareAlerts", &obj.NumSoftwareAlerts)
	if err != nil {
		err = core.SDKErrorf(err, "", "numSoftwareAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numSoftwareCriticalAlerts", &obj.NumSoftwareCriticalAlerts)
	if err != nil {
		err = core.SDKErrorf(err, "", "numSoftwareCriticalAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numSoftwareInfoAlerts", &obj.NumSoftwareInfoAlerts)
	if err != nil {
		err = core.SDKErrorf(err, "", "numSoftwareInfoAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numSoftwareWarningAlerts", &obj.NumSoftwareWarningAlerts)
	if err != nil {
		err = core.SDKErrorf(err, "", "numSoftwareWarningAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numWarningAlerts", &obj.NumWarningAlerts)
	if err != nil {
		err = core.SDKErrorf(err, "", "numWarningAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numWarningAlertsCategories", &obj.NumWarningAlertsCategories)
	if err != nil {
		err = core.SDKErrorf(err, "", "numWarningAlertsCategories-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// Alert : Specifies the fields of an alert.
type Alert struct {
	// Specifies the alert category.
	AlertCategory *string `json:"alertCategory,omitempty"`

	// Specifies a unique code that categorizes the Alert, for example: CE00200014, where CE stands for IBM Error, the
	// alert state next 3 digits is the id of the Alert Category (e.g. 002 for 'kNode') and the last 5 digits is the id of
	// the Alert Type (e.g. 00014 for 'kNodeHighCpuUsage').
	AlertCode *string `json:"alertCode,omitempty"`

	// Specifies the fields of alert document.
	AlertDocument *AlertDocument `json:"alertDocument,omitempty"`

	// Specifies the alert state.
	AlertState *string `json:"alertState,omitempty"`

	// Specifies the alert type.
	AlertType *int64 `json:"alertType,omitempty"`

	// Specifies the Alert type bucket.
	AlertTypeBucket *string `json:"alertTypeBucket,omitempty"`

	// Id of the cluster which the alert is associated.
	ClusterID *int64 `json:"clusterId,omitempty"`

	// Specifies the name of cluster which alert is raised from.
	ClusterName *string `json:"clusterName,omitempty"`

	// Specifies the dedup count of alert.
	DedupCount *int64 `json:"dedupCount,omitempty"`

	// Specifies Unix epoch Timestamps (in microseconds) for the last 25 occurrences of duplicated Alerts that are stored
	// with the original/primary Alert. Alerts are grouped into one Alert if the Alerts are the same type, are reporting on
	// the same Object and occur within one hour. 'dedupCount' always reports the total count of duplicated Alerts even if
	// there are more than 25 occurrences. For example, if there are 100 occurrences of this Alert, dedupTimestamps stores
	// the timestamps of the last 25 occurrences and dedupCount equals 100.
	DedupTimestamps []int64 `json:"dedupTimestamps,omitempty"`

	// SpeSpecifies Unix epoch Timestamp (in microseconds) of the first occurrence of the Alert.
	FirstTimestampUsecs *int64 `json:"firstTimestampUsecs,omitempty"`

	// Specifies unique id of the alert.
	ID *string `json:"id,omitempty"`

	// SpeSpecifies Unix epoch Timestamp (in microseconds) of the most recent occurrence of the Alert.
	LatestTimestampUsecs *int64 `json:"latestTimestampUsecs,omitempty"`

	// List of property key and values associated with alert.
	PropertyList []Label `json:"propertyList,omitempty"`

	// Specifies the region id of the alert.
	RegionID *string `json:"regionId,omitempty"`

	// Specifies the resolution id of the alert if its resolved.
	ResolutionIdString *string `json:"resolutionIdString,omitempty"`

	// Id of the serrvice instance which the alert is associated.
	ServiceInstanceID *string `json:"serviceInstanceId,omitempty"`

	// Specifies the alert severity.
	Severity *string `json:"severity,omitempty"`

	// Specifies information about vaults where source object associated with alert is vaulted. This could be empty if
	// alert is not related to any source object or it is not vaulted.
	Vaults []Vault `json:"vaults,omitempty"`
}

// Constants associated with the Alert.AlertCategory property.
// Specifies the alert category.
const (
	Alert_AlertCategory_Kagent                  = "kAgent"
	Alert_AlertCategory_Kappmarketplace         = "kAppMarketPlace"
	Alert_AlertCategory_Karchivalrestore        = "kArchivalRestore"
	Alert_AlertCategory_Kauditlog               = "kAuditLog"
	Alert_AlertCategory_Kbackuprestore          = "kBackupRestore"
	Alert_AlertCategory_Kcdp                    = "kCDP"
	Alert_AlertCategory_Kchassis                = "kChassis"
	Alert_AlertCategory_Kcluster                = "kCluster"
	Alert_AlertCategory_Kclustermanagement      = "kClusterManagement"
	Alert_AlertCategory_Kconfiguration          = "kConfiguration"
	Alert_AlertCategory_Kcpu                    = "kCPU"
	Alert_AlertCategory_Kdatapath               = "kDataPath"
	Alert_AlertCategory_Kdisasterrecovery       = "kDisasterRecovery"
	Alert_AlertCategory_Kdisk                   = "kDisk"
	Alert_AlertCategory_Kfan                    = "kFan"
	Alert_AlertCategory_Kfaulttolerance         = "kFaultTolerance"
	Alert_AlertCategory_Kfirmware               = "kFirmware"
	Alert_AlertCategory_Kgeneralsoftwarefailure = "kGeneralSoftwareFailure"
	Alert_AlertCategory_Khelios                 = "kHelios"
	Alert_AlertCategory_Kindexing               = "kIndexing"
	Alert_AlertCategory_Klicense                = "kLicense"
	Alert_AlertCategory_Kmemory                 = "kMemory"
	Alert_AlertCategory_Kmetadata               = "kMetadata"
	Alert_AlertCategory_Knetworking             = "kNetworking"
	Alert_AlertCategory_Knic                    = "kNIC"
	Alert_AlertCategory_Knode                   = "kNode"
	Alert_AlertCategory_Knodehealth             = "kNodeHealth"
	Alert_AlertCategory_Koperatingsystem        = "kOperatingSystem"
	Alert_AlertCategory_Kpowersupply            = "kPowerSupply"
	Alert_AlertCategory_Kquota                  = "kQuota"
	Alert_AlertCategory_Kremotereplication      = "kRemoteReplication"
	Alert_AlertCategory_Ksecurity               = "kSecurity"
	Alert_AlertCategory_Kstoragedevice          = "kStorageDevice"
	Alert_AlertCategory_Kstoragepool            = "kStoragePool"
	Alert_AlertCategory_Kstorageusage           = "kStorageUsage"
	Alert_AlertCategory_Ksystemservice          = "kSystemService"
	Alert_AlertCategory_Ktemperature            = "kTemperature"
	Alert_AlertCategory_Kupgrade                = "kUpgrade"
	Alert_AlertCategory_Kviewfailover           = "kViewFailover"
)

// Constants associated with the Alert.AlertState property.
// Specifies the alert state.
const (
	Alert_AlertState_Knote       = "kNote"
	Alert_AlertState_Kopen       = "kOpen"
	Alert_AlertState_Kresolved   = "kResolved"
	Alert_AlertState_Ksuppressed = "kSuppressed"
)

// Constants associated with the Alert.AlertTypeBucket property.
// Specifies the Alert type bucket.
const (
	Alert_AlertTypeBucket_Kdataservice = "kDataService"
	Alert_AlertTypeBucket_Khardware    = "kHardware"
	Alert_AlertTypeBucket_Kmaintenance = "kMaintenance"
	Alert_AlertTypeBucket_Ksoftware    = "kSoftware"
)

// Constants associated with the Alert.Severity property.
// Specifies the alert severity.
const (
	Alert_Severity_Kcritical = "kCritical"
	Alert_Severity_Kinfo     = "kInfo"
	Alert_Severity_Kwarning  = "kWarning"
)

// UnmarshalAlert unmarshals an instance of Alert from the specified map of raw messages.
func UnmarshalAlert(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(Alert)
	err = core.UnmarshalPrimitive(m, "alertCategory", &obj.AlertCategory)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertCategory-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "alertCode", &obj.AlertCode)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertCode-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "alertDocument", &obj.AlertDocument, UnmarshalAlertDocument)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertDocument-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "alertState", &obj.AlertState)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertState-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "alertType", &obj.AlertType)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertType-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "alertTypeBucket", &obj.AlertTypeBucket)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertTypeBucket-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "clusterId", &obj.ClusterID)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "clusterName", &obj.ClusterName)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterName-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "dedupCount", &obj.DedupCount)
	if err != nil {
		err = core.SDKErrorf(err, "", "dedupCount-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "dedupTimestamps", &obj.DedupTimestamps)
	if err != nil {
		err = core.SDKErrorf(err, "", "dedupTimestamps-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "firstTimestampUsecs", &obj.FirstTimestampUsecs)
	if err != nil {
		err = core.SDKErrorf(err, "", "firstTimestampUsecs-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "id", &obj.ID)
	if err != nil {
		err = core.SDKErrorf(err, "", "id-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "latestTimestampUsecs", &obj.LatestTimestampUsecs)
	if err != nil {
		err = core.SDKErrorf(err, "", "latestTimestampUsecs-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "propertyList", &obj.PropertyList, UnmarshalLabel)
	if err != nil {
		err = core.SDKErrorf(err, "", "propertyList-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "regionId", &obj.RegionID)
	if err != nil {
		err = core.SDKErrorf(err, "", "regionId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "resolutionIdString", &obj.ResolutionIdString)
	if err != nil {
		err = core.SDKErrorf(err, "", "resolutionIdString-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "serviceInstanceId", &obj.ServiceInstanceID)
	if err != nil {
		err = core.SDKErrorf(err, "", "serviceInstanceId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "severity", &obj.Severity)
	if err != nil {
		err = core.SDKErrorf(err, "", "severity-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "vaults", &obj.Vaults, UnmarshalVault)
	if err != nil {
		err = core.SDKErrorf(err, "", "vaults-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// AlertDocument : Specifies the fields of alert document.
type AlertDocument struct {
	// Specifies the cause of alert.
	AlertCause *string `json:"alertCause,omitempty"`

	// Specifies the description of alert.
	AlertDescription *string `json:"alertDescription,omitempty"`

	// Specifies the help text for alert.
	AlertHelpText *string `json:"alertHelpText,omitempty"`

	// Specifies the name of alert.
	AlertName *string `json:"alertName,omitempty"`

	// Short description for the alert.
	AlertSummary *string `json:"alertSummary,omitempty"`
}

// UnmarshalAlertDocument unmarshals an instance of AlertDocument from the specified map of raw messages.
func UnmarshalAlertDocument(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(AlertDocument)
	err = core.UnmarshalPrimitive(m, "alertCause", &obj.AlertCause)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertCause-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "alertDescription", &obj.AlertDescription)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertDescription-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "alertHelpText", &obj.AlertHelpText)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertHelpText-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "alertName", &obj.AlertName)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertName-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "alertSummary", &obj.AlertSummary)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertSummary-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// AlertGroupSummary : Specifies alerts summary grouped for an alert category.
type AlertGroupSummary struct {
	// Category of alerts by which summary is grouped.
	Category *string `json:"category,omitempty"`

	// Specifies count of critical alerts.
	CriticalCount *int64 `json:"criticalCount,omitempty"`

	// Specifies count of info alerts.
	InfoCount *int64 `json:"infoCount,omitempty"`

	// Specifies count of total alerts.
	TotalCount *int64 `json:"totalCount,omitempty"`

	// Type/bucket that this alert category belongs to.
	Type *string `json:"type,omitempty"`

	// Specifies count of warning alerts.
	WarningCount *int64 `json:"warningCount,omitempty"`
}

// Constants associated with the AlertGroupSummary.Category property.
// Category of alerts by which summary is grouped.
const (
	AlertGroupSummary_Category_Kagent                  = "kAgent"
	AlertGroupSummary_Category_Kappmarketplace         = "kAppMarketPlace"
	AlertGroupSummary_Category_Karchivalrestore        = "kArchivalRestore"
	AlertGroupSummary_Category_Kauditlog               = "kAuditLog"
	AlertGroupSummary_Category_Kbackuprestore          = "kBackupRestore"
	AlertGroupSummary_Category_Kcdp                    = "kCDP"
	AlertGroupSummary_Category_Kchassis                = "kChassis"
	AlertGroupSummary_Category_Kcluster                = "kCluster"
	AlertGroupSummary_Category_Kclustermanagement      = "kClusterManagement"
	AlertGroupSummary_Category_Kconfiguration          = "kConfiguration"
	AlertGroupSummary_Category_Kcpu                    = "kCPU"
	AlertGroupSummary_Category_Kdatapath               = "kDataPath"
	AlertGroupSummary_Category_Kdisasterrecovery       = "kDisasterRecovery"
	AlertGroupSummary_Category_Kdisk                   = "kDisk"
	AlertGroupSummary_Category_Kfan                    = "kFan"
	AlertGroupSummary_Category_Kfaulttolerance         = "kFaultTolerance"
	AlertGroupSummary_Category_Kfirmware               = "kFirmware"
	AlertGroupSummary_Category_Kgeneralsoftwarefailure = "kGeneralSoftwareFailure"
	AlertGroupSummary_Category_Khelios                 = "kHelios"
	AlertGroupSummary_Category_Kindexing               = "kIndexing"
	AlertGroupSummary_Category_Klicense                = "kLicense"
	AlertGroupSummary_Category_Kmemory                 = "kMemory"
	AlertGroupSummary_Category_Kmetadata               = "kMetadata"
	AlertGroupSummary_Category_Knetworking             = "kNetworking"
	AlertGroupSummary_Category_Knic                    = "kNIC"
	AlertGroupSummary_Category_Knode                   = "kNode"
	AlertGroupSummary_Category_Knodehealth             = "kNodeHealth"
	AlertGroupSummary_Category_Koperatingsystem        = "kOperatingSystem"
	AlertGroupSummary_Category_Kpowersupply            = "kPowerSupply"
	AlertGroupSummary_Category_Kquota                  = "kQuota"
	AlertGroupSummary_Category_Kremotereplication      = "kRemoteReplication"
	AlertGroupSummary_Category_Ksecurity               = "kSecurity"
	AlertGroupSummary_Category_Kstoragedevice          = "kStorageDevice"
	AlertGroupSummary_Category_Kstoragepool            = "kStoragePool"
	AlertGroupSummary_Category_Kstorageusage           = "kStorageUsage"
	AlertGroupSummary_Category_Ksystemservice          = "kSystemService"
	AlertGroupSummary_Category_Ktemperature            = "kTemperature"
	AlertGroupSummary_Category_Kupgrade                = "kUpgrade"
	AlertGroupSummary_Category_Kviewfailover           = "kViewFailover"
)

// UnmarshalAlertGroupSummary unmarshals an instance of AlertGroupSummary from the specified map of raw messages.
func UnmarshalAlertGroupSummary(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(AlertGroupSummary)
	err = core.UnmarshalPrimitive(m, "category", &obj.Category)
	if err != nil {
		err = core.SDKErrorf(err, "", "category-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "criticalCount", &obj.CriticalCount)
	if err != nil {
		err = core.SDKErrorf(err, "", "criticalCount-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "infoCount", &obj.InfoCount)
	if err != nil {
		err = core.SDKErrorf(err, "", "infoCount-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "totalCount", &obj.TotalCount)
	if err != nil {
		err = core.SDKErrorf(err, "", "totalCount-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "type", &obj.Type)
	if err != nil {
		err = core.SDKErrorf(err, "", "type-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "warningCount", &obj.WarningCount)
	if err != nil {
		err = core.SDKErrorf(err, "", "warningCount-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// AlertInfo : Specifies the fields of an alert.
type AlertInfo struct {
	// Specifies the alert category.
	AlertCategory *string `json:"alertCategory,omitempty"`

	// Specifies a unique code that categorizes the Alert, for example: CE00200014, where CE stands for IBM Error, the
	// alert state next 3 digits is the id of the Alert Category (e.g. 002 for 'kNode') and the last 5 digits is the id of
	// the Alert Type (e.g. 00014 for 'kNodeHighCpuUsage').
	AlertCode *string `json:"alertCode,omitempty"`

	// Specifies the fields of alert document.
	AlertDocument *AlertDocument `json:"alertDocument,omitempty"`

	// Specifies the alert state.
	AlertState *string `json:"alertState,omitempty"`

	// Specifies the alert type.
	AlertType *int64 `json:"alertType,omitempty"`

	// Specifies the Alert type bucket.
	AlertTypeBucket *string `json:"alertTypeBucket,omitempty"`

	// Id of the cluster which the alert is associated.
	ClusterID *int64 `json:"clusterId,omitempty"`

	// Specifies the name of cluster which alert is raised from.
	ClusterName *string `json:"clusterName,omitempty"`

	// Specifies the dedup count of alert.
	DedupCount *int64 `json:"dedupCount,omitempty"`

	// Specifies Unix epoch Timestamps (in microseconds) for the last 25 occurrences of duplicated Alerts that are stored
	// with the original/primary Alert. Alerts are grouped into one Alert if the Alerts are the same type, are reporting on
	// the same Object and occur within one hour. 'dedupCount' always reports the total count of duplicated Alerts even if
	// there are more than 25 occurrences. For example, if there are 100 occurrences of this Alert, dedupTimestamps stores
	// the timestamps of the last 25 occurrences and dedupCount equals 100.
	DedupTimestamps []int64 `json:"dedupTimestamps,omitempty"`

	// Specifies source where the event occurred.
	EventSource *string `json:"eventSource,omitempty"`

	// Specifies Unix epoch Timestamp (in microseconds) of the first occurrence of the Alert.
	FirstTimestampUsecs *int64 `json:"firstTimestampUsecs,omitempty"`

	// Specifies unique id of the alert.
	ID *string `json:"id,omitempty"`

	// Specifies the labels for which this alert has been raised.
	LabelIds []string `json:"labelIds,omitempty"`

	// Specifies Unix epoch Timestamp (in microseconds) of the most recent occurrence of the Alert.
	LatestTimestampUsecs *int64 `json:"latestTimestampUsecs,omitempty"`

	// List of property key and values associated with alert.
	PropertyList []Label `json:"propertyList,omitempty"`

	// Specifies the region id of the alert.
	RegionID *string `json:"regionId,omitempty"`

	// Specifies information about the Alert Resolution.
	ResolutionDetails *AlertResolutionDetails `json:"resolutionDetails,omitempty"`

	// Resolution Id String.
	ResolutionIdString *string `json:"resolutionIdString,omitempty"`

	// Specifies Unix epoch Timestamps in microseconds when alert is resolved.
	ResolvedTimestampUsecs *int64 `json:"resolvedTimestampUsecs,omitempty"`

	// Specifies the alert severity.
	Severity *string `json:"severity,omitempty"`

	// Specifies unique id generated when the Alert is suppressed by the admin.
	SuppressionID *int64 `json:"suppressionId,omitempty"`

	// Specifies the tenants for which this alert has been raised.
	TenantIds []string `json:"tenantIds,omitempty"`

	// Specifies information about vaults where source object associated with alert is vaulted. This could be empty if
	// alert is not related to any source object or it is not vaulted.
	Vaults []Vault `json:"vaults,omitempty"`
}

// Constants associated with the AlertInfo.AlertCategory property.
// Specifies the alert category.
const (
	AlertInfo_AlertCategory_Kagent                  = "kAgent"
	AlertInfo_AlertCategory_Kappmarketplace         = "kAppMarketPlace"
	AlertInfo_AlertCategory_Karchivalrestore        = "kArchivalRestore"
	AlertInfo_AlertCategory_Kauditlog               = "kAuditLog"
	AlertInfo_AlertCategory_Kbackuprestore          = "kBackupRestore"
	AlertInfo_AlertCategory_Kcdp                    = "kCDP"
	AlertInfo_AlertCategory_Kchassis                = "kChassis"
	AlertInfo_AlertCategory_Kcluster                = "kCluster"
	AlertInfo_AlertCategory_Kclustermanagement      = "kClusterManagement"
	AlertInfo_AlertCategory_Kconfiguration          = "kConfiguration"
	AlertInfo_AlertCategory_Kcpu                    = "kCPU"
	AlertInfo_AlertCategory_Kdatapath               = "kDataPath"
	AlertInfo_AlertCategory_Kdisasterrecovery       = "kDisasterRecovery"
	AlertInfo_AlertCategory_Kdisk                   = "kDisk"
	AlertInfo_AlertCategory_Kfan                    = "kFan"
	AlertInfo_AlertCategory_Kfaulttolerance         = "kFaultTolerance"
	AlertInfo_AlertCategory_Kfirmware               = "kFirmware"
	AlertInfo_AlertCategory_Kgeneralsoftwarefailure = "kGeneralSoftwareFailure"
	AlertInfo_AlertCategory_Khelios                 = "kHelios"
	AlertInfo_AlertCategory_Kindexing               = "kIndexing"
	AlertInfo_AlertCategory_Klicense                = "kLicense"
	AlertInfo_AlertCategory_Kmemory                 = "kMemory"
	AlertInfo_AlertCategory_Kmetadata               = "kMetadata"
	AlertInfo_AlertCategory_Knetworking             = "kNetworking"
	AlertInfo_AlertCategory_Knic                    = "kNIC"
	AlertInfo_AlertCategory_Knode                   = "kNode"
	AlertInfo_AlertCategory_Knodehealth             = "kNodeHealth"
	AlertInfo_AlertCategory_Koperatingsystem        = "kOperatingSystem"
	AlertInfo_AlertCategory_Kpowersupply            = "kPowerSupply"
	AlertInfo_AlertCategory_Kquota                  = "kQuota"
	AlertInfo_AlertCategory_Kremotereplication      = "kRemoteReplication"
	AlertInfo_AlertCategory_Ksecurity               = "kSecurity"
	AlertInfo_AlertCategory_Kstoragedevice          = "kStorageDevice"
	AlertInfo_AlertCategory_Kstoragepool            = "kStoragePool"
	AlertInfo_AlertCategory_Kstorageusage           = "kStorageUsage"
	AlertInfo_AlertCategory_Ksystemservice          = "kSystemService"
	AlertInfo_AlertCategory_Ktemperature            = "kTemperature"
	AlertInfo_AlertCategory_Kupgrade                = "kUpgrade"
	AlertInfo_AlertCategory_Kviewfailover           = "kViewFailover"
)

// Constants associated with the AlertInfo.AlertState property.
// Specifies the alert state.
const (
	AlertInfo_AlertState_Knote       = "kNote"
	AlertInfo_AlertState_Kopen       = "kOpen"
	AlertInfo_AlertState_Kresolved   = "kResolved"
	AlertInfo_AlertState_Ksuppressed = "kSuppressed"
)

// Constants associated with the AlertInfo.AlertTypeBucket property.
// Specifies the Alert type bucket.
const (
	AlertInfo_AlertTypeBucket_Kdataservice = "kDataService"
	AlertInfo_AlertTypeBucket_Khardware    = "kHardware"
	AlertInfo_AlertTypeBucket_Kmaintenance = "kMaintenance"
	AlertInfo_AlertTypeBucket_Ksoftware    = "kSoftware"
)

// Constants associated with the AlertInfo.Severity property.
// Specifies the alert severity.
const (
	AlertInfo_Severity_Kcritical = "kCritical"
	AlertInfo_Severity_Kinfo     = "kInfo"
	AlertInfo_Severity_Kwarning  = "kWarning"
)

// UnmarshalAlertInfo unmarshals an instance of AlertInfo from the specified map of raw messages.
func UnmarshalAlertInfo(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(AlertInfo)
	err = core.UnmarshalPrimitive(m, "alertCategory", &obj.AlertCategory)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertCategory-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "alertCode", &obj.AlertCode)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertCode-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "alertDocument", &obj.AlertDocument, UnmarshalAlertDocument)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertDocument-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "alertState", &obj.AlertState)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertState-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "alertType", &obj.AlertType)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertType-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "alertTypeBucket", &obj.AlertTypeBucket)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertTypeBucket-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "clusterId", &obj.ClusterID)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "clusterName", &obj.ClusterName)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterName-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "dedupCount", &obj.DedupCount)
	if err != nil {
		err = core.SDKErrorf(err, "", "dedupCount-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "dedupTimestamps", &obj.DedupTimestamps)
	if err != nil {
		err = core.SDKErrorf(err, "", "dedupTimestamps-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "eventSource", &obj.EventSource)
	if err != nil {
		err = core.SDKErrorf(err, "", "eventSource-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "firstTimestampUsecs", &obj.FirstTimestampUsecs)
	if err != nil {
		err = core.SDKErrorf(err, "", "firstTimestampUsecs-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "id", &obj.ID)
	if err != nil {
		err = core.SDKErrorf(err, "", "id-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "labelIds", &obj.LabelIds)
	if err != nil {
		err = core.SDKErrorf(err, "", "labelIds-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "latestTimestampUsecs", &obj.LatestTimestampUsecs)
	if err != nil {
		err = core.SDKErrorf(err, "", "latestTimestampUsecs-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "propertyList", &obj.PropertyList, UnmarshalLabel)
	if err != nil {
		err = core.SDKErrorf(err, "", "propertyList-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "regionId", &obj.RegionID)
	if err != nil {
		err = core.SDKErrorf(err, "", "regionId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "resolutionDetails", &obj.ResolutionDetails, UnmarshalAlertResolutionDetails)
	if err != nil {
		err = core.SDKErrorf(err, "", "resolutionDetails-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "resolutionIdString", &obj.ResolutionIdString)
	if err != nil {
		err = core.SDKErrorf(err, "", "resolutionIdString-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "resolvedTimestampUsecs", &obj.ResolvedTimestampUsecs)
	if err != nil {
		err = core.SDKErrorf(err, "", "resolvedTimestampUsecs-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "severity", &obj.Severity)
	if err != nil {
		err = core.SDKErrorf(err, "", "severity-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "suppressionId", &obj.SuppressionID)
	if err != nil {
		err = core.SDKErrorf(err, "", "suppressionId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "tenantIds", &obj.TenantIds)
	if err != nil {
		err = core.SDKErrorf(err, "", "tenantIds-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "vaults", &obj.Vaults, UnmarshalVault)
	if err != nil {
		err = core.SDKErrorf(err, "", "vaults-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// AlertList : Specifies the response of get alerts.
type AlertList struct {
	// Specifies the list of alerts.
	Alerts []AlertInfo `json:"alerts" validate:"required"`
}

// UnmarshalAlertList unmarshals an instance of AlertList from the specified map of raw messages.
func UnmarshalAlertList(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(AlertList)
	err = core.UnmarshalModel(m, "alerts", &obj.Alerts, UnmarshalAlertInfo)
	if err != nil {
		err = core.SDKErrorf(err, "", "alerts-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// AlertResolution : Provides Resolution details and the list of Alerts resolved.
type AlertResolution struct {
	// Specifies account id of the user who create the resolution.
	AccountID *string `json:"accountId,omitempty"`

	// Specifies unix epoch timestamp (in microseconds) when the resolution is created.
	CreatedTimeUsecs *int64 `json:"createdTimeUsecs,omitempty"`

	// Specifies the full description about the Resolution.
	Description *string `json:"description,omitempty"`

	// Specifies the external key assigned outside of management console, with the form of "clusterid:resolutionid".
	ExternalKey *string `json:"externalKey,omitempty"`

	// Specifies the unique reslution id assigned in management console.
	ResolutionID *string `json:"resolutionId,omitempty"`

	// Specifies the unique name of the resolution.
	ResolutionName *string `json:"resolutionName,omitempty"`

	ResolvedAlerts []ResolvedAlertInfo `json:"resolvedAlerts,omitempty"`

	// Specifies the time duration (in minutes) for silencing alerts. If unspecified or set zero, a silence rule will be
	// created with default expiry time. No silence rule will be created if value < 0.
	SilenceMinutes *int64 `json:"silenceMinutes,omitempty"`

	// Specifies tenant id of the user who create the resolution.
	TenantID *string `json:"tenantId,omitempty"`
}

// UnmarshalAlertResolution unmarshals an instance of AlertResolution from the specified map of raw messages.
func UnmarshalAlertResolution(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(AlertResolution)
	err = core.UnmarshalPrimitive(m, "accountId", &obj.AccountID)
	if err != nil {
		err = core.SDKErrorf(err, "", "accountId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "createdTimeUsecs", &obj.CreatedTimeUsecs)
	if err != nil {
		err = core.SDKErrorf(err, "", "createdTimeUsecs-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "description", &obj.Description)
	if err != nil {
		err = core.SDKErrorf(err, "", "description-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "externalKey", &obj.ExternalKey)
	if err != nil {
		err = core.SDKErrorf(err, "", "externalKey-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "resolutionId", &obj.ResolutionID)
	if err != nil {
		err = core.SDKErrorf(err, "", "resolutionId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "resolutionName", &obj.ResolutionName)
	if err != nil {
		err = core.SDKErrorf(err, "", "resolutionName-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "resolvedAlerts", &obj.ResolvedAlerts, UnmarshalResolvedAlertInfo)
	if err != nil {
		err = core.SDKErrorf(err, "", "resolvedAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "silenceMinutes", &obj.SilenceMinutes)
	if err != nil {
		err = core.SDKErrorf(err, "", "silenceMinutes-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "tenantId", &obj.TenantID)
	if err != nil {
		err = core.SDKErrorf(err, "", "tenantId-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// AlertResolutionDetails : Specifies information about the Alert Resolution.
type AlertResolutionDetails struct {
	// Specifies detailed notes about the Resolution.
	ResolutionDetails *string `json:"resolutionDetails,omitempty"`

	// Specifies the unique resolution id assigned in management console.
	ResolutionID *int64 `json:"resolutionId,omitempty"`

	// Specifies short description about the Resolution.
	ResolutionSummary *string `json:"resolutionSummary,omitempty"`

	// Specifies unix epoch timestamp (in microseconds) when the Alert was resolved.
	TimestampUsecs *int64 `json:"timestampUsecs,omitempty"`

	// Specifies name of the IBM Cluster user who resolved the Alerts.
	UserName *string `json:"userName,omitempty"`
}

// UnmarshalAlertResolutionDetails unmarshals an instance of AlertResolutionDetails from the specified map of raw messages.
func UnmarshalAlertResolutionDetails(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(AlertResolutionDetails)
	err = core.UnmarshalPrimitive(m, "resolutionDetails", &obj.ResolutionDetails)
	if err != nil {
		err = core.SDKErrorf(err, "", "resolutionDetails-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "resolutionId", &obj.ResolutionID)
	if err != nil {
		err = core.SDKErrorf(err, "", "resolutionId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "resolutionSummary", &obj.ResolutionSummary)
	if err != nil {
		err = core.SDKErrorf(err, "", "resolutionSummary-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "timestampUsecs", &obj.TimestampUsecs)
	if err != nil {
		err = core.SDKErrorf(err, "", "timestampUsecs-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "userName", &obj.UserName)
	if err != nil {
		err = core.SDKErrorf(err, "", "userName-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// AlertResolutionsList : Specifies the alert resolutions for get alert resolutions.
type AlertResolutionsList struct {
	// List of alert resolutions.
	AlertResolutionsList []AlertResolution `json:"alertResolutionsList,omitempty"`
}

// UnmarshalAlertResolutionsList unmarshals an instance of AlertResolutionsList from the specified map of raw messages.
func UnmarshalAlertResolutionsList(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(AlertResolutionsList)
	err = core.UnmarshalModel(m, "alertResolutionsList", &obj.AlertResolutionsList, UnmarshalAlertResolution)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertResolutionsList-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// AlertsList : Specifies the response of get alerts.
type AlertsList struct {
	// Specifies the list of alerts.
	AlertsList []Alert `json:"alertsList" validate:"required"`
}

// UnmarshalAlertsList unmarshals an instance of AlertsList from the specified map of raw messages.
func UnmarshalAlertsList(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(AlertsList)
	err = core.UnmarshalModel(m, "alertsList", &obj.AlertsList, UnmarshalAlert)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertsList-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// AlertsSummaryResponse : Specifies the response of alerts summary.
type AlertsSummaryResponse struct {
	// Specifies a list of alerts summary grouped by category.
	AlertsSummary []AlertGroupSummary `json:"alertsSummary,omitempty"`
}

// UnmarshalAlertsSummaryResponse unmarshals an instance of AlertsSummaryResponse from the specified map of raw messages.
func UnmarshalAlertsSummaryResponse(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(AlertsSummaryResponse)
	err = core.UnmarshalModel(m, "alertsSummary", &obj.AlertsSummary, UnmarshalAlertGroupSummary)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertsSummary-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// AuthHeaderForClusterUpgrade : Specifies a key-value pair for a header.
type AuthHeaderForClusterUpgrade struct {
	// Specifies the key or name of the header.
	Key *string `json:"key" validate:"required"`

	// Specifies the value of the header.
	Value *string `json:"value" validate:"required"`
}

// NewAuthHeaderForClusterUpgrade : Instantiate AuthHeaderForClusterUpgrade (Generic Model Constructor)
func (*BackupRecoveryManagementSreApiV1) NewAuthHeaderForClusterUpgrade(key string, value string) (_model *AuthHeaderForClusterUpgrade, err error) {
	_model = &AuthHeaderForClusterUpgrade{
		Key:   core.StringPtr(key),
		Value: core.StringPtr(value),
	}
	err = core.ValidateStruct(_model, "required parameters")
	if err != nil {
		err = core.SDKErrorf(err, "", "model-missing-required", common.GetComponentInfo())
	}
	return
}

// UnmarshalAuthHeaderForClusterUpgrade unmarshals an instance of AuthHeaderForClusterUpgrade from the specified map of raw messages.
func UnmarshalAuthHeaderForClusterUpgrade(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(AuthHeaderForClusterUpgrade)
	err = core.UnmarshalPrimitive(m, "key", &obj.Key)
	if err != nil {
		err = core.SDKErrorf(err, "", "key-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "value", &obj.Value)
	if err != nil {
		err = core.SDKErrorf(err, "", "value-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// AvailablePatchRelease : Specifies the details of the available patch release.
type AvailablePatchRelease struct {
	// Specifies patch release's notes.
	Notes *string `json:"notes,omitempty"`

	// Release's type.
	ReleaseType *string `json:"releaseType,omitempty"`

	// Specifies patch release's version.
	Version *string `json:"version,omitempty"`
}

// UnmarshalAvailablePatchRelease unmarshals an instance of AvailablePatchRelease from the specified map of raw messages.
func UnmarshalAvailablePatchRelease(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(AvailablePatchRelease)
	err = core.UnmarshalPrimitive(m, "notes", &obj.Notes)
	if err != nil {
		err = core.SDKErrorf(err, "", "notes-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "releaseType", &obj.ReleaseType)
	if err != nil {
		err = core.SDKErrorf(err, "", "releaseType-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "version", &obj.Version)
	if err != nil {
		err = core.SDKErrorf(err, "", "version-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// AvailableReleaseVersion : Specifies release information like release version, release notes and release upgrade URL.
type AvailableReleaseVersion struct {
	// Specifies release's notes.
	Notes *string `json:"notes,omitempty"`

	// Specifies the details of the available patch release.
	PatchDetails *AvailablePatchRelease `json:"patchDetails,omitempty"`

	// Specifies the stage of a release.
	ReleaseStage *string `json:"releaseStage,omitempty"`

	// Release's type e.g, LTS, Feature, Patch, MCM.
	ReleaseType *string `json:"releaseType,omitempty"`

	// Specifies the type of package or release.
	Type *string `json:"type,omitempty"`

	// Specifies release's version.
	Version *string `json:"version,omitempty"`
}

// Constants associated with the AvailableReleaseVersion.Type property.
// Specifies the type of package or release.
const (
	AvailableReleaseVersion_Type_Patch        = "Patch"
	AvailableReleaseVersion_Type_Upgrade      = "Upgrade"
	AvailableReleaseVersion_Type_Upgradepatch = "UpgradePatch"
)

// UnmarshalAvailableReleaseVersion unmarshals an instance of AvailableReleaseVersion from the specified map of raw messages.
func UnmarshalAvailableReleaseVersion(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(AvailableReleaseVersion)
	err = core.UnmarshalPrimitive(m, "notes", &obj.Notes)
	if err != nil {
		err = core.SDKErrorf(err, "", "notes-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "patchDetails", &obj.PatchDetails, UnmarshalAvailablePatchRelease)
	if err != nil {
		err = core.SDKErrorf(err, "", "patchDetails-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "releaseStage", &obj.ReleaseStage)
	if err != nil {
		err = core.SDKErrorf(err, "", "releaseStage-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "releaseType", &obj.ReleaseType)
	if err != nil {
		err = core.SDKErrorf(err, "", "releaseType-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "type", &obj.Type)
	if err != nil {
		err = core.SDKErrorf(err, "", "type-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "version", &obj.Version)
	if err != nil {
		err = core.SDKErrorf(err, "", "version-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// ClusterAlertStats : Specifies the cluster statistics based on active alerts.
type ClusterAlertStats struct {
	// Specifies the count of clusters with at least one critical alert.
	NumClustersWithCriticalAlerts *int64 `json:"numClustersWithCriticalAlerts,omitempty"`

	// Specifies the count of clusters with at least one warning category alert and no critical alerts.
	NumClustersWithWarningAlerts *int64 `json:"numClustersWithWarningAlerts,omitempty"`

	// Specifies the count of clusters with no warning or critical alerts.
	NumHealthyClusters *int64 `json:"numHealthyClusters,omitempty"`
}

// UnmarshalClusterAlertStats unmarshals an instance of ClusterAlertStats from the specified map of raw messages.
func UnmarshalClusterAlertStats(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(ClusterAlertStats)
	err = core.UnmarshalPrimitive(m, "numClustersWithCriticalAlerts", &obj.NumClustersWithCriticalAlerts)
	if err != nil {
		err = core.SDKErrorf(err, "", "numClustersWithCriticalAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numClustersWithWarningAlerts", &obj.NumClustersWithWarningAlerts)
	if err != nil {
		err = core.SDKErrorf(err, "", "numClustersWithWarningAlerts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numHealthyClusters", &obj.NumHealthyClusters)
	if err != nil {
		err = core.SDKErrorf(err, "", "numHealthyClusters-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// ClusterDetails : Specifies the array of clusters details including internally and externally claimed clusters.
type ClusterDetails struct {
	// Specifies the array of clusters upgrade details.
	CohesityClusters []ClusterInfo `json:"cohesityClusters,omitempty"`

	// Specifies the array of clusters claimed from IBM Storage Protect environment.
	SpClusters []SPClusterInfo `json:"spClusters,omitempty"`
}

// UnmarshalClusterDetails unmarshals an instance of ClusterDetails from the specified map of raw messages.
func UnmarshalClusterDetails(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(ClusterDetails)
	err = core.UnmarshalModel(m, "cohesityClusters", &obj.CohesityClusters, UnmarshalClusterInfo)
	if err != nil {
		err = core.SDKErrorf(err, "", "cohesityClusters-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "spClusters", &obj.SpClusters, UnmarshalSPClusterInfo)
	if err != nil {
		err = core.SDKErrorf(err, "", "spClusters-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// ClusterInfo : Specifies the clusters hardware type, memory used and total memory capacity, health, connected or not, current
// version, available versions and the upgrade status.
type ClusterInfo struct {
	// If cluster can support authHeader for upgrade or not.
	AuthSupportForPkgDownloads *bool `json:"authSupportForPkgDownloads,omitempty"`

	// Specifies the release versions the cluster can upgrade to.
	AvailableVersions []AvailableReleaseVersion `json:"availableVersions,omitempty"`

	// Specifies cluster id.
	ClusterID *int64 `json:"clusterId,omitempty"`

	// Specifies cluster incarnation id.
	ClusterIncarnationID *int64 `json:"clusterIncarnationId,omitempty"`

	// Specifies cluster's name.
	ClusterName *string `json:"clusterName,omitempty"`

	// Specifies current patch version of the cluster.
	CurrentPatchVersion *string `json:"currentPatchVersion,omitempty"`

	// Specifies if the cluster is connected to management console.
	CurrentVersion *string `json:"currentVersion,omitempty"`

	// Specifies the health of the cluster.
	Health *string `json:"health,omitempty"`

	// Specifies if the cluster is connected to management console.
	IsConnectedToHelios *bool `json:"isConnectedToHelios,omitempty"`

	// Specifies the location of the cluster.
	Location *string `json:"location,omitempty"`

	// Specifies if multi tenancy is enabled in the cluster.
	MultiTenancyEnabled *bool `json:"multiTenancyEnabled,omitempty"`

	// Specifies an array of node ips for the cluster.
	NodeIps []string `json:"nodeIps,omitempty"`

	// Specifies the number of nodes in the cluster.
	NumberOfNodes *int64 `json:"numberOfNodes,omitempty"`

	// Specifies the patch package URL for the cluster. This is populated for patch update only.
	PatchTargetUpgradeURL *string `json:"patchTargetUpgradeUrl,omitempty"`

	// Specifies target version to which clusters are upgrading. This is populated for patch update only.
	PatchTargetVersion *string `json:"patchTargetVersion,omitempty"`

	// Specifies the type of the cluster provider.
	ProviderType *string `json:"providerType,omitempty"`

	// Time at which an upgrade is scheduled.
	ScheduledTimestamp *int64 `json:"scheduledTimestamp,omitempty"`

	// Specifies the upgrade status of the cluster.
	Status *string `json:"status,omitempty"`

	// Specifies the upgrade URL for the cluster. This is populated for upgrade only.
	TargetUpgradeURL *string `json:"targetUpgradeUrl,omitempty"`

	// Specifies target version to which clusters are to be upgraded. This is populated for upgrade only.
	TargetVersion *string `json:"targetVersion,omitempty"`

	// Specifies how total memory capacity of the cluster.
	TotalCapacity *int64 `json:"totalCapacity,omitempty"`

	// Specifies the type of the cluster.
	Type *string `json:"type,omitempty"`

	// Specifies the type of upgrade performed on a cluster. This is to be used with status field to know the status of the
	// upgrade action performed on cluster.
	UpdateType *string `json:"updateType,omitempty"`

	// Specifies how much of the cluster capacity is consumed.
	UsedCapacity *int64 `json:"usedCapacity,omitempty"`
}

// Constants associated with the ClusterInfo.Health property.
// Specifies the health of the cluster.
const (
	ClusterInfo_Health_Critical    = "Critical"
	ClusterInfo_Health_Noncritical = "NonCritical"
)

// Constants associated with the ClusterInfo.ProviderType property.
// Specifies the type of the cluster provider.
const (
	ClusterInfo_ProviderType_Kibmstorageprotect = "kIBMStorageProtect"
	ClusterInfo_ProviderType_Kmcmcohesity       = "kMCMCohesity"
)

// Constants associated with the ClusterInfo.Status property.
// Specifies the upgrade status of the cluster.
const (
	ClusterInfo_Status_Clusterunreachable = "ClusterUnreachable"
	ClusterInfo_Status_Failed             = "Failed"
	ClusterInfo_Status_Inprogress         = "InProgress"
	ClusterInfo_Status_Scheduled          = "Scheduled"
	ClusterInfo_Status_Upgradeavailable   = "UpgradeAvailable"
	ClusterInfo_Status_Uptodate           = "UpToDate"
)

// Constants associated with the ClusterInfo.Type property.
// Specifies the type of the cluster.
const (
	ClusterInfo_Type_Physical = "Physical"
	ClusterInfo_Type_Vmrobo   = "VMRobo"
)

// Constants associated with the ClusterInfo.UpdateType property.
// Specifies the type of upgrade performed on a cluster. This is to be used with status field to know the status of the
// upgrade action performed on cluster.
const (
	ClusterInfo_UpdateType_Patch        = "Patch"
	ClusterInfo_UpdateType_Upgrade      = "Upgrade"
	ClusterInfo_UpdateType_Upgradepatch = "UpgradePatch"
)

// UnmarshalClusterInfo unmarshals an instance of ClusterInfo from the specified map of raw messages.
func UnmarshalClusterInfo(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(ClusterInfo)
	err = core.UnmarshalPrimitive(m, "authSupportForPkgDownloads", &obj.AuthSupportForPkgDownloads)
	if err != nil {
		err = core.SDKErrorf(err, "", "authSupportForPkgDownloads-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "availableVersions", &obj.AvailableVersions, UnmarshalAvailableReleaseVersion)
	if err != nil {
		err = core.SDKErrorf(err, "", "availableVersions-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "clusterId", &obj.ClusterID)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "clusterIncarnationId", &obj.ClusterIncarnationID)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterIncarnationId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "clusterName", &obj.ClusterName)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterName-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "currentPatchVersion", &obj.CurrentPatchVersion)
	if err != nil {
		err = core.SDKErrorf(err, "", "currentPatchVersion-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "currentVersion", &obj.CurrentVersion)
	if err != nil {
		err = core.SDKErrorf(err, "", "currentVersion-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "health", &obj.Health)
	if err != nil {
		err = core.SDKErrorf(err, "", "health-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "isConnectedToHelios", &obj.IsConnectedToHelios)
	if err != nil {
		err = core.SDKErrorf(err, "", "isConnectedToHelios-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "location", &obj.Location)
	if err != nil {
		err = core.SDKErrorf(err, "", "location-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "multiTenancyEnabled", &obj.MultiTenancyEnabled)
	if err != nil {
		err = core.SDKErrorf(err, "", "multiTenancyEnabled-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "nodeIps", &obj.NodeIps)
	if err != nil {
		err = core.SDKErrorf(err, "", "nodeIps-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numberOfNodes", &obj.NumberOfNodes)
	if err != nil {
		err = core.SDKErrorf(err, "", "numberOfNodes-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "patchTargetUpgradeUrl", &obj.PatchTargetUpgradeURL)
	if err != nil {
		err = core.SDKErrorf(err, "", "patchTargetUpgradeUrl-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "patchTargetVersion", &obj.PatchTargetVersion)
	if err != nil {
		err = core.SDKErrorf(err, "", "patchTargetVersion-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "providerType", &obj.ProviderType)
	if err != nil {
		err = core.SDKErrorf(err, "", "providerType-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "scheduledTimestamp", &obj.ScheduledTimestamp)
	if err != nil {
		err = core.SDKErrorf(err, "", "scheduledTimestamp-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "status", &obj.Status)
	if err != nil {
		err = core.SDKErrorf(err, "", "status-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "targetUpgradeUrl", &obj.TargetUpgradeURL)
	if err != nil {
		err = core.SDKErrorf(err, "", "targetUpgradeUrl-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "targetVersion", &obj.TargetVersion)
	if err != nil {
		err = core.SDKErrorf(err, "", "targetVersion-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "totalCapacity", &obj.TotalCapacity)
	if err != nil {
		err = core.SDKErrorf(err, "", "totalCapacity-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "type", &obj.Type)
	if err != nil {
		err = core.SDKErrorf(err, "", "type-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "updateType", &obj.UpdateType)
	if err != nil {
		err = core.SDKErrorf(err, "", "updateType-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "usedCapacity", &obj.UsedCapacity)
	if err != nil {
		err = core.SDKErrorf(err, "", "usedCapacity-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// ClustersUpgradesInfoOptions : The ClustersUpgradesInfo options.
type ClustersUpgradesInfoOptions struct {
	// Fetch upgrade progress details for a list of cluster identifiers in format clusterId:clusterIncarnationId.
	ClusterIdentifiers []string `json:"clusterIdentifiers,omitempty"`

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// NewClustersUpgradesInfoOptions : Instantiate ClustersUpgradesInfoOptions
func (*BackupRecoveryManagementSreApiV1) NewClustersUpgradesInfoOptions() *ClustersUpgradesInfoOptions {
	return &ClustersUpgradesInfoOptions{}
}

// SetClusterIdentifiers : Allow user to set ClusterIdentifiers
func (_options *ClustersUpgradesInfoOptions) SetClusterIdentifiers(clusterIdentifiers []string) *ClustersUpgradesInfoOptions {
	_options.ClusterIdentifiers = clusterIdentifiers
	return _options
}

// SetHeaders : Allow user to set Headers
func (options *ClustersUpgradesInfoOptions) SetHeaders(param map[string]string) *ClustersUpgradesInfoOptions {
	options.Headers = param
	return options
}

// CompatibleCluster : Specifies a cluster compatible for a release.
type CompatibleCluster struct {
	// Specifies cluster id.
	ClusterID *int64 `json:"clusterId,omitempty"`

	// Specifies cluster incarnation id.
	ClusterIncarnationID *int64 `json:"clusterIncarnationId,omitempty"`

	// Specifies cluster's name.
	ClusterName *string `json:"clusterName,omitempty"`

	// Specifies the current version of the cluster.
	CurrentVersion *string `json:"currentVersion,omitempty"`
}

// UnmarshalCompatibleCluster unmarshals an instance of CompatibleCluster from the specified map of raw messages.
func UnmarshalCompatibleCluster(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(CompatibleCluster)
	err = core.UnmarshalPrimitive(m, "clusterId", &obj.ClusterID)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "clusterIncarnationId", &obj.ClusterIncarnationID)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterIncarnationId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "clusterName", &obj.ClusterName)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterName-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "currentVersion", &obj.CurrentVersion)
	if err != nil {
		err = core.SDKErrorf(err, "", "currentVersion-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// CompatibleClustersForReleaseOptions : The CompatibleClustersForRelease options.
type CompatibleClustersForReleaseOptions struct {
	ReleaseVersion *string `json:"releaseVersion,omitempty"`

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// NewCompatibleClustersForReleaseOptions : Instantiate CompatibleClustersForReleaseOptions
func (*BackupRecoveryManagementSreApiV1) NewCompatibleClustersForReleaseOptions() *CompatibleClustersForReleaseOptions {
	return &CompatibleClustersForReleaseOptions{}
}

// SetReleaseVersion : Allow user to set ReleaseVersion
func (_options *CompatibleClustersForReleaseOptions) SetReleaseVersion(releaseVersion string) *CompatibleClustersForReleaseOptions {
	_options.ReleaseVersion = core.StringPtr(releaseVersion)
	return _options
}

// SetHeaders : Allow user to set Headers
func (options *CompatibleClustersForReleaseOptions) SetHeaders(param map[string]string) *CompatibleClustersForReleaseOptions {
	options.Headers = param
	return options
}

// CreateClustersUpgradesOptions : The CreateClustersUpgrades options.
type CreateClustersUpgradesOptions struct {
	// Specifies the optional headers for upgrade request.
	AuthHeaders []AuthHeaderForClusterUpgrade `json:"authHeaders,omitempty"`

	// Array for clusters to be upgraded.
	Clusters []Upgrade `json:"clusters,omitempty"`

	// Specifies the difference of time between two cluster's upgrade.
	IntervalForRollingUpgradeInHours *int64 `json:"intervalForRollingUpgradeInHours,omitempty"`

	// Specifies URL from which package can be downloaded. Note: This option is only supported in Multi-Cluster Manager
	// (MCM).
	PackageURL *string `json:"packageUrl,omitempty"`

	// Specifies the parameters for patch upgrade request.
	PatchUpgradeParams *PatchUpgradeParams `json:"patchUpgradeParams,omitempty"`

	// Specifies target version to which clusters are to be upgraded.
	TargetVersion *string `json:"targetVersion,omitempty"`

	// Specifies the time in msecs at which the cluster has to be upgraded.
	TimeStampToUpgradeAtMsecs *int64 `json:"timeStampToUpgradeAtMsecs,omitempty"`

	// Specifies the type of upgrade to be performed on a cluster.
	Type *string `json:"type,omitempty"`

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// Constants associated with the CreateClustersUpgradesOptions.Type property.
// Specifies the type of upgrade to be performed on a cluster.
const (
	CreateClustersUpgradesOptions_Type_Patch        = "Patch"
	CreateClustersUpgradesOptions_Type_Upgrade      = "Upgrade"
	CreateClustersUpgradesOptions_Type_Upgradepatch = "UpgradePatch"
)

// NewCreateClustersUpgradesOptions : Instantiate CreateClustersUpgradesOptions
func (*BackupRecoveryManagementSreApiV1) NewCreateClustersUpgradesOptions() *CreateClustersUpgradesOptions {
	return &CreateClustersUpgradesOptions{}
}

// SetAuthHeaders : Allow user to set AuthHeaders
func (_options *CreateClustersUpgradesOptions) SetAuthHeaders(authHeaders []AuthHeaderForClusterUpgrade) *CreateClustersUpgradesOptions {
	_options.AuthHeaders = authHeaders
	return _options
}

// SetClusters : Allow user to set Clusters
func (_options *CreateClustersUpgradesOptions) SetClusters(clusters []Upgrade) *CreateClustersUpgradesOptions {
	_options.Clusters = clusters
	return _options
}

// SetIntervalForRollingUpgradeInHours : Allow user to set IntervalForRollingUpgradeInHours
func (_options *CreateClustersUpgradesOptions) SetIntervalForRollingUpgradeInHours(intervalForRollingUpgradeInHours int64) *CreateClustersUpgradesOptions {
	_options.IntervalForRollingUpgradeInHours = core.Int64Ptr(intervalForRollingUpgradeInHours)
	return _options
}

// SetPackageURL : Allow user to set PackageURL
func (_options *CreateClustersUpgradesOptions) SetPackageURL(packageURL string) *CreateClustersUpgradesOptions {
	_options.PackageURL = core.StringPtr(packageURL)
	return _options
}

// SetPatchUpgradeParams : Allow user to set PatchUpgradeParams
func (_options *CreateClustersUpgradesOptions) SetPatchUpgradeParams(patchUpgradeParams *PatchUpgradeParams) *CreateClustersUpgradesOptions {
	_options.PatchUpgradeParams = patchUpgradeParams
	return _options
}

// SetTargetVersion : Allow user to set TargetVersion
func (_options *CreateClustersUpgradesOptions) SetTargetVersion(targetVersion string) *CreateClustersUpgradesOptions {
	_options.TargetVersion = core.StringPtr(targetVersion)
	return _options
}

// SetTimeStampToUpgradeAtMsecs : Allow user to set TimeStampToUpgradeAtMsecs
func (_options *CreateClustersUpgradesOptions) SetTimeStampToUpgradeAtMsecs(timeStampToUpgradeAtMsecs int64) *CreateClustersUpgradesOptions {
	_options.TimeStampToUpgradeAtMsecs = core.Int64Ptr(timeStampToUpgradeAtMsecs)
	return _options
}

// SetType : Allow user to set Type
func (_options *CreateClustersUpgradesOptions) SetType(typeVar string) *CreateClustersUpgradesOptions {
	_options.Type = core.StringPtr(typeVar)
	return _options
}

// SetHeaders : Allow user to set Headers
func (options *CreateClustersUpgradesOptions) SetHeaders(param map[string]string) *CreateClustersUpgradesOptions {
	options.Headers = param
	return options
}

// DeleteClustersUpgradesOptions : The DeleteClustersUpgrades options.
type DeleteClustersUpgradesOptions struct {
	// Specifies the list of cluster identifiers. The format is clusterId:clusterIncarnationId.
	ClusterIdentifiers []string `json:"clusterIdentifiers,omitempty"`

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// NewDeleteClustersUpgradesOptions : Instantiate DeleteClustersUpgradesOptions
func (*BackupRecoveryManagementSreApiV1) NewDeleteClustersUpgradesOptions() *DeleteClustersUpgradesOptions {
	return &DeleteClustersUpgradesOptions{}
}

// SetClusterIdentifiers : Allow user to set ClusterIdentifiers
func (_options *DeleteClustersUpgradesOptions) SetClusterIdentifiers(clusterIdentifiers []string) *DeleteClustersUpgradesOptions {
	_options.ClusterIdentifiers = clusterIdentifiers
	return _options
}

// SetHeaders : Allow user to set Headers
func (options *DeleteClustersUpgradesOptions) SetHeaders(param map[string]string) *DeleteClustersUpgradesOptions {
	options.Headers = param
	return options
}

// GetAlertSummaryOptions : The GetAlertSummary options.
type GetAlertSummaryOptions struct {
	// Filter by start time. Specify the start time as a Unix epoch Timestamp (in microseconds). By default it is current
	// time minus a day.
	StartTimeUsecs *int64 `json:"startTimeUsecs,omitempty"`

	// Filter by end time. Specify the end time as a Unix epoch Timestamp (in microseconds). By default it is current time.
	EndTimeUsecs *int64 `json:"endTimeUsecs,omitempty"`

	// IncludeTenants specifies if alerts of all the tenants under the hierarchy of the logged in user's organization
	// should be used to compute summary.
	IncludeTenants *bool `json:"includeTenants,omitempty"`

	// TenantIds contains ids of the tenants for which alerts are to be used to compute summary.
	TenantIds []string `json:"tenantIds,omitempty"`

	// Specifies list of alert states to filter alerts by. If not specified, only open alerts will be used to get summary.
	StatesList []string `json:"statesList,omitempty"`

	// This field uniquely represents a service        instance. Please specify the values as "service-instance-id:
	// <value>".
	XScopeIdentifier *string `json:"X-Scope-Identifier,omitempty"`

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// Constants associated with the GetAlertSummaryOptions.StatesList property.
const (
	GetAlertSummaryOptions_StatesList_Knote       = "kNote"
	GetAlertSummaryOptions_StatesList_Kopen       = "kOpen"
	GetAlertSummaryOptions_StatesList_Kresolved   = "kResolved"
	GetAlertSummaryOptions_StatesList_Ksuppressed = "kSuppressed"
)

// NewGetAlertSummaryOptions : Instantiate GetAlertSummaryOptions
func (*BackupRecoveryManagementSreApiV1) NewGetAlertSummaryOptions() *GetAlertSummaryOptions {
	return &GetAlertSummaryOptions{}
}

// SetStartTimeUsecs : Allow user to set StartTimeUsecs
func (_options *GetAlertSummaryOptions) SetStartTimeUsecs(startTimeUsecs int64) *GetAlertSummaryOptions {
	_options.StartTimeUsecs = core.Int64Ptr(startTimeUsecs)
	return _options
}

// SetEndTimeUsecs : Allow user to set EndTimeUsecs
func (_options *GetAlertSummaryOptions) SetEndTimeUsecs(endTimeUsecs int64) *GetAlertSummaryOptions {
	_options.EndTimeUsecs = core.Int64Ptr(endTimeUsecs)
	return _options
}

// SetIncludeTenants : Allow user to set IncludeTenants
func (_options *GetAlertSummaryOptions) SetIncludeTenants(includeTenants bool) *GetAlertSummaryOptions {
	_options.IncludeTenants = core.BoolPtr(includeTenants)
	return _options
}

// SetTenantIds : Allow user to set TenantIds
func (_options *GetAlertSummaryOptions) SetTenantIds(tenantIds []string) *GetAlertSummaryOptions {
	_options.TenantIds = tenantIds
	return _options
}

// SetStatesList : Allow user to set StatesList
func (_options *GetAlertSummaryOptions) SetStatesList(statesList []string) *GetAlertSummaryOptions {
	_options.StatesList = statesList
	return _options
}

// SetXScopeIdentifier : Allow user to set XScopeIdentifier
func (_options *GetAlertSummaryOptions) SetXScopeIdentifier(xScopeIdentifier string) *GetAlertSummaryOptions {
	_options.XScopeIdentifier = core.StringPtr(xScopeIdentifier)
	return _options
}

// SetHeaders : Allow user to set Headers
func (options *GetAlertSummaryOptions) SetHeaders(param map[string]string) *GetAlertSummaryOptions {
	options.Headers = param
	return options
}

// GetAlertsOptions : The GetAlerts options.
type GetAlertsOptions struct {
	// Filter by list of alert ids.
	AlertIds []string `json:"alertIds,omitempty"`

	// Filter by list of alert types.
	AlertTypes []int64 `json:"alertTypes,omitempty"`

	// Filter by list of alert categories.
	AlertCategories []string `json:"alertCategories,omitempty"`

	// Filter by list of alert states.
	AlertStates []string `json:"alertStates,omitempty"`

	// Filter by list of alert severity types.
	AlertSeverities []string `json:"alertSeverities,omitempty"`

	// Filter by list of alert type buckets.
	AlertTypeBuckets []string `json:"alertTypeBuckets,omitempty"`

	// Specifies start time Unix epoch time in microseconds to filter alerts by.
	StartTimeUsecs *int64 `json:"startTimeUsecs,omitempty"`

	// Specifies end time Unix epoch time in microseconds to filter alerts by.
	EndTimeUsecs *int64 `json:"endTimeUsecs,omitempty"`

	// Specifies maximum number of alerts to return.The default value is 100 and maximum allowed value is 1000.
	MaxAlerts *int64 `json:"maxAlerts,omitempty"`

	// Specifies name of the property to filter alerts by.
	PropertyKey *string `json:"propertyKey,omitempty"`

	// Specifies value of the property to filter alerts by.
	PropertyValue *string `json:"propertyValue,omitempty"`

	// Specifies name of alert to filter alerts by.
	AlertName *string `json:"alertName,omitempty"`

	// Specifies alert resolution ids to filter alerts by.
	ResolutionIds []int64 `json:"resolutionIds,omitempty"`

	// Filter by tenant ids.
	TenantIds []string `json:"tenantIds,omitempty"`

	// Filter by objects of all the tenants under the hierarchy of the logged in user's organization.
	AllUnderHierarchy *bool `json:"allUnderHierarchy,omitempty"`

	// This field uniquely represents a service        instance. Please specify the values as "service-instance-id:
	// <value>".
	XScopeIdentifier *string `json:"X-Scope-Identifier,omitempty"`

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// Constants associated with the GetAlertsOptions.AlertCategories property.
const (
	GetAlertsOptions_AlertCategories_Kagent                  = "kAgent"
	GetAlertsOptions_AlertCategories_Kappmarketplace         = "kAppMarketPlace"
	GetAlertsOptions_AlertCategories_Karchivalrestore        = "kArchivalRestore"
	GetAlertsOptions_AlertCategories_Kauditlog               = "kAuditLog"
	GetAlertsOptions_AlertCategories_Kbackuprestore          = "kBackupRestore"
	GetAlertsOptions_AlertCategories_Kcdp                    = "kCDP"
	GetAlertsOptions_AlertCategories_Kchassis                = "kChassis"
	GetAlertsOptions_AlertCategories_Kcluster                = "kCluster"
	GetAlertsOptions_AlertCategories_Kclustermanagement      = "kClusterManagement"
	GetAlertsOptions_AlertCategories_Kconfiguration          = "kConfiguration"
	GetAlertsOptions_AlertCategories_Kcpu                    = "kCPU"
	GetAlertsOptions_AlertCategories_Kdatapath               = "kDataPath"
	GetAlertsOptions_AlertCategories_Kdisasterrecovery       = "kDisasterRecovery"
	GetAlertsOptions_AlertCategories_Kdisk                   = "kDisk"
	GetAlertsOptions_AlertCategories_Kfan                    = "kFan"
	GetAlertsOptions_AlertCategories_Kfaulttolerance         = "kFaultTolerance"
	GetAlertsOptions_AlertCategories_Kfirmware               = "kFirmware"
	GetAlertsOptions_AlertCategories_Kgeneralsoftwarefailure = "kGeneralSoftwareFailure"
	GetAlertsOptions_AlertCategories_Khelios                 = "kHelios"
	GetAlertsOptions_AlertCategories_Kindexing               = "kIndexing"
	GetAlertsOptions_AlertCategories_Klicense                = "kLicense"
	GetAlertsOptions_AlertCategories_Kmemory                 = "kMemory"
	GetAlertsOptions_AlertCategories_Kmetadata               = "kMetadata"
	GetAlertsOptions_AlertCategories_Knetworking             = "kNetworking"
	GetAlertsOptions_AlertCategories_Knic                    = "kNIC"
	GetAlertsOptions_AlertCategories_Knode                   = "kNode"
	GetAlertsOptions_AlertCategories_Knodehealth             = "kNodeHealth"
	GetAlertsOptions_AlertCategories_Koperatingsystem        = "kOperatingSystem"
	GetAlertsOptions_AlertCategories_Kpowersupply            = "kPowerSupply"
	GetAlertsOptions_AlertCategories_Kquota                  = "kQuota"
	GetAlertsOptions_AlertCategories_Kremotereplication      = "kRemoteReplication"
	GetAlertsOptions_AlertCategories_Ksecurity               = "kSecurity"
	GetAlertsOptions_AlertCategories_Kstoragedevice          = "kStorageDevice"
	GetAlertsOptions_AlertCategories_Kstoragepool            = "kStoragePool"
	GetAlertsOptions_AlertCategories_Kstorageusage           = "kStorageUsage"
	GetAlertsOptions_AlertCategories_Ksystemservice          = "kSystemService"
	GetAlertsOptions_AlertCategories_Ktemperature            = "kTemperature"
	GetAlertsOptions_AlertCategories_Kupgrade                = "kUpgrade"
	GetAlertsOptions_AlertCategories_Kviewfailover           = "kViewFailover"
)

// Constants associated with the GetAlertsOptions.AlertStates property.
const (
	GetAlertsOptions_AlertStates_Knote       = "kNote"
	GetAlertsOptions_AlertStates_Kopen       = "kOpen"
	GetAlertsOptions_AlertStates_Kresolved   = "kResolved"
	GetAlertsOptions_AlertStates_Ksuppressed = "kSuppressed"
)

// Constants associated with the GetAlertsOptions.AlertSeverities property.
const (
	GetAlertsOptions_AlertSeverities_Kcritical = "kCritical"
	GetAlertsOptions_AlertSeverities_Kinfo     = "kInfo"
	GetAlertsOptions_AlertSeverities_Kwarning  = "kWarning"
)

// Constants associated with the GetAlertsOptions.AlertTypeBuckets property.
const (
	GetAlertsOptions_AlertTypeBuckets_Kdataservice = "kDataService"
	GetAlertsOptions_AlertTypeBuckets_Khardware    = "kHardware"
	GetAlertsOptions_AlertTypeBuckets_Kmaintenance = "kMaintenance"
	GetAlertsOptions_AlertTypeBuckets_Ksoftware    = "kSoftware"
)

// NewGetAlertsOptions : Instantiate GetAlertsOptions
func (*BackupRecoveryManagementSreApiV1) NewGetAlertsOptions() *GetAlertsOptions {
	return &GetAlertsOptions{}
}

// SetAlertIds : Allow user to set AlertIds
func (_options *GetAlertsOptions) SetAlertIds(alertIds []string) *GetAlertsOptions {
	_options.AlertIds = alertIds
	return _options
}

// SetAlertTypes : Allow user to set AlertTypes
func (_options *GetAlertsOptions) SetAlertTypes(alertTypes []int64) *GetAlertsOptions {
	_options.AlertTypes = alertTypes
	return _options
}

// SetAlertCategories : Allow user to set AlertCategories
func (_options *GetAlertsOptions) SetAlertCategories(alertCategories []string) *GetAlertsOptions {
	_options.AlertCategories = alertCategories
	return _options
}

// SetAlertStates : Allow user to set AlertStates
func (_options *GetAlertsOptions) SetAlertStates(alertStates []string) *GetAlertsOptions {
	_options.AlertStates = alertStates
	return _options
}

// SetAlertSeverities : Allow user to set AlertSeverities
func (_options *GetAlertsOptions) SetAlertSeverities(alertSeverities []string) *GetAlertsOptions {
	_options.AlertSeverities = alertSeverities
	return _options
}

// SetAlertTypeBuckets : Allow user to set AlertTypeBuckets
func (_options *GetAlertsOptions) SetAlertTypeBuckets(alertTypeBuckets []string) *GetAlertsOptions {
	_options.AlertTypeBuckets = alertTypeBuckets
	return _options
}

// SetStartTimeUsecs : Allow user to set StartTimeUsecs
func (_options *GetAlertsOptions) SetStartTimeUsecs(startTimeUsecs int64) *GetAlertsOptions {
	_options.StartTimeUsecs = core.Int64Ptr(startTimeUsecs)
	return _options
}

// SetEndTimeUsecs : Allow user to set EndTimeUsecs
func (_options *GetAlertsOptions) SetEndTimeUsecs(endTimeUsecs int64) *GetAlertsOptions {
	_options.EndTimeUsecs = core.Int64Ptr(endTimeUsecs)
	return _options
}

// SetMaxAlerts : Allow user to set MaxAlerts
func (_options *GetAlertsOptions) SetMaxAlerts(maxAlerts int64) *GetAlertsOptions {
	_options.MaxAlerts = core.Int64Ptr(maxAlerts)
	return _options
}

// SetPropertyKey : Allow user to set PropertyKey
func (_options *GetAlertsOptions) SetPropertyKey(propertyKey string) *GetAlertsOptions {
	_options.PropertyKey = core.StringPtr(propertyKey)
	return _options
}

// SetPropertyValue : Allow user to set PropertyValue
func (_options *GetAlertsOptions) SetPropertyValue(propertyValue string) *GetAlertsOptions {
	_options.PropertyValue = core.StringPtr(propertyValue)
	return _options
}

// SetAlertName : Allow user to set AlertName
func (_options *GetAlertsOptions) SetAlertName(alertName string) *GetAlertsOptions {
	_options.AlertName = core.StringPtr(alertName)
	return _options
}

// SetResolutionIds : Allow user to set ResolutionIds
func (_options *GetAlertsOptions) SetResolutionIds(resolutionIds []int64) *GetAlertsOptions {
	_options.ResolutionIds = resolutionIds
	return _options
}

// SetTenantIds : Allow user to set TenantIds
func (_options *GetAlertsOptions) SetTenantIds(tenantIds []string) *GetAlertsOptions {
	_options.TenantIds = tenantIds
	return _options
}

// SetAllUnderHierarchy : Allow user to set AllUnderHierarchy
func (_options *GetAlertsOptions) SetAllUnderHierarchy(allUnderHierarchy bool) *GetAlertsOptions {
	_options.AllUnderHierarchy = core.BoolPtr(allUnderHierarchy)
	return _options
}

// SetXScopeIdentifier : Allow user to set XScopeIdentifier
func (_options *GetAlertsOptions) SetXScopeIdentifier(xScopeIdentifier string) *GetAlertsOptions {
	_options.XScopeIdentifier = core.StringPtr(xScopeIdentifier)
	return _options
}

// SetHeaders : Allow user to set Headers
func (options *GetAlertsOptions) SetHeaders(param map[string]string) *GetAlertsOptions {
	options.Headers = param
	return options
}

// GetClustersInfoOptions : The GetClustersInfo options.
type GetClustersInfoOptions struct {

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// NewGetClustersInfoOptions : Instantiate GetClustersInfoOptions
func (*BackupRecoveryManagementSreApiV1) NewGetClustersInfoOptions() *GetClustersInfoOptions {
	return &GetClustersInfoOptions{}
}

// SetHeaders : Allow user to set Headers
func (options *GetClustersInfoOptions) SetHeaders(param map[string]string) *GetClustersInfoOptions {
	options.Headers = param
	return options
}

// GetManagementAlertResolutionOptions : The GetManagementAlertResolution options.
type GetManagementAlertResolutionOptions struct {
	// Specifies the max number of Resolutions to be returned, from the latest created to the earliest created.
	MaxResolutions *int64 `json:"maxResolutions" validate:"required"`

	// Specifies Alert Resolution Name to query.
	ResolutionName *string `json:"resolutionName,omitempty"`

	// Specifies Alert Resolution id to query.
	ResolutionID *string `json:"resolutionId,omitempty"`

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// NewGetManagementAlertResolutionOptions : Instantiate GetManagementAlertResolutionOptions
func (*BackupRecoveryManagementSreApiV1) NewGetManagementAlertResolutionOptions(maxResolutions int64) *GetManagementAlertResolutionOptions {
	return &GetManagementAlertResolutionOptions{
		MaxResolutions: core.Int64Ptr(maxResolutions),
	}
}

// SetMaxResolutions : Allow user to set MaxResolutions
func (_options *GetManagementAlertResolutionOptions) SetMaxResolutions(maxResolutions int64) *GetManagementAlertResolutionOptions {
	_options.MaxResolutions = core.Int64Ptr(maxResolutions)
	return _options
}

// SetResolutionName : Allow user to set ResolutionName
func (_options *GetManagementAlertResolutionOptions) SetResolutionName(resolutionName string) *GetManagementAlertResolutionOptions {
	_options.ResolutionName = core.StringPtr(resolutionName)
	return _options
}

// SetResolutionID : Allow user to set ResolutionID
func (_options *GetManagementAlertResolutionOptions) SetResolutionID(resolutionID string) *GetManagementAlertResolutionOptions {
	_options.ResolutionID = core.StringPtr(resolutionID)
	return _options
}

// SetHeaders : Allow user to set Headers
func (options *GetManagementAlertResolutionOptions) SetHeaders(param map[string]string) *GetManagementAlertResolutionOptions {
	options.Headers = param
	return options
}

// GetManagementAlertsOptions : The GetManagementAlerts options.
type GetManagementAlertsOptions struct {
	// Filter by list of alert ids.
	AlertIdList []string `json:"alertIdList,omitempty"`

	// Filter by list of alert states.
	AlertStateList []string `json:"alertStateList,omitempty"`

	// Filter by list of alert types.
	AlertTypeList []int64 `json:"alertTypeList,omitempty"`

	// Filter by list of alert severity types.
	AlertSeverityList []string `json:"alertSeverityList,omitempty"`

	// Filter by list of region ids.
	RegionIds []string `json:"regionIds,omitempty"`

	// Filter by list of cluster ids.
	ClusterIdentifiers []string `json:"clusterIdentifiers,omitempty"`

	// Specifies the start time of the alerts to be returned. All the alerts returned are raised after the specified start
	// time. This value should be in Unix timestamp epoch in microseconds.
	StartDateUsecs *int64 `json:"startDateUsecs,omitempty"`

	// Specifies the end time of the alerts to be returned. All the alerts returned are raised before the specified end
	// time. This value should be in Unix timestamp epoch in microseconds.
	EndDateUsecs *int64 `json:"endDateUsecs,omitempty"`

	// Specifies maximum number of alerts to return.
	MaxAlerts *int64 `json:"maxAlerts,omitempty"`

	// Filter by list of alert categories.
	AlertCategoryList []string `json:"alertCategoryList,omitempty"`

	// Filter by tenant ids.
	TenantIds []string `json:"tenantIds,omitempty"`

	// Filter by list of alert type buckets.
	AlertTypeBucketList []string `json:"alertTypeBucketList,omitempty"`

	// Specifies list of the alert property keys to query.
	AlertPropertyKeyList []string `json:"alertPropertyKeyList,omitempty"`

	// Specifies list of the alert property value, multiple values for one key should be joined by '|'.
	AlertPropertyValueList []string `json:"alertPropertyValueList,omitempty"`

	// Specifies name of alert to filter alerts by.
	AlertName *string `json:"alertName,omitempty"`

	// Specifies services instance ids to filter alerts for IBM customers.
	ServiceInstanceIds []string `json:"serviceInstanceIds,omitempty"`

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// Constants associated with the GetManagementAlertsOptions.AlertSeverityList property.
const (
	GetManagementAlertsOptions_AlertSeverityList_Kcritical = "kCritical"
	GetManagementAlertsOptions_AlertSeverityList_Kinfo     = "kInfo"
	GetManagementAlertsOptions_AlertSeverityList_Kwarning  = "kWarning"
)

// Constants associated with the GetManagementAlertsOptions.AlertCategoryList property.
const (
	GetManagementAlertsOptions_AlertCategoryList_Kagent                  = "kAgent"
	GetManagementAlertsOptions_AlertCategoryList_Kappmarketplace         = "kAppMarketPlace"
	GetManagementAlertsOptions_AlertCategoryList_Karchivalrestore        = "kArchivalRestore"
	GetManagementAlertsOptions_AlertCategoryList_Kauditlog               = "kAuditLog"
	GetManagementAlertsOptions_AlertCategoryList_Kbackuprestore          = "kBackupRestore"
	GetManagementAlertsOptions_AlertCategoryList_Kcdp                    = "kCDP"
	GetManagementAlertsOptions_AlertCategoryList_Kchassis                = "kChassis"
	GetManagementAlertsOptions_AlertCategoryList_Kcluster                = "kCluster"
	GetManagementAlertsOptions_AlertCategoryList_Kclustermanagement      = "kClusterManagement"
	GetManagementAlertsOptions_AlertCategoryList_Kconfiguration          = "kConfiguration"
	GetManagementAlertsOptions_AlertCategoryList_Kcpu                    = "kCPU"
	GetManagementAlertsOptions_AlertCategoryList_Kdatapath               = "kDataPath"
	GetManagementAlertsOptions_AlertCategoryList_Kdisasterrecovery       = "kDisasterRecovery"
	GetManagementAlertsOptions_AlertCategoryList_Kdisk                   = "kDisk"
	GetManagementAlertsOptions_AlertCategoryList_Kfan                    = "kFan"
	GetManagementAlertsOptions_AlertCategoryList_Kfaulttolerance         = "kFaultTolerance"
	GetManagementAlertsOptions_AlertCategoryList_Kfirmware               = "kFirmware"
	GetManagementAlertsOptions_AlertCategoryList_Kgeneralsoftwarefailure = "kGeneralSoftwareFailure"
	GetManagementAlertsOptions_AlertCategoryList_Khelios                 = "kHelios"
	GetManagementAlertsOptions_AlertCategoryList_Kindexing               = "kIndexing"
	GetManagementAlertsOptions_AlertCategoryList_Klicense                = "kLicense"
	GetManagementAlertsOptions_AlertCategoryList_Kmemory                 = "kMemory"
	GetManagementAlertsOptions_AlertCategoryList_Kmetadata               = "kMetadata"
	GetManagementAlertsOptions_AlertCategoryList_Knetworking             = "kNetworking"
	GetManagementAlertsOptions_AlertCategoryList_Knic                    = "kNIC"
	GetManagementAlertsOptions_AlertCategoryList_Knode                   = "kNode"
	GetManagementAlertsOptions_AlertCategoryList_Knodehealth             = "kNodeHealth"
	GetManagementAlertsOptions_AlertCategoryList_Koperatingsystem        = "kOperatingSystem"
	GetManagementAlertsOptions_AlertCategoryList_Kpowersupply            = "kPowerSupply"
	GetManagementAlertsOptions_AlertCategoryList_Kquota                  = "kQuota"
	GetManagementAlertsOptions_AlertCategoryList_Kremotereplication      = "kRemoteReplication"
	GetManagementAlertsOptions_AlertCategoryList_Ksecurity               = "kSecurity"
	GetManagementAlertsOptions_AlertCategoryList_Kstoragedevice          = "kStorageDevice"
	GetManagementAlertsOptions_AlertCategoryList_Kstoragepool            = "kStoragePool"
	GetManagementAlertsOptions_AlertCategoryList_Kstorageusage           = "kStorageUsage"
	GetManagementAlertsOptions_AlertCategoryList_Ksystemservice          = "kSystemService"
	GetManagementAlertsOptions_AlertCategoryList_Ktemperature            = "kTemperature"
	GetManagementAlertsOptions_AlertCategoryList_Kupgrade                = "kUpgrade"
	GetManagementAlertsOptions_AlertCategoryList_Kviewfailover           = "kViewFailover"
)

// Constants associated with the GetManagementAlertsOptions.AlertTypeBucketList property.
const (
	GetManagementAlertsOptions_AlertTypeBucketList_Kdataservice = "kDataService"
	GetManagementAlertsOptions_AlertTypeBucketList_Khardware    = "kHardware"
	GetManagementAlertsOptions_AlertTypeBucketList_Kmaintenance = "kMaintenance"
	GetManagementAlertsOptions_AlertTypeBucketList_Ksoftware    = "kSoftware"
)

// NewGetManagementAlertsOptions : Instantiate GetManagementAlertsOptions
func (*BackupRecoveryManagementSreApiV1) NewGetManagementAlertsOptions() *GetManagementAlertsOptions {
	return &GetManagementAlertsOptions{}
}

// SetAlertIdList : Allow user to set AlertIdList
func (_options *GetManagementAlertsOptions) SetAlertIdList(alertIdList []string) *GetManagementAlertsOptions {
	_options.AlertIdList = alertIdList
	return _options
}

// SetAlertStateList : Allow user to set AlertStateList
func (_options *GetManagementAlertsOptions) SetAlertStateList(alertStateList []string) *GetManagementAlertsOptions {
	_options.AlertStateList = alertStateList
	return _options
}

// SetAlertTypeList : Allow user to set AlertTypeList
func (_options *GetManagementAlertsOptions) SetAlertTypeList(alertTypeList []int64) *GetManagementAlertsOptions {
	_options.AlertTypeList = alertTypeList
	return _options
}

// SetAlertSeverityList : Allow user to set AlertSeverityList
func (_options *GetManagementAlertsOptions) SetAlertSeverityList(alertSeverityList []string) *GetManagementAlertsOptions {
	_options.AlertSeverityList = alertSeverityList
	return _options
}

// SetRegionIds : Allow user to set RegionIds
func (_options *GetManagementAlertsOptions) SetRegionIds(regionIds []string) *GetManagementAlertsOptions {
	_options.RegionIds = regionIds
	return _options
}

// SetClusterIdentifiers : Allow user to set ClusterIdentifiers
func (_options *GetManagementAlertsOptions) SetClusterIdentifiers(clusterIdentifiers []string) *GetManagementAlertsOptions {
	_options.ClusterIdentifiers = clusterIdentifiers
	return _options
}

// SetStartDateUsecs : Allow user to set StartDateUsecs
func (_options *GetManagementAlertsOptions) SetStartDateUsecs(startDateUsecs int64) *GetManagementAlertsOptions {
	_options.StartDateUsecs = core.Int64Ptr(startDateUsecs)
	return _options
}

// SetEndDateUsecs : Allow user to set EndDateUsecs
func (_options *GetManagementAlertsOptions) SetEndDateUsecs(endDateUsecs int64) *GetManagementAlertsOptions {
	_options.EndDateUsecs = core.Int64Ptr(endDateUsecs)
	return _options
}

// SetMaxAlerts : Allow user to set MaxAlerts
func (_options *GetManagementAlertsOptions) SetMaxAlerts(maxAlerts int64) *GetManagementAlertsOptions {
	_options.MaxAlerts = core.Int64Ptr(maxAlerts)
	return _options
}

// SetAlertCategoryList : Allow user to set AlertCategoryList
func (_options *GetManagementAlertsOptions) SetAlertCategoryList(alertCategoryList []string) *GetManagementAlertsOptions {
	_options.AlertCategoryList = alertCategoryList
	return _options
}

// SetTenantIds : Allow user to set TenantIds
func (_options *GetManagementAlertsOptions) SetTenantIds(tenantIds []string) *GetManagementAlertsOptions {
	_options.TenantIds = tenantIds
	return _options
}

// SetAlertTypeBucketList : Allow user to set AlertTypeBucketList
func (_options *GetManagementAlertsOptions) SetAlertTypeBucketList(alertTypeBucketList []string) *GetManagementAlertsOptions {
	_options.AlertTypeBucketList = alertTypeBucketList
	return _options
}

// SetAlertPropertyKeyList : Allow user to set AlertPropertyKeyList
func (_options *GetManagementAlertsOptions) SetAlertPropertyKeyList(alertPropertyKeyList []string) *GetManagementAlertsOptions {
	_options.AlertPropertyKeyList = alertPropertyKeyList
	return _options
}

// SetAlertPropertyValueList : Allow user to set AlertPropertyValueList
func (_options *GetManagementAlertsOptions) SetAlertPropertyValueList(alertPropertyValueList []string) *GetManagementAlertsOptions {
	_options.AlertPropertyValueList = alertPropertyValueList
	return _options
}

// SetAlertName : Allow user to set AlertName
func (_options *GetManagementAlertsOptions) SetAlertName(alertName string) *GetManagementAlertsOptions {
	_options.AlertName = core.StringPtr(alertName)
	return _options
}

// SetServiceInstanceIds : Allow user to set ServiceInstanceIds
func (_options *GetManagementAlertsOptions) SetServiceInstanceIds(serviceInstanceIds []string) *GetManagementAlertsOptions {
	_options.ServiceInstanceIds = serviceInstanceIds
	return _options
}

// SetHeaders : Allow user to set Headers
func (options *GetManagementAlertsOptions) SetHeaders(param map[string]string) *GetManagementAlertsOptions {
	options.Headers = param
	return options
}

// GetManagementAlertsStatsOptions : The GetManagementAlertsStats options.
type GetManagementAlertsStatsOptions struct {
	// Specifies the start time Unix time epoch in microseconds from which the active alerts stats are computed.
	StartTimeUsecs *int64 `json:"startTimeUsecs" validate:"required"`

	// Specifies the end time Unix time epoch in microseconds to which the active alerts stats are computed.
	EndTimeUsecs *int64 `json:"endTimeUsecs" validate:"required"`

	// Specifies the list of cluster IDs.
	ClusterIds []int64 `json:"clusterIds,omitempty"`

	// Specifies list of service instance ids to filter alert stats by.
	ServiceInstanceIds []string `json:"serviceInstanceIds,omitempty"`

	// Filter by a list of region ids.
	RegionIds []string `json:"regionIds,omitempty"`

	// Specifies if stats of active alerts per cluster needs to be excluded. If set to false (default value), stats of
	// active alerts per cluster is included in the response. If set to true, only aggregated stats summary will be present
	// in the response.
	ExcludeStatsByCluster *bool `json:"excludeStatsByCluster,omitempty"`

	// Specifies a list of alert origination source. If not specified, all alerts from all the sources are considered in
	// the response.
	AlertSource *string `json:"alertSource,omitempty"`

	// Specifies a list of tenants.
	TenantIds []string `json:"tenantIds,omitempty"`

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// Constants associated with the GetManagementAlertsStatsOptions.AlertSource property.
// Specifies a list of alert origination source. If not specified, all alerts from all the sources are considered in the
// response.
const (
	GetManagementAlertsStatsOptions_AlertSource_Kcluster = "kCluster"
	GetManagementAlertsStatsOptions_AlertSource_Khelios  = "kHelios"
)

// NewGetManagementAlertsStatsOptions : Instantiate GetManagementAlertsStatsOptions
func (*BackupRecoveryManagementSreApiV1) NewGetManagementAlertsStatsOptions(startTimeUsecs int64, endTimeUsecs int64) *GetManagementAlertsStatsOptions {
	return &GetManagementAlertsStatsOptions{
		StartTimeUsecs: core.Int64Ptr(startTimeUsecs),
		EndTimeUsecs:   core.Int64Ptr(endTimeUsecs),
	}
}

// SetStartTimeUsecs : Allow user to set StartTimeUsecs
func (_options *GetManagementAlertsStatsOptions) SetStartTimeUsecs(startTimeUsecs int64) *GetManagementAlertsStatsOptions {
	_options.StartTimeUsecs = core.Int64Ptr(startTimeUsecs)
	return _options
}

// SetEndTimeUsecs : Allow user to set EndTimeUsecs
func (_options *GetManagementAlertsStatsOptions) SetEndTimeUsecs(endTimeUsecs int64) *GetManagementAlertsStatsOptions {
	_options.EndTimeUsecs = core.Int64Ptr(endTimeUsecs)
	return _options
}

// SetClusterIds : Allow user to set ClusterIds
func (_options *GetManagementAlertsStatsOptions) SetClusterIds(clusterIds []int64) *GetManagementAlertsStatsOptions {
	_options.ClusterIds = clusterIds
	return _options
}

// SetServiceInstanceIds : Allow user to set ServiceInstanceIds
func (_options *GetManagementAlertsStatsOptions) SetServiceInstanceIds(serviceInstanceIds []string) *GetManagementAlertsStatsOptions {
	_options.ServiceInstanceIds = serviceInstanceIds
	return _options
}

// SetRegionIds : Allow user to set RegionIds
func (_options *GetManagementAlertsStatsOptions) SetRegionIds(regionIds []string) *GetManagementAlertsStatsOptions {
	_options.RegionIds = regionIds
	return _options
}

// SetExcludeStatsByCluster : Allow user to set ExcludeStatsByCluster
func (_options *GetManagementAlertsStatsOptions) SetExcludeStatsByCluster(excludeStatsByCluster bool) *GetManagementAlertsStatsOptions {
	_options.ExcludeStatsByCluster = core.BoolPtr(excludeStatsByCluster)
	return _options
}

// SetAlertSource : Allow user to set AlertSource
func (_options *GetManagementAlertsStatsOptions) SetAlertSource(alertSource string) *GetManagementAlertsStatsOptions {
	_options.AlertSource = core.StringPtr(alertSource)
	return _options
}

// SetTenantIds : Allow user to set TenantIds
func (_options *GetManagementAlertsStatsOptions) SetTenantIds(tenantIds []string) *GetManagementAlertsStatsOptions {
	_options.TenantIds = tenantIds
	return _options
}

// SetHeaders : Allow user to set Headers
func (options *GetManagementAlertsStatsOptions) SetHeaders(param map[string]string) *GetManagementAlertsStatsOptions {
	options.Headers = param
	return options
}

// GetManagementAlertsSummaryOptions : The GetManagementAlertsSummary options.
type GetManagementAlertsSummaryOptions struct {
	// Specifies the list of cluster identifiers. Format is clusterId:clusterIncarnationId.
	ClusterIdentifiers []string `json:"clusterIdentifiers,omitempty"`

	// Filter by start time. Specify the start time as a Unix epoch Timestamp (in microseconds). By default it is current
	// time minus a day.
	StartTimeUsecs *int64 `json:"startTimeUsecs,omitempty"`

	// Filter by end time. Specify the end time as a Unix epoch Timestamp (in microseconds). By default it is current time.
	EndTimeUsecs *int64 `json:"endTimeUsecs,omitempty"`

	// Specifies list of alert states to filter alerts by. If not specified, only open alerts will be used to get summary.
	StatesList []string `json:"statesList,omitempty"`

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// Constants associated with the GetManagementAlertsSummaryOptions.StatesList property.
const (
	GetManagementAlertsSummaryOptions_StatesList_Knote       = "kNote"
	GetManagementAlertsSummaryOptions_StatesList_Kopen       = "kOpen"
	GetManagementAlertsSummaryOptions_StatesList_Kresolved   = "kResolved"
	GetManagementAlertsSummaryOptions_StatesList_Ksuppressed = "kSuppressed"
)

// NewGetManagementAlertsSummaryOptions : Instantiate GetManagementAlertsSummaryOptions
func (*BackupRecoveryManagementSreApiV1) NewGetManagementAlertsSummaryOptions() *GetManagementAlertsSummaryOptions {
	return &GetManagementAlertsSummaryOptions{}
}

// SetClusterIdentifiers : Allow user to set ClusterIdentifiers
func (_options *GetManagementAlertsSummaryOptions) SetClusterIdentifiers(clusterIdentifiers []string) *GetManagementAlertsSummaryOptions {
	_options.ClusterIdentifiers = clusterIdentifiers
	return _options
}

// SetStartTimeUsecs : Allow user to set StartTimeUsecs
func (_options *GetManagementAlertsSummaryOptions) SetStartTimeUsecs(startTimeUsecs int64) *GetManagementAlertsSummaryOptions {
	_options.StartTimeUsecs = core.Int64Ptr(startTimeUsecs)
	return _options
}

// SetEndTimeUsecs : Allow user to set EndTimeUsecs
func (_options *GetManagementAlertsSummaryOptions) SetEndTimeUsecs(endTimeUsecs int64) *GetManagementAlertsSummaryOptions {
	_options.EndTimeUsecs = core.Int64Ptr(endTimeUsecs)
	return _options
}

// SetStatesList : Allow user to set StatesList
func (_options *GetManagementAlertsSummaryOptions) SetStatesList(statesList []string) *GetManagementAlertsSummaryOptions {
	_options.StatesList = statesList
	return _options
}

// SetHeaders : Allow user to set Headers
func (options *GetManagementAlertsSummaryOptions) SetHeaders(param map[string]string) *GetManagementAlertsSummaryOptions {
	options.Headers = param
	return options
}

// Label : Label struct
type Label struct {
	// Key of the Label.
	Key *string `json:"key" validate:"required"`

	// Value of the Label, multiple values should be joined by '|'.
	Value *string `json:"value" validate:"required"`
}

// UnmarshalLabel unmarshals an instance of Label from the specified map of raw messages.
func UnmarshalLabel(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(Label)
	err = core.UnmarshalPrimitive(m, "key", &obj.Key)
	if err != nil {
		err = core.SDKErrorf(err, "", "key-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "value", &obj.Value)
	if err != nil {
		err = core.SDKErrorf(err, "", "value-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// McmActiveAlertsStats : Specifies the active alerts stats response.
type McmActiveAlertsStats struct {
	// Specifies the active alert statistics details.
	AggregatedAlertsStats *ActiveAlertsStats `json:"aggregatedAlertsStats,omitempty"`

	// Specifies the cluster statistics based on active alerts.
	AggregatedClusterStats *ClusterAlertStats `json:"aggregatedClusterStats,omitempty"`

	// Specifies the active Alerts stats by clusters.
	StatsByCluster []McmActiveAlertsStatsByCluster `json:"statsByCluster,omitempty"`
}

// UnmarshalMcmActiveAlertsStats unmarshals an instance of McmActiveAlertsStats from the specified map of raw messages.
func UnmarshalMcmActiveAlertsStats(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(McmActiveAlertsStats)
	err = core.UnmarshalModel(m, "aggregatedAlertsStats", &obj.AggregatedAlertsStats, UnmarshalActiveAlertsStats)
	if err != nil {
		err = core.SDKErrorf(err, "", "aggregatedAlertsStats-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "aggregatedClusterStats", &obj.AggregatedClusterStats, UnmarshalClusterAlertStats)
	if err != nil {
		err = core.SDKErrorf(err, "", "aggregatedClusterStats-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "statsByCluster", &obj.StatsByCluster, UnmarshalMcmActiveAlertsStatsByCluster)
	if err != nil {
		err = core.SDKErrorf(err, "", "statsByCluster-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// McmActiveAlertsStatsByCluster : Specifies the active alerts stats by cluster.
type McmActiveAlertsStatsByCluster struct {
	// Specifies the active alert statistics details.
	AlertsStats *ActiveAlertsStats `json:"alertsStats,omitempty"`

	// Specifies the Cluster Id.
	ClusterID *int64 `json:"clusterId,omitempty"`

	// Specifies the region id of cluster.
	RegionID *string `json:"regionId,omitempty"`
}

// UnmarshalMcmActiveAlertsStatsByCluster unmarshals an instance of McmActiveAlertsStatsByCluster from the specified map of raw messages.
func UnmarshalMcmActiveAlertsStatsByCluster(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(McmActiveAlertsStatsByCluster)
	err = core.UnmarshalModel(m, "alertsStats", &obj.AlertsStats, UnmarshalActiveAlertsStats)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertsStats-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "clusterId", &obj.ClusterID)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "regionId", &obj.RegionID)
	if err != nil {
		err = core.SDKErrorf(err, "", "regionId-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// NodeUpgradeLog : Specifies the upgrade logs for a node.
type NodeUpgradeLog struct {
	// Upgrade logs for the node.
	Logs []UpgradeLog `json:"logs,omitempty"`

	// Id of the node.
	NodeID *string `json:"nodeId,omitempty"`
}

// UnmarshalNodeUpgradeLog unmarshals an instance of NodeUpgradeLog from the specified map of raw messages.
func UnmarshalNodeUpgradeLog(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(NodeUpgradeLog)
	err = core.UnmarshalModel(m, "logs", &obj.Logs, UnmarshalUpgradeLog)
	if err != nil {
		err = core.SDKErrorf(err, "", "logs-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "nodeId", &obj.NodeID)
	if err != nil {
		err = core.SDKErrorf(err, "", "nodeId-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// PatchUpgradeParams : Specifies the parameters for patch upgrade request.
type PatchUpgradeParams struct {
	// Specifies the optional headers for the patch cluster request.
	AuthHeaders []AuthHeaderForClusterUpgrade `json:"authHeaders,omitempty"`

	// Specify if pre check results can be ignored.
	IgnorePreChecksFailure *bool `json:"ignorePreChecksFailure,omitempty"`

	// Specifies URL from which patch package can be downloaded. Note: This option is only supported in Multi-Cluster
	// Manager (MCM).
	PackageURL *string `json:"packageUrl,omitempty"`

	// Specifies target patch version to which clusters are to be upgraded.
	TargetVersion *string `json:"targetVersion,omitempty"`
}

// UnmarshalPatchUpgradeParams unmarshals an instance of PatchUpgradeParams from the specified map of raw messages.
func UnmarshalPatchUpgradeParams(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(PatchUpgradeParams)
	err = core.UnmarshalModel(m, "authHeaders", &obj.AuthHeaders, UnmarshalAuthHeaderForClusterUpgrade)
	if err != nil {
		err = core.SDKErrorf(err, "", "authHeaders-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "ignorePreChecksFailure", &obj.IgnorePreChecksFailure)
	if err != nil {
		err = core.SDKErrorf(err, "", "ignorePreChecksFailure-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "packageUrl", &obj.PackageURL)
	if err != nil {
		err = core.SDKErrorf(err, "", "packageUrl-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "targetVersion", &obj.TargetVersion)
	if err != nil {
		err = core.SDKErrorf(err, "", "targetVersion-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// ResolvedAlertInfo : The infomation of the alert being resolved.
type ResolvedAlertInfo struct {
	// Id of the alert.
	AlertID *int64 `json:"alertId,omitempty"`

	// Alert Id with string format.
	AlertIdStr *string `json:"alertIdStr,omitempty"`

	// Name of the alert being resolved.
	AlertName *string `json:"alertName,omitempty"`

	// Id of the cluster which the alert is associated.
	ClusterID *int64 `json:"clusterId,omitempty"`

	// First occurrence of the alert.
	FirstTimestampUsecs *int64 `json:"firstTimestampUsecs,omitempty"`

	ResolvedTimeUsec *int64 `json:"resolvedTimeUsec,omitempty"`

	// Id of the service instance which the alert is associated.
	ServiceInstanceID *string `json:"serviceInstanceId,omitempty"`
}

// UnmarshalResolvedAlertInfo unmarshals an instance of ResolvedAlertInfo from the specified map of raw messages.
func UnmarshalResolvedAlertInfo(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(ResolvedAlertInfo)
	err = core.UnmarshalPrimitive(m, "alertId", &obj.AlertID)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "alertIdStr", &obj.AlertIdStr)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertIdStr-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "alertName", &obj.AlertName)
	if err != nil {
		err = core.SDKErrorf(err, "", "alertName-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "clusterId", &obj.ClusterID)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "firstTimestampUsecs", &obj.FirstTimestampUsecs)
	if err != nil {
		err = core.SDKErrorf(err, "", "firstTimestampUsecs-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "resolvedTimeUsec", &obj.ResolvedTimeUsec)
	if err != nil {
		err = core.SDKErrorf(err, "", "resolvedTimeUsec-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "serviceInstanceId", &obj.ServiceInstanceID)
	if err != nil {
		err = core.SDKErrorf(err, "", "serviceInstanceId-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// SPClusterInfo : Specifies the details of a cluster claimed from IBM Storage Protect environment.
type SPClusterInfo struct {
	// Specifies cluster id.
	ClusterID *int64 `json:"clusterId,omitempty"`

	// Specifies cluster incarnation id.
	ClusterIncarnationID *int64 `json:"clusterIncarnationId,omitempty"`

	// Specifies cluster's name.
	ClusterName *string `json:"clusterName,omitempty"`

	// Specifies the currently running version on cluster.
	CurrentVersion *string `json:"currentVersion,omitempty"`

	// Specifies the health of the cluster.
	Health *string `json:"health,omitempty"`

	// Specifies if the cluster is connected to management console.
	IsConnectedToHelios *bool `json:"isConnectedToHelios,omitempty"`

	// Specifies an array of node ips for the cluster.
	NodeIps []string `json:"nodeIps,omitempty"`

	// Specifies the number of nodes in the cluster.
	NumberOfNodes *int64 `json:"numberOfNodes,omitempty"`

	// Specifies the type of the cluster provider.
	ProviderType *string `json:"providerType,omitempty"`

	// Specifies total capacity of the cluster in bytes.
	TotalCapacity *int64 `json:"totalCapacity,omitempty"`

	// Specifies the type of the SP cluster.
	Type *string `json:"type,omitempty"`

	// Specifies how much of the cluster capacity is consumed in bytes.
	UsedCapacity *int64 `json:"usedCapacity,omitempty"`
}

// Constants associated with the SPClusterInfo.Health property.
// Specifies the health of the cluster.
const (
	SPClusterInfo_Health_Critical    = "Critical"
	SPClusterInfo_Health_Noncritical = "NonCritical"
)

// Constants associated with the SPClusterInfo.ProviderType property.
// Specifies the type of the cluster provider.
const (
	SPClusterInfo_ProviderType_Kibmstorageprotect = "kIBMStorageProtect"
	SPClusterInfo_ProviderType_Kmcmcohesity       = "kMCMCohesity"
)

// UnmarshalSPClusterInfo unmarshals an instance of SPClusterInfo from the specified map of raw messages.
func UnmarshalSPClusterInfo(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(SPClusterInfo)
	err = core.UnmarshalPrimitive(m, "clusterId", &obj.ClusterID)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "clusterIncarnationId", &obj.ClusterIncarnationID)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterIncarnationId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "clusterName", &obj.ClusterName)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterName-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "currentVersion", &obj.CurrentVersion)
	if err != nil {
		err = core.SDKErrorf(err, "", "currentVersion-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "health", &obj.Health)
	if err != nil {
		err = core.SDKErrorf(err, "", "health-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "isConnectedToHelios", &obj.IsConnectedToHelios)
	if err != nil {
		err = core.SDKErrorf(err, "", "isConnectedToHelios-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "nodeIps", &obj.NodeIps)
	if err != nil {
		err = core.SDKErrorf(err, "", "nodeIps-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "numberOfNodes", &obj.NumberOfNodes)
	if err != nil {
		err = core.SDKErrorf(err, "", "numberOfNodes-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "providerType", &obj.ProviderType)
	if err != nil {
		err = core.SDKErrorf(err, "", "providerType-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "totalCapacity", &obj.TotalCapacity)
	if err != nil {
		err = core.SDKErrorf(err, "", "totalCapacity-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "type", &obj.Type)
	if err != nil {
		err = core.SDKErrorf(err, "", "type-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "usedCapacity", &obj.UsedCapacity)
	if err != nil {
		err = core.SDKErrorf(err, "", "usedCapacity-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// UpdateClustersUpgradesOptions : The UpdateClustersUpgrades options.
type UpdateClustersUpgradesOptions struct {
	// Specifies the optional headers for upgrade request.
	AuthHeaders []AuthHeaderForClusterUpgrade `json:"authHeaders,omitempty"`

	// Array for clusters to be upgraded.
	Clusters []Upgrade `json:"clusters,omitempty"`

	// Specifies the difference of time between two cluster's upgrade.
	IntervalForRollingUpgradeInHours *int64 `json:"intervalForRollingUpgradeInHours,omitempty"`

	// Specifies URL from which package can be downloaded. Note: This option is only supported in Multi-Cluster Manager
	// (MCM).
	PackageURL *string `json:"packageUrl,omitempty"`

	// Specifies the parameters for patch upgrade request.
	PatchUpgradeParams *PatchUpgradeParams `json:"patchUpgradeParams,omitempty"`

	// Specifies target version to which clusters are to be upgraded.
	TargetVersion *string `json:"targetVersion,omitempty"`

	// Specifies the time in msecs at which the cluster has to be upgraded.
	TimeStampToUpgradeAtMsecs *int64 `json:"timeStampToUpgradeAtMsecs,omitempty"`

	// Specifies the type of upgrade to be performed on a cluster.
	Type *string `json:"type,omitempty"`

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// Constants associated with the UpdateClustersUpgradesOptions.Type property.
// Specifies the type of upgrade to be performed on a cluster.
const (
	UpdateClustersUpgradesOptions_Type_Patch        = "Patch"
	UpdateClustersUpgradesOptions_Type_Upgrade      = "Upgrade"
	UpdateClustersUpgradesOptions_Type_Upgradepatch = "UpgradePatch"
)

// NewUpdateClustersUpgradesOptions : Instantiate UpdateClustersUpgradesOptions
func (*BackupRecoveryManagementSreApiV1) NewUpdateClustersUpgradesOptions() *UpdateClustersUpgradesOptions {
	return &UpdateClustersUpgradesOptions{}
}

// SetAuthHeaders : Allow user to set AuthHeaders
func (_options *UpdateClustersUpgradesOptions) SetAuthHeaders(authHeaders []AuthHeaderForClusterUpgrade) *UpdateClustersUpgradesOptions {
	_options.AuthHeaders = authHeaders
	return _options
}

// SetClusters : Allow user to set Clusters
func (_options *UpdateClustersUpgradesOptions) SetClusters(clusters []Upgrade) *UpdateClustersUpgradesOptions {
	_options.Clusters = clusters
	return _options
}

// SetIntervalForRollingUpgradeInHours : Allow user to set IntervalForRollingUpgradeInHours
func (_options *UpdateClustersUpgradesOptions) SetIntervalForRollingUpgradeInHours(intervalForRollingUpgradeInHours int64) *UpdateClustersUpgradesOptions {
	_options.IntervalForRollingUpgradeInHours = core.Int64Ptr(intervalForRollingUpgradeInHours)
	return _options
}

// SetPackageURL : Allow user to set PackageURL
func (_options *UpdateClustersUpgradesOptions) SetPackageURL(packageURL string) *UpdateClustersUpgradesOptions {
	_options.PackageURL = core.StringPtr(packageURL)
	return _options
}

// SetPatchUpgradeParams : Allow user to set PatchUpgradeParams
func (_options *UpdateClustersUpgradesOptions) SetPatchUpgradeParams(patchUpgradeParams *PatchUpgradeParams) *UpdateClustersUpgradesOptions {
	_options.PatchUpgradeParams = patchUpgradeParams
	return _options
}

// SetTargetVersion : Allow user to set TargetVersion
func (_options *UpdateClustersUpgradesOptions) SetTargetVersion(targetVersion string) *UpdateClustersUpgradesOptions {
	_options.TargetVersion = core.StringPtr(targetVersion)
	return _options
}

// SetTimeStampToUpgradeAtMsecs : Allow user to set TimeStampToUpgradeAtMsecs
func (_options *UpdateClustersUpgradesOptions) SetTimeStampToUpgradeAtMsecs(timeStampToUpgradeAtMsecs int64) *UpdateClustersUpgradesOptions {
	_options.TimeStampToUpgradeAtMsecs = core.Int64Ptr(timeStampToUpgradeAtMsecs)
	return _options
}

// SetType : Allow user to set Type
func (_options *UpdateClustersUpgradesOptions) SetType(typeVar string) *UpdateClustersUpgradesOptions {
	_options.Type = core.StringPtr(typeVar)
	return _options
}

// SetHeaders : Allow user to set Headers
func (options *UpdateClustersUpgradesOptions) SetHeaders(param map[string]string) *UpdateClustersUpgradesOptions {
	options.Headers = param
	return options
}

// Upgrade : Specifies list of cluster upgrades.
type Upgrade struct {
	// Specifies cluster id.
	ClusterID *int64 `json:"clusterId,omitempty"`

	// Specifies cluster incarnation id.
	ClusterIncarnationID *int64 `json:"clusterIncarnationId,omitempty"`

	// Specifies current version of cluster.
	CurrentVersion *string `json:"currentVersion,omitempty"`
}

// UnmarshalUpgrade unmarshals an instance of Upgrade from the specified map of raw messages.
func UnmarshalUpgrade(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(Upgrade)
	err = core.UnmarshalPrimitive(m, "clusterId", &obj.ClusterID)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "clusterIncarnationId", &obj.ClusterIncarnationID)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterIncarnationId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "currentVersion", &obj.CurrentVersion)
	if err != nil {
		err = core.SDKErrorf(err, "", "currentVersion-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// UpgradeCancelResponse : Specifies a cluster scheduled upgrade cancel response.
type UpgradeCancelResponse struct {
	// Specifies an error message if failed to cancel a scheduled upgrade.
	ErrorMessage *string `json:"ErrorMessage,omitempty"`

	// Specifies if scheduled upgrade cancelling was successful.
	IsUpgradeCancelSuccessful *bool `json:"IsUpgradeCancelSuccessful,omitempty"`

	// Specifies cluster id.
	ClusterID *int64 `json:"clusterId,omitempty"`

	// Specifies cluster incarnation id.
	ClusterIncarnationID *int64 `json:"clusterIncarnationId,omitempty"`
}

// UnmarshalUpgradeCancelResponse unmarshals an instance of UpgradeCancelResponse from the specified map of raw messages.
func UnmarshalUpgradeCancelResponse(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(UpgradeCancelResponse)
	err = core.UnmarshalPrimitive(m, "ErrorMessage", &obj.ErrorMessage)
	if err != nil {
		err = core.SDKErrorf(err, "", "ErrorMessage-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "IsUpgradeCancelSuccessful", &obj.IsUpgradeCancelSuccessful)
	if err != nil {
		err = core.SDKErrorf(err, "", "IsUpgradeCancelSuccessful-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "clusterId", &obj.ClusterID)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "clusterIncarnationId", &obj.ClusterIncarnationID)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterIncarnationId-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// UpgradeInfo : Upgrade progress and upgrade status of cluster. It returns the percentage complete.
type UpgradeInfo struct {
	// Specifies cluster's id.
	ClusterID *int64 `json:"clusterId,omitempty"`

	// Specifies cluster's incarnation id.
	ClusterIncarnationID *int64 `json:"clusterIncarnationId,omitempty"`

	// Patch software version against which these logs are generated. This is specified for Patch type only.
	PatchSoftwareVersion *string `json:"patchSoftwareVersion,omitempty"`

	// Upgrade software version against which these logs are generated.
	SoftwareVersion *string `json:"softwareVersion,omitempty"`

	// Specifies the type of upgrade on a cluster.
	Type *string `json:"type,omitempty"`

	// Upgrade logs per node.
	UpgradeLogs []NodeUpgradeLog `json:"upgradeLogs,omitempty"`

	// Upgrade percentage complete so far.
	UpgradePercentComplete *float64 `json:"upgradePercentComplete,omitempty"`

	// Upgrade status.
	UpgradeStatus *string `json:"upgradeStatus,omitempty"`
}

// Constants associated with the UpgradeInfo.Type property.
// Specifies the type of upgrade on a cluster.
const (
	UpgradeInfo_Type_Patch        = "Patch"
	UpgradeInfo_Type_Upgrade      = "Upgrade"
	UpgradeInfo_Type_Upgradepatch = "UpgradePatch"
)

// Constants associated with the UpgradeInfo.UpgradeStatus property.
// Upgrade status.
const (
	UpgradeInfo_UpgradeStatus_Clusterunreachable = "ClusterUnreachable"
	UpgradeInfo_UpgradeStatus_Complete           = "Complete"
	UpgradeInfo_UpgradeStatus_Failed             = "Failed"
	UpgradeInfo_UpgradeStatus_Inprogress         = "InProgress"
	UpgradeInfo_UpgradeStatus_Scheduled          = "Scheduled"
)

// UnmarshalUpgradeInfo unmarshals an instance of UpgradeInfo from the specified map of raw messages.
func UnmarshalUpgradeInfo(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(UpgradeInfo)
	err = core.UnmarshalPrimitive(m, "clusterId", &obj.ClusterID)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "clusterIncarnationId", &obj.ClusterIncarnationID)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterIncarnationId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "patchSoftwareVersion", &obj.PatchSoftwareVersion)
	if err != nil {
		err = core.SDKErrorf(err, "", "patchSoftwareVersion-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "softwareVersion", &obj.SoftwareVersion)
	if err != nil {
		err = core.SDKErrorf(err, "", "softwareVersion-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "type", &obj.Type)
	if err != nil {
		err = core.SDKErrorf(err, "", "type-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "upgradeLogs", &obj.UpgradeLogs, UnmarshalNodeUpgradeLog)
	if err != nil {
		err = core.SDKErrorf(err, "", "upgradeLogs-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "upgradePercentComplete", &obj.UpgradePercentComplete)
	if err != nil {
		err = core.SDKErrorf(err, "", "upgradePercentComplete-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "upgradeStatus", &obj.UpgradeStatus)
	if err != nil {
		err = core.SDKErrorf(err, "", "upgradeStatus-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// UpgradeLog : Specifies the log of an upgrade.
type UpgradeLog struct {
	// One log statement of the complete logs.
	Log *string `json:"log,omitempty"`

	// Time at which this log got generated.
	TimeStamp *int64 `json:"timeStamp,omitempty"`
}

// UnmarshalUpgradeLog unmarshals an instance of UpgradeLog from the specified map of raw messages.
func UnmarshalUpgradeLog(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(UpgradeLog)
	err = core.UnmarshalPrimitive(m, "log", &obj.Log)
	if err != nil {
		err = core.SDKErrorf(err, "", "log-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "timeStamp", &obj.TimeStamp)
	if err != nil {
		err = core.SDKErrorf(err, "", "timeStamp-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// UpgradeResponse : Specifies a cluster upgrade created or updated response.
type UpgradeResponse struct {
	// Specifies error message if failed to schedule upgrade.
	ErrorMessage *string `json:"ErrorMessage,omitempty"`

	// Specifies if upgrade scheduling was successsful.
	IsUpgradeSchedulingSuccessful *bool `json:"IsUpgradeSchedulingSuccessful,omitempty"`

	// Specifies cluster id.
	ClusterID *int64 `json:"clusterId,omitempty"`

	// Specifies cluster incarnation id.
	ClusterIncarnationID *int64 `json:"clusterIncarnationId,omitempty"`
}

// UnmarshalUpgradeResponse unmarshals an instance of UpgradeResponse from the specified map of raw messages.
func UnmarshalUpgradeResponse(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(UpgradeResponse)
	err = core.UnmarshalPrimitive(m, "ErrorMessage", &obj.ErrorMessage)
	if err != nil {
		err = core.SDKErrorf(err, "", "ErrorMessage-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "IsUpgradeSchedulingSuccessful", &obj.IsUpgradeSchedulingSuccessful)
	if err != nil {
		err = core.SDKErrorf(err, "", "IsUpgradeSchedulingSuccessful-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "clusterId", &obj.ClusterID)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "clusterIncarnationId", &obj.ClusterIncarnationID)
	if err != nil {
		err = core.SDKErrorf(err, "", "clusterIncarnationId-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// Vault : Specifies the fields of vault.
type Vault struct {
	// Specifies Global vault id.
	GlobalVaultID *string `json:"globalVaultId,omitempty"`

	// Specifies id of region where vault resides.
	RegionID *string `json:"regionId,omitempty"`

	// Specifies name of region where vault resides.
	RegionName *string `json:"regionName,omitempty"`

	// Specifies name of vault.
	VaultName *string `json:"vaultName,omitempty"`
}

// UnmarshalVault unmarshals an instance of Vault from the specified map of raw messages.
func UnmarshalVault(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(Vault)
	err = core.UnmarshalPrimitive(m, "globalVaultId", &obj.GlobalVaultID)
	if err != nil {
		err = core.SDKErrorf(err, "", "globalVaultId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "regionId", &obj.RegionID)
	if err != nil {
		err = core.SDKErrorf(err, "", "regionId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "regionName", &obj.RegionName)
	if err != nil {
		err = core.SDKErrorf(err, "", "regionName-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "vaultName", &obj.VaultName)
	if err != nil {
		err = core.SDKErrorf(err, "", "vaultName-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}
