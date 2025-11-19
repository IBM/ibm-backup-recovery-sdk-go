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
// Package backuprecoveryv1 : Operations and models for the BackupRecoveryV1 service
package backuprecoveryv1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/IBM/go-sdk-core/v5/core"
	common "github.com/IBM/ibm-backup-recovery-sdk-go/common"
)

// BackupRecoveryManagementReportingApiV1 : Public APIs exposed by IBM Management Console Reporting service.
//
// # Getting Started
//
// This SDK provides operations for interfacing with the IBM Management Console Reporting service.
//
// ## Base URL
//
// # The base URL for making API calls is
//
// `https://<MANAGEMENT_CONSOLE_REPORTING_URL>/heliosreporting/api/v1`
//
// ## Authentication
//
// An apiKey is needed in order to authenticate the requests to the Reporting service.
//
// ---
// **NOTE**
//
// The apiKey has no expiration and is valid until deleted explicitly.
//
// ---
//
// API Version: v1
type BackupRecoveryManagementReportingApiV1 struct {
	Service *core.BaseService
}

// DefaultServiceURL is the default URL to make service requests to.
const DefaultManagementReportingServiceURL = "https://manager.backup-recovery.cloud.ibm.com/heliosreporting/api/v1"

// DefaultServiceName is the default key used to find external configuration information.
const DefaultManagementReportingServiceName = "backup_recovery_management_reporting_api"

// BackupRecoveryManagementReportingApiV1Options : Service options
type BackupRecoveryManagementReportingApiV1Options struct {
	ServiceName   string
	URL           string
	Authenticator core.Authenticator
}

// NewBackupRecoveryManagementReportingApiV1UsingExternalConfig : constructs an instance of BackupRecoveryManagementReportingApiV1 with passed in options and external configuration.
func NewBackupRecoveryManagementReportingApiV1UsingExternalConfig(options *BackupRecoveryManagementReportingApiV1Options) (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1, err error) {
	if options.ServiceName == "" {
		options.ServiceName = DefaultManagementReportingServiceName
	}

	if options.Authenticator == nil {
		options.Authenticator, err = core.GetAuthenticatorFromEnvironment(options.ServiceName)
		if err != nil {
			err = core.SDKErrorf(err, "", "env-auth-error", common.GetComponentInfo())
			return
		}
	}

	backupRecoveryManagementReportingApi, err = NewBackupRecoveryManagementReportingApiV1(options)
	err = core.RepurposeSDKProblem(err, "new-client-error")
	if err != nil {
		return
	}

	err = backupRecoveryManagementReportingApi.Service.ConfigureService(options.ServiceName)
	if err != nil {
		err = core.SDKErrorf(err, "", "client-config-error", common.GetComponentInfo())
		return
	}

	if options.URL != "" {
		err = backupRecoveryManagementReportingApi.Service.SetServiceURL(options.URL)
		err = core.RepurposeSDKProblem(err, "url-set-error")
	}
	return
}

// NewBackupRecoveryManagementReportingApiV1 : constructs an instance of BackupRecoveryManagementReportingApiV1 with passed in options.
func NewBackupRecoveryManagementReportingApiV1(options *BackupRecoveryManagementReportingApiV1Options) (service *BackupRecoveryManagementReportingApiV1, err error) {
	serviceOptions := &core.ServiceOptions{
		URL:           DefaultManagementReportingServiceURL,
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

	service = &BackupRecoveryManagementReportingApiV1{
		Service: baseService,
	}

	return
}

// GetServiceURLForRegion returns the service URL to be used for the specified region
func GetManagementReportingServiceURLForRegion(region string) (string, error) {
	return "", core.SDKErrorf(nil, "service does not support regional URLs", "no-regional-support", common.GetComponentInfo())
}

// Clone makes a copy of "backupRecoveryManagementReportingApi" suitable for processing requests.
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) Clone() *BackupRecoveryManagementReportingApiV1 {
	if core.IsNil(backupRecoveryManagementReportingApi) {
		return nil
	}
	clone := *backupRecoveryManagementReportingApi
	clone.Service = backupRecoveryManagementReportingApi.Service.Clone()
	return &clone
}

// SetServiceURL sets the service URL
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) SetServiceURL(url string) error {
	err := backupRecoveryManagementReportingApi.Service.SetServiceURL(url)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-set-error", common.GetComponentInfo())
	}
	return err
}

// GetServiceURL returns the service URL
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) GetServiceURL() string {
	return backupRecoveryManagementReportingApi.Service.GetServiceURL()
}

// SetDefaultHeaders sets HTTP headers to be sent in every request
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) SetDefaultHeaders(headers http.Header) {
	backupRecoveryManagementReportingApi.Service.SetDefaultHeaders(headers)
}

// SetEnableGzipCompression sets the service's EnableGzipCompression field
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) SetEnableGzipCompression(enableGzip bool) {
	backupRecoveryManagementReportingApi.Service.SetEnableGzipCompression(enableGzip)
}

// GetEnableGzipCompression returns the service's EnableGzipCompression field
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) GetEnableGzipCompression() bool {
	return backupRecoveryManagementReportingApi.Service.GetEnableGzipCompression()
}

// EnableRetries enables automatic retries for requests invoked for this service instance.
// If either parameter is specified as 0, then a default value is used instead.
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) EnableRetries(maxRetries int, maxRetryInterval time.Duration) {
	backupRecoveryManagementReportingApi.Service.EnableRetries(maxRetries, maxRetryInterval)
}

// DisableRetries disables automatic retries for requests invoked for this service instance.
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) DisableRetries() {
	backupRecoveryManagementReportingApi.Service.DisableRetries()
}

// GetComponents : List Report Components
// Fetches list of all report components accessible by logged in user.
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) GetComponents(getComponentsOptions *GetComponentsOptions) (result *Components, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementReportingApi.GetComponentsWithContext(context.Background(), getComponentsOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// GetComponentsWithContext is an alternate form of the GetComponents method which supports a Context parameter
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) GetComponentsWithContext(ctx context.Context, getComponentsOptions *GetComponentsOptions) (result *Components, response *core.DetailedResponse, err error) {
	err = core.ValidateStruct(getComponentsOptions, "getComponentsOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	builder := core.NewRequestBuilder(core.GET)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementReportingApi.GetEnableGzipCompression()
	_, err = builder.ResolveRequestURL(backupRecoveryManagementReportingApi.Service.Options.URL, `/public/components`, nil)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_reporting_api", "V1", "GetComponents")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range getComponentsOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")

	if getComponentsOptions.Ids != nil {
		builder.AddQuery("ids", strings.Join(getComponentsOptions.Ids, ","))
	}

	request, err := builder.Build()
	if err != nil {
		err = core.SDKErrorf(err, "", "build-error", common.GetComponentInfo())
		return
	}

	var rawResponse map[string]json.RawMessage
	response, err = backupRecoveryManagementReportingApi.Service.Request(request, &rawResponse)
	if err != nil {
		core.EnrichHTTPProblem(err, "GetComponents", getManagementServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}
	if rawResponse != nil {
		err = core.UnmarshalModel(rawResponse, "", &result, UnmarshalComponents)
		if err != nil {
			err = core.SDKErrorf(err, "", "unmarshal-resp-error", common.GetComponentInfo())
			return
		}
		response.Result = result
	}

	return
}

// GetComponentByID : Fetch a Report Component
// Get information about a Report component.
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) GetComponentByID(getComponentByIdOptions *GetComponentByIdOptions) (result *Component, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementReportingApi.GetComponentByIDWithContext(context.Background(), getComponentByIdOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// GetComponentByIDWithContext is an alternate form of the GetComponentByID method which supports a Context parameter
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) GetComponentByIDWithContext(ctx context.Context, getComponentByIdOptions *GetComponentByIdOptions) (result *Component, response *core.DetailedResponse, err error) {
	err = core.ValidateNotNil(getComponentByIdOptions, "getComponentByIdOptions cannot be nil")
	if err != nil {
		err = core.SDKErrorf(err, "", "unexpected-nil-param", common.GetComponentInfo())
		return
	}
	err = core.ValidateStruct(getComponentByIdOptions, "getComponentByIdOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	pathParamsMap := map[string]string{
		"id": *getComponentByIdOptions.ID,
	}

	builder := core.NewRequestBuilder(core.GET)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementReportingApi.GetEnableGzipCompression()
	_, err = builder.ResolveRequestURL(backupRecoveryManagementReportingApi.Service.Options.URL, `/public/components/{id}`, pathParamsMap)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_reporting_api", "V1", "GetComponentByID")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range getComponentByIdOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")

	request, err := builder.Build()
	if err != nil {
		err = core.SDKErrorf(err, "", "build-error", common.GetComponentInfo())
		return
	}

	var rawResponse map[string]json.RawMessage
	response, err = backupRecoveryManagementReportingApi.Service.Request(request, &rawResponse)
	if err != nil {
		core.EnrichHTTPProblem(err, "GetComponentById", getManagementServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}
	if rawResponse != nil {
		err = core.UnmarshalModel(rawResponse, "", &result, UnmarshalComponent)
		if err != nil {
			err = core.SDKErrorf(err, "", "unmarshal-resp-error", common.GetComponentInfo())
			return
		}
		response.Result = result
	}

	return
}

// GetComponentPreview : Fetch Preview of a component
// Get preview for a component specified by Id.
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) GetComponentPreview(getComponentPreviewOptions *GetComponentPreviewOptions) (result *ComponentPreview, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementReportingApi.GetComponentPreviewWithContext(context.Background(), getComponentPreviewOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// GetComponentPreviewWithContext is an alternate form of the GetComponentPreview method which supports a Context parameter
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) GetComponentPreviewWithContext(ctx context.Context, getComponentPreviewOptions *GetComponentPreviewOptions) (result *ComponentPreview, response *core.DetailedResponse, err error) {
	err = core.ValidateNotNil(getComponentPreviewOptions, "getComponentPreviewOptions cannot be nil")
	if err != nil {
		err = core.SDKErrorf(err, "", "unexpected-nil-param", common.GetComponentInfo())
		return
	}
	err = core.ValidateStruct(getComponentPreviewOptions, "getComponentPreviewOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	pathParamsMap := map[string]string{
		"id": *getComponentPreviewOptions.ID,
	}

	builder := core.NewRequestBuilder(core.POST)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementReportingApi.GetEnableGzipCompression()
	_, err = builder.ResolveRequestURL(backupRecoveryManagementReportingApi.Service.Options.URL, `/public/components/{id}/preview`, pathParamsMap)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_reporting_api", "V1", "GetComponentPreview")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range getComponentPreviewOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")
	builder.AddHeader("Content-Type", "application/json")

	body := make(map[string]interface{})
	if getComponentPreviewOptions.Filters != nil {
		body["filters"] = getComponentPreviewOptions.Filters
	}
	if getComponentPreviewOptions.Limit != nil {
		body["limit"] = getComponentPreviewOptions.Limit
	}
	if getComponentPreviewOptions.Sort != nil {
		body["sort"] = getComponentPreviewOptions.Sort
	}
	if getComponentPreviewOptions.Timezone != nil {
		body["timezone"] = getComponentPreviewOptions.Timezone
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

	var rawResponse map[string]json.RawMessage
	response, err = backupRecoveryManagementReportingApi.Service.Request(request, &rawResponse)
	if err != nil {
		core.EnrichHTTPProblem(err, "GetComponentPreview", getManagementServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}
	if rawResponse != nil {
		err = core.UnmarshalModel(rawResponse, "", &result, UnmarshalComponentPreview)
		if err != nil {
			err = core.SDKErrorf(err, "", "unmarshal-resp-error", common.GetComponentInfo())
			return
		}
		response.Result = result
	}

	return
}

// GetResources : Get resources
// Get different kinds of resources available which are discovered on Management Console Reporting service. These values can be used for filtering
// options.
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) GetResources(getResourcesOptions *GetResourcesOptions) (result *Resources, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementReportingApi.GetResourcesWithContext(context.Background(), getResourcesOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// GetResourcesWithContext is an alternate form of the GetResources method which supports a Context parameter
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) GetResourcesWithContext(ctx context.Context, getResourcesOptions *GetResourcesOptions) (result *Resources, response *core.DetailedResponse, err error) {
	err = core.ValidateNotNil(getResourcesOptions, "getResourcesOptions cannot be nil")
	if err != nil {
		err = core.SDKErrorf(err, "", "unexpected-nil-param", common.GetComponentInfo())
		return
	}
	err = core.ValidateStruct(getResourcesOptions, "getResourcesOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	builder := core.NewRequestBuilder(core.POST)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementReportingApi.GetEnableGzipCompression()
	_, err = builder.ResolveRequestURL(backupRecoveryManagementReportingApi.Service.Options.URL, `/public/resources`, nil)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_reporting_api", "V1", "GetResources")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range getResourcesOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")
	builder.AddHeader("Content-Type", "application/json")

	body := make(map[string]interface{})
	if getResourcesOptions.ResourceType != nil {
		body["resourceType"] = getResourcesOptions.ResourceType
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

	var rawResponse map[string]json.RawMessage
	response, err = backupRecoveryManagementReportingApi.Service.Request(request, &rawResponse)
	if err != nil {
		core.EnrichHTTPProblem(err, "GetResources", getManagementServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}
	if rawResponse != nil {
		err = core.UnmarshalModel(rawResponse, "", &result, UnmarshalResources)
		if err != nil {
			err = core.SDKErrorf(err, "", "unmarshal-resp-error", common.GetComponentInfo())
			return
		}
		response.Result = result
	}

	return
}

// GetProviderInstances : List Provider Instances
// List all the Provider Instances with details. The API returns the detailed listing of provider instances. The
// providers are responsible for managing the life cycle of instances and this API only intends to provide metadata
// information about the instances.
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) GetProviderInstances(getProviderInstancesOptions *GetProviderInstancesOptions) (result *ProviderInstancesList, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementReportingApi.GetProviderInstancesWithContext(context.Background(), getProviderInstancesOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// GetProviderInstancesWithContext is an alternate form of the GetProviderInstances method which supports a Context parameter
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) GetProviderInstancesWithContext(ctx context.Context, getProviderInstancesOptions *GetProviderInstancesOptions) (result *ProviderInstancesList, response *core.DetailedResponse, err error) {
	err = core.ValidateStruct(getProviderInstancesOptions, "getProviderInstancesOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	builder := core.NewRequestBuilder(core.GET)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementReportingApi.GetEnableGzipCompression()
	providerInstancesUrl := strings.TrimSuffix(backupRecoveryManagementReportingApi.Service.Options.URL, "/heliosreporting/api/v1")
	_, err = builder.ResolveRequestURL(providerInstancesUrl, `/v2/mcm/provider-instances`, nil)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_reporting_api", "V2", "GetProviderInstances")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range getProviderInstancesOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")

	if getProviderInstancesOptions.InstanceIds != nil {
		builder.AddQuery("instanceIds", strings.Join(getProviderInstancesOptions.InstanceIds, ","))
	}
	if getProviderInstancesOptions.Regions != nil {
		builder.AddQuery("regions", strings.Join(getProviderInstancesOptions.Regions, ","))
	}
	if getProviderInstancesOptions.IncludeServiceInstanceStatus != nil {
		builder.AddQuery("includeServiceInstanceStatus", fmt.Sprint(*getProviderInstancesOptions.IncludeServiceInstanceStatus))
	}

	request, err := builder.Build()
	if err != nil {
		err = core.SDKErrorf(err, "", "build-error", common.GetComponentInfo())
		return
	}

	var rawResponse map[string]json.RawMessage
	response, err = backupRecoveryManagementReportingApi.Service.Request(request, &rawResponse)
	if err != nil {
		core.EnrichHTTPProblem(err, "GetProviderInstances", getManagementServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}
	if rawResponse != nil {
		err = core.UnmarshalModel(rawResponse, "", &result, UnmarshalProviderInstancesList)
		if err != nil {
			err = core.SDKErrorf(err, "", "unmarshal-resp-error", common.GetComponentInfo())
			return
		}
		response.Result = result
	}

	return
}

// GetReportType : List properties of a report type
// Fetches list of properties of a report type.
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) GetReportType(getReportTypeOptions *GetReportTypeOptions) (result *ReportTypeAttributes, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementReportingApi.GetReportTypeWithContext(context.Background(), getReportTypeOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// GetReportTypeWithContext is an alternate form of the GetReportType method which supports a Context parameter
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) GetReportTypeWithContext(ctx context.Context, getReportTypeOptions *GetReportTypeOptions) (result *ReportTypeAttributes, response *core.DetailedResponse, err error) {
	err = core.ValidateNotNil(getReportTypeOptions, "getReportTypeOptions cannot be nil")
	if err != nil {
		err = core.SDKErrorf(err, "", "unexpected-nil-param", common.GetComponentInfo())
		return
	}
	err = core.ValidateStruct(getReportTypeOptions, "getReportTypeOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	pathParamsMap := map[string]string{
		"reportType": *getReportTypeOptions.ReportType,
	}

	builder := core.NewRequestBuilder(core.GET)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementReportingApi.GetEnableGzipCompression()
	_, err = builder.ResolveRequestURL(backupRecoveryManagementReportingApi.Service.Options.URL, `/public/reports/report-types/{reportType}`, pathParamsMap)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_reporting_api", "V1", "GetReportType")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range getReportTypeOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")

	request, err := builder.Build()
	if err != nil {
		err = core.SDKErrorf(err, "", "build-error", common.GetComponentInfo())
		return
	}

	var rawResponse map[string]json.RawMessage
	response, err = backupRecoveryManagementReportingApi.Service.Request(request, &rawResponse)
	if err != nil {
		core.EnrichHTTPProblem(err, "GetReportType", getManagementServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}
	if rawResponse != nil {
		err = core.UnmarshalModel(rawResponse, "", &result, UnmarshalReportTypeAttributes)
		if err != nil {
			err = core.SDKErrorf(err, "", "unmarshal-resp-error", common.GetComponentInfo())
			return
		}
		response.Result = result
	}

	return
}

// GetReports : List Reports
// Fetches list of all reports accessible by logged in user.
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) GetReports(getReportsOptions *GetReportsOptions) (result *Reports, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementReportingApi.GetReportsWithContext(context.Background(), getReportsOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// GetReportsWithContext is an alternate form of the GetReports method which supports a Context parameter
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) GetReportsWithContext(ctx context.Context, getReportsOptions *GetReportsOptions) (result *Reports, response *core.DetailedResponse, err error) {
	err = core.ValidateStruct(getReportsOptions, "getReportsOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	builder := core.NewRequestBuilder(core.GET)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementReportingApi.GetEnableGzipCompression()
	_, err = builder.ResolveRequestURL(backupRecoveryManagementReportingApi.Service.Options.URL, `/public/reports`, nil)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_reporting_api", "V1", "GetReports")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range getReportsOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")

	if getReportsOptions.Ids != nil {
		builder.AddQuery("ids", strings.Join(getReportsOptions.Ids, ","))
	}
	if getReportsOptions.UserContext != nil {
		builder.AddQuery("userContext", fmt.Sprint(*getReportsOptions.UserContext))
	}

	request, err := builder.Build()
	if err != nil {
		err = core.SDKErrorf(err, "", "build-error", common.GetComponentInfo())
		return
	}

	var rawResponse map[string]json.RawMessage
	response, err = backupRecoveryManagementReportingApi.Service.Request(request, &rawResponse)
	if err != nil {
		core.EnrichHTTPProblem(err, "GetReports", getManagementServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}
	if rawResponse != nil {
		err = core.UnmarshalModel(rawResponse, "", &result, UnmarshalReports)
		if err != nil {
			err = core.SDKErrorf(err, "", "unmarshal-resp-error", common.GetComponentInfo())
			return
		}
		response.Result = result
	}

	return
}

// GetReportByID : Fetch a Report
// Get a report for a given id.
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) GetReportByID(getReportByIdOptions *GetReportByIdOptions) (result *Report, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementReportingApi.GetReportByIDWithContext(context.Background(), getReportByIdOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// GetReportByIDWithContext is an alternate form of the GetReportByID method which supports a Context parameter
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) GetReportByIDWithContext(ctx context.Context, getReportByIdOptions *GetReportByIdOptions) (result *Report, response *core.DetailedResponse, err error) {
	err = core.ValidateNotNil(getReportByIdOptions, "getReportByIdOptions cannot be nil")
	if err != nil {
		err = core.SDKErrorf(err, "", "unexpected-nil-param", common.GetComponentInfo())
		return
	}
	err = core.ValidateStruct(getReportByIdOptions, "getReportByIdOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	pathParamsMap := map[string]string{
		"id": *getReportByIdOptions.ID,
	}

	builder := core.NewRequestBuilder(core.GET)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementReportingApi.GetEnableGzipCompression()
	_, err = builder.ResolveRequestURL(backupRecoveryManagementReportingApi.Service.Options.URL, `/public/reports/{id}`, pathParamsMap)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_reporting_api", "V1", "GetReportByID")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range getReportByIdOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")

	request, err := builder.Build()
	if err != nil {
		err = core.SDKErrorf(err, "", "build-error", common.GetComponentInfo())
		return
	}

	var rawResponse map[string]json.RawMessage
	response, err = backupRecoveryManagementReportingApi.Service.Request(request, &rawResponse)
	if err != nil {
		core.EnrichHTTPProblem(err, "GetReportById", getManagementServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}
	if rawResponse != nil {
		err = core.UnmarshalModel(rawResponse, "", &result, UnmarshalReport)
		if err != nil {
			err = core.SDKErrorf(err, "", "unmarshal-resp-error", common.GetComponentInfo())
			return
		}
		response.Result = result
	}

	return
}

// GetReportPreview : Fetch a Report Preview
// Get preview of a configured report.
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) GetReportPreview(getReportPreviewOptions *GetReportPreviewOptions) (result *ReportPreview, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementReportingApi.GetReportPreviewWithContext(context.Background(), getReportPreviewOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// GetReportPreviewWithContext is an alternate form of the GetReportPreview method which supports a Context parameter
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) GetReportPreviewWithContext(ctx context.Context, getReportPreviewOptions *GetReportPreviewOptions) (result *ReportPreview, response *core.DetailedResponse, err error) {
	err = core.ValidateNotNil(getReportPreviewOptions, "getReportPreviewOptions cannot be nil")
	if err != nil {
		err = core.SDKErrorf(err, "", "unexpected-nil-param", common.GetComponentInfo())
		return
	}
	err = core.ValidateStruct(getReportPreviewOptions, "getReportPreviewOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	pathParamsMap := map[string]string{
		"id": *getReportPreviewOptions.ID,
	}

	builder := core.NewRequestBuilder(core.POST)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementReportingApi.GetEnableGzipCompression()
	_, err = builder.ResolveRequestURL(backupRecoveryManagementReportingApi.Service.Options.URL, `/public/reports/{id}/preview`, pathParamsMap)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_reporting_api", "V1", "GetReportPreview")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range getReportPreviewOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")
	builder.AddHeader("Content-Type", "application/json")

	body := make(map[string]interface{})
	if getReportPreviewOptions.ComponentIds != nil {
		body["componentIds"] = getReportPreviewOptions.ComponentIds
	}
	if getReportPreviewOptions.Filters != nil {
		body["filters"] = getReportPreviewOptions.Filters
	}
	if getReportPreviewOptions.Timezone != nil {
		body["timezone"] = getReportPreviewOptions.Timezone
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

	var rawResponse map[string]json.RawMessage
	response, err = backupRecoveryManagementReportingApi.Service.Request(request, &rawResponse)
	if err != nil {
		core.EnrichHTTPProblem(err, "GetReportPreview", getManagementServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}
	if rawResponse != nil {
		err = core.UnmarshalModel(rawResponse, "", &result, UnmarshalReportPreview)
		if err != nil {
			err = core.SDKErrorf(err, "", "unmarshal-resp-error", common.GetComponentInfo())
			return
		}
		response.Result = result
	}

	return
}

// ExportReport : Export a Report
// Export a configured report.
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) ExportReport(exportReportOptions *ExportReportOptions) (result io.ReadCloser, response *core.DetailedResponse, err error) {
	result, response, err = backupRecoveryManagementReportingApi.ExportReportWithContext(context.Background(), exportReportOptions)
	err = core.RepurposeSDKProblem(err, "")
	return
}

// ExportReportWithContext is an alternate form of the ExportReport method which supports a Context parameter
func (backupRecoveryManagementReportingApi *BackupRecoveryManagementReportingApiV1) ExportReportWithContext(ctx context.Context, exportReportOptions *ExportReportOptions) (result io.ReadCloser, response *core.DetailedResponse, err error) {
	err = core.ValidateNotNil(exportReportOptions, "exportReportOptions cannot be nil")
	if err != nil {
		err = core.SDKErrorf(err, "", "unexpected-nil-param", common.GetComponentInfo())
		return
	}
	err = core.ValidateStruct(exportReportOptions, "exportReportOptions")
	if err != nil {
		err = core.SDKErrorf(err, "", "struct-validation-error", common.GetComponentInfo())
		return
	}

	pathParamsMap := map[string]string{
		"id": *exportReportOptions.ID,
	}

	builder := core.NewRequestBuilder(core.POST)
	builder = builder.WithContext(ctx)
	builder.EnableGzipCompression = backupRecoveryManagementReportingApi.GetEnableGzipCompression()
	_, err = builder.ResolveRequestURL(backupRecoveryManagementReportingApi.Service.Options.URL, `/public/reports/{id}/export`, pathParamsMap)
	if err != nil {
		err = core.SDKErrorf(err, "", "url-resolve-error", common.GetComponentInfo())
		return
	}

	sdkHeaders := common.GetSdkHeaders("backup_recovery_management_reporting_api", "V1", "ExportReport")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}

	for headerName, headerValue := range exportReportOptions.Headers {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")
	builder.AddHeader("Content-Type", "application/json")

	body := make(map[string]interface{})
	if exportReportOptions.Async != nil {
		body["async"] = exportReportOptions.Async
	}
	if exportReportOptions.Filters != nil {
		body["filters"] = exportReportOptions.Filters
	}
	if exportReportOptions.Layout != nil {
		body["layout"] = exportReportOptions.Layout
	}
	if exportReportOptions.ReportFormat != nil {
		body["reportFormat"] = exportReportOptions.ReportFormat
	}
	if exportReportOptions.Timezone != nil {
		body["timezone"] = exportReportOptions.Timezone
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

	response, err = backupRecoveryManagementReportingApi.Service.Request(request, &result)
	if err != nil {
		core.EnrichHTTPProblem(err, "ExportReport", getManagementServiceComponentInfo())
		err = core.SDKErrorf(err, "", "http-request-err", common.GetComponentInfo())
		return
	}

	return
}
func getManagementServiceComponentInfo() *core.ProblemComponent {
	return core.NewProblemComponent(DefaultServiceName, "v1")
}

// AggregatedAttributesParams : Specifies list of aggregation properties to be applied on the attributes.
type AggregatedAttributesParams struct {
	// Specifies the aggregation type.
	AggregationType *string `json:"aggregationType" validate:"required"`

	// Specifies the attribute name.
	Attribute *string `json:"attribute" validate:"required"`

	// Specifies the label to be generated for this aggregated attribute. If not specified, then by default label of the
	// column in the output will be combination of aggregation type and attribute.
	Label *string `json:"label,omitempty"`
}

// Constants associated with the AggregatedAttributesParams.AggregationType property.
// Specifies the aggregation type.
const (
	AggregatedAttributesParams_AggregationType_Avg           = "avg"
	AggregatedAttributesParams_AggregationType_Count         = "count"
	AggregatedAttributesParams_AggregationType_Countdistinct = "countDistinct"
	AggregatedAttributesParams_AggregationType_Max           = "max"
	AggregatedAttributesParams_AggregationType_Min           = "min"
	AggregatedAttributesParams_AggregationType_Sum           = "sum"
)

// UnmarshalAggregatedAttributesParams unmarshals an instance of AggregatedAttributesParams from the specified map of raw messages.
func UnmarshalAggregatedAttributesParams(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(AggregatedAttributesParams)
	err = core.UnmarshalPrimitive(m, "aggregationType", &obj.AggregationType)
	if err != nil {
		err = core.SDKErrorf(err, "", "aggregationType-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "attribute", &obj.Attribute)
	if err != nil {
		err = core.SDKErrorf(err, "", "attribute-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "label", &obj.Label)
	if err != nil {
		err = core.SDKErrorf(err, "", "label-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// AttributeAggregations : Specifies the aggregation related information that needs to be applied on the attributes.
type AttributeAggregations struct {
	// Specifies list of aggregation properties to be applied on the attributes.
	AggregatedAttributes []AggregatedAttributesParams `json:"aggregatedAttributes,omitempty"`

	// Specifies list of attributes to be grouped in the aggregation.
	GroupedAttributes []string `json:"groupedAttributes,omitempty"`
}

// UnmarshalAttributeAggregations unmarshals an instance of AttributeAggregations from the specified map of raw messages.
func UnmarshalAttributeAggregations(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(AttributeAggregations)
	err = core.UnmarshalModel(m, "aggregatedAttributes", &obj.AggregatedAttributes, UnmarshalAggregatedAttributesParams)
	if err != nil {
		err = core.SDKErrorf(err, "", "aggregatedAttributes-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "groupedAttributes", &obj.GroupedAttributes)
	if err != nil {
		err = core.SDKErrorf(err, "", "groupedAttributes-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// AttributeFilter : Specifies the filters that are applied on attributes.
type AttributeFilter struct {
	// Specifies the attribute.
	Attribute *string `json:"attribute" validate:"required"`

	// Specifies the type of the filter that needs to be applied.
	FilterType *string `json:"filterType" validate:"required"`

	// Specifies the in filter that are applied on attributes.
	InFilterParams *InFilterParams `json:"inFilterParams,omitempty"`

	// Specifies the filters that are applied on attributes.
	RangeFilterParams *RangeFilterParams `json:"rangeFilterParams,omitempty"`

	// Specifies the systems filter. Specifying this will pre filter all the results provided list of system identifier
	// before applying aggregations.
	SystemsFilterParams *SystemsFilterParams `json:"systemsFilterParams,omitempty"`

	// Specifies the time range filter. Specifying this will pre filter all the results on necessary resources like
	// Protection Runs etc before applying aggregations. Currently, maximum allowed time range is 60 days.
	TimeRangeFilterParams *TimeRangeFilterParams `json:"timeRangeFilterParams,omitempty"`
}

// Constants associated with the AttributeFilter.FilterType property.
// Specifies the type of the filter that needs to be applied.
const (
	AttributeFilter_FilterType_In        = "In"
	AttributeFilter_FilterType_Range     = "Range"
	AttributeFilter_FilterType_Systems   = "Systems"
	AttributeFilter_FilterType_Tenants   = "Tenants"
	AttributeFilter_FilterType_Timerange = "TimeRange"
)

// NewAttributeFilter : Instantiate AttributeFilter (Generic Model Constructor)
func (*BackupRecoveryManagementReportingApiV1) NewAttributeFilter(attribute string, filterType string) (_model *AttributeFilter, err error) {
	_model = &AttributeFilter{
		Attribute:  core.StringPtr(attribute),
		FilterType: core.StringPtr(filterType),
	}
	err = core.ValidateStruct(_model, "required parameters")
	if err != nil {
		err = core.SDKErrorf(err, "", "model-missing-required", common.GetComponentInfo())
	}
	return
}

// UnmarshalAttributeFilter unmarshals an instance of AttributeFilter from the specified map of raw messages.
func UnmarshalAttributeFilter(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(AttributeFilter)
	err = core.UnmarshalPrimitive(m, "attribute", &obj.Attribute)
	if err != nil {
		err = core.SDKErrorf(err, "", "attribute-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "filterType", &obj.FilterType)
	if err != nil {
		err = core.SDKErrorf(err, "", "filterType-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "inFilterParams", &obj.InFilterParams, UnmarshalInFilterParams)
	if err != nil {
		err = core.SDKErrorf(err, "", "inFilterParams-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "rangeFilterParams", &obj.RangeFilterParams, UnmarshalRangeFilterParams)
	if err != nil {
		err = core.SDKErrorf(err, "", "rangeFilterParams-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "systemsFilterParams", &obj.SystemsFilterParams, UnmarshalSystemsFilterParams)
	if err != nil {
		err = core.SDKErrorf(err, "", "systemsFilterParams-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "timeRangeFilterParams", &obj.TimeRangeFilterParams, UnmarshalTimeRangeFilterParams)
	if err != nil {
		err = core.SDKErrorf(err, "", "timeRangeFilterParams-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// AttributeSort : Specifies the sorting (ordering) parameters to be applied to the resulting data.
type AttributeSort struct {
	// Specifies the name of the attribute.
	Attribute *string `json:"attribute" validate:"required"`

	// Specifies whether the sorting order should be descending. Default value is false.
	Desc *bool `json:"desc,omitempty"`
}

// NewAttributeSort : Instantiate AttributeSort (Generic Model Constructor)
func (*BackupRecoveryManagementReportingApiV1) NewAttributeSort(attribute string) (_model *AttributeSort, err error) {
	_model = &AttributeSort{
		Attribute: core.StringPtr(attribute),
	}
	err = core.ValidateStruct(_model, "required parameters")
	if err != nil {
		err = core.SDKErrorf(err, "", "model-missing-required", common.GetComponentInfo())
	}
	return
}

// UnmarshalAttributeSort unmarshals an instance of AttributeSort from the specified map of raw messages.
func UnmarshalAttributeSort(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(AttributeSort)
	err = core.UnmarshalPrimitive(m, "attribute", &obj.Attribute)
	if err != nil {
		err = core.SDKErrorf(err, "", "attribute-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "desc", &obj.Desc)
	if err != nil {
		err = core.SDKErrorf(err, "", "desc-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// Component : Specifies a Report Component.
type Component struct {
	// Specifies the aggregation related information that needs to be applied on the attributes.
	Aggs *AttributeAggregations `json:"aggs,omitempty"`

	// Specifies the configuration parameters to customize and format the columns in the report artifacts like excel, pdf
	// etc.
	Config *CustomConfigParams `json:"config,omitempty"`

	// Specifies the data returned after evaluating the component.
	Data []map[string]interface{} `json:"data,omitempty"`

	// Specifies description of the Component.
	Description *string `json:"description,omitempty"`

	// Specifies the filters that are applied on specific report type attributes in order to compose this component.
	Filters []AttributeFilter `json:"filters,omitempty"`

	// Specifies the id of the Component.
	ID *string `json:"id,omitempty"`

	// Specifies the parameters to limit the resulting dataset.
	Limit *LimitParams `json:"limit,omitempty"`

	// Specifies the name of the Component.
	Name *string `json:"name,omitempty"`

	// Specifies the report type on top of which this Component is created from.
	ReportType *string `json:"reportType,omitempty"`

	// Specifies the sorting (ordering) parameters to be applied to the resulting data.
	Sort []AttributeSort `json:"sort,omitempty"`
}

// Constants associated with the Component.ReportType property.
// Specifies the report type on top of which this Component is created from.
const (
	Component_ReportType_Failures                    = "Failures"
	Component_ReportType_Protectedobjects            = "ProtectedObjects"
	Component_ReportType_Protectedunprotectedobjects = "ProtectedUnprotectedObjects"
	Component_ReportType_Protectionactivity          = "ProtectionActivity"
	Component_ReportType_Protectiongroupsummary      = "ProtectionGroupSummary"
	Component_ReportType_Protectionruns              = "ProtectionRuns"
	Component_ReportType_Protectionrunstrend         = "ProtectionRunsTrend"
	Component_ReportType_Recovery                    = "Recovery"
)

// UnmarshalComponent unmarshals an instance of Component from the specified map of raw messages.
func UnmarshalComponent(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(Component)
	err = core.UnmarshalModel(m, "aggs", &obj.Aggs, UnmarshalAttributeAggregations)
	if err != nil {
		err = core.SDKErrorf(err, "", "aggs-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "config", &obj.Config, UnmarshalCustomConfigParams)
	if err != nil {
		err = core.SDKErrorf(err, "", "config-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "data", &obj.Data)
	if err != nil {
		err = core.SDKErrorf(err, "", "data-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "description", &obj.Description)
	if err != nil {
		err = core.SDKErrorf(err, "", "description-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "filters", &obj.Filters, UnmarshalAttributeFilter)
	if err != nil {
		err = core.SDKErrorf(err, "", "filters-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "id", &obj.ID)
	if err != nil {
		err = core.SDKErrorf(err, "", "id-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "limit", &obj.Limit, UnmarshalLimitParams)
	if err != nil {
		err = core.SDKErrorf(err, "", "limit-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "name", &obj.Name)
	if err != nil {
		err = core.SDKErrorf(err, "", "name-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "reportType", &obj.ReportType)
	if err != nil {
		err = core.SDKErrorf(err, "", "reportType-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "sort", &obj.Sort, UnmarshalAttributeSort)
	if err != nil {
		err = core.SDKErrorf(err, "", "sort-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// ComponentPreview : Specifies preview of a component.
type ComponentPreview struct {
	// Specifies a Report Component.
	Component *Component `json:"component,omitempty"`

	// Specifies list of global filters that are applicable to given components in the report.
	Filters []AttributeFilter `json:"filters,omitempty"`

	// Specifies the epoch timestamp in UTC in microseconds.
	GeneratedTimestampUsecs *int64 `json:"generatedTimestampUsecs,omitempty"`

	// Specifies the last refresh timestamp of data used to evaluate the component. If this parameter is not returned, then
	// 'generatedTimestampUsecs' can be used for last refreshed timestamp of the data. This is epoch timestamp in UTC in
	// microseconds.
	LastRefreshTimestampUsecs *int64 `json:"lastRefreshTimestampUsecs,omitempty"`

	// Specifies timezone of the user. If nil, defaults to UTC. The time specified should be a location name in the IANA
	// Time Zone database, for example, 'America/Los_Angeles'.
	Timezone *string `json:"timezone,omitempty"`
}

// UnmarshalComponentPreview unmarshals an instance of ComponentPreview from the specified map of raw messages.
func UnmarshalComponentPreview(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(ComponentPreview)
	err = core.UnmarshalModel(m, "component", &obj.Component, UnmarshalComponent)
	if err != nil {
		err = core.SDKErrorf(err, "", "component-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "filters", &obj.Filters, UnmarshalAttributeFilter)
	if err != nil {
		err = core.SDKErrorf(err, "", "filters-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "generatedTimestampUsecs", &obj.GeneratedTimestampUsecs)
	if err != nil {
		err = core.SDKErrorf(err, "", "generatedTimestampUsecs-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "lastRefreshTimestampUsecs", &obj.LastRefreshTimestampUsecs)
	if err != nil {
		err = core.SDKErrorf(err, "", "lastRefreshTimestampUsecs-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "timezone", &obj.Timezone)
	if err != nil {
		err = core.SDKErrorf(err, "", "timezone-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// Components : Specifies a list of report components.
type Components struct {
	// Specifies list of components.
	Components []Component `json:"components,omitempty"`
}

// UnmarshalComponents unmarshals an instance of Components from the specified map of raw messages.
func UnmarshalComponents(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(Components)
	err = core.UnmarshalModel(m, "components", &obj.Components, UnmarshalComponent)
	if err != nil {
		err = core.SDKErrorf(err, "", "components-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// CustomConfigParams : Specifies the configuration parameters to customize and format the columns in the report artifacts like excel, pdf
// etc.
type CustomConfigParams struct {
	// Specifies the configuration parameters to customize a component in excel report.
	XlsxParams *XlsxCustomConfigParams `json:"xlsxParams,omitempty"`
}

// UnmarshalCustomConfigParams unmarshals an instance of CustomConfigParams from the specified map of raw messages.
func UnmarshalCustomConfigParams(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(CustomConfigParams)
	err = core.UnmarshalModel(m, "xlsxParams", &obj.XlsxParams, UnmarshalXlsxCustomConfigParams)
	if err != nil {
		err = core.SDKErrorf(err, "", "xlsxParams-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// ExportReportOptions : The ExportReport options.
type ExportReportOptions struct {
	// Specifies the id of the report.
	ID *string `json:"id" validate:"required,ne="`

	// Specifies if the report should be generated asynchronously.
	Async *bool `json:"async,omitempty"`

	// Specifies list of global filters that are applicable to given components in the report.
	Filters []AttributeFilter `json:"filters,omitempty"`

	// The layout of the report which needs to be exported.
	Layout *string `json:"layout,omitempty"`

	// The format in which the report needs to be exported.
	ReportFormat *string `json:"reportFormat,omitempty"`

	// Specifies timezone of the user. If nil, defaults to UTC. The time specified should be a location name in the IANA
	// Time Zone database, for example, 'America/Los_Angeles'.
	Timezone *string `json:"timezone,omitempty"`

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// Constants associated with the ExportReportOptions.ReportFormat property.
// The format in which the report needs to be exported.
const (
	ExportReportOptions_ReportFormat_Csv = "CSV"
	ExportReportOptions_ReportFormat_Xls = "XLS"
)

// NewExportReportOptions : Instantiate ExportReportOptions
func (*BackupRecoveryManagementReportingApiV1) NewExportReportOptions(id string) *ExportReportOptions {
	return &ExportReportOptions{
		ID: core.StringPtr(id),
	}
}

// SetID : Allow user to set ID
func (_options *ExportReportOptions) SetID(id string) *ExportReportOptions {
	_options.ID = core.StringPtr(id)
	return _options
}

// SetAsync : Allow user to set Async
func (_options *ExportReportOptions) SetAsync(async bool) *ExportReportOptions {
	_options.Async = core.BoolPtr(async)
	return _options
}

// SetFilters : Allow user to set Filters
func (_options *ExportReportOptions) SetFilters(filters []AttributeFilter) *ExportReportOptions {
	_options.Filters = filters
	return _options
}

// SetLayout : Allow user to set Layout
func (_options *ExportReportOptions) SetLayout(layout string) *ExportReportOptions {
	_options.Layout = core.StringPtr(layout)
	return _options
}

// SetReportFormat : Allow user to set ReportFormat
func (_options *ExportReportOptions) SetReportFormat(reportFormat string) *ExportReportOptions {
	_options.ReportFormat = core.StringPtr(reportFormat)
	return _options
}

// SetTimezone : Allow user to set Timezone
func (_options *ExportReportOptions) SetTimezone(timezone string) *ExportReportOptions {
	_options.Timezone = core.StringPtr(timezone)
	return _options
}

// SetHeaders : Allow user to set Headers
func (options *ExportReportOptions) SetHeaders(param map[string]string) *ExportReportOptions {
	options.Headers = param
	return options
}

// ExternalTarget : Specifies an External target.
type ExternalTarget struct {
	// Specifies the id of the External target.
	ID *string `json:"id,omitempty"`

	// Specifies the name of the External target.
	Name *string `json:"name,omitempty"`

	// Specifies the id of the System.
	SystemID *string `json:"systemId,omitempty"`

	// Specifies the name of the System.
	SystemName *string `json:"systemName,omitempty"`

	// Specifies the type of the External target.
	TargetType *string `json:"targetType,omitempty"`
}

// UnmarshalExternalTarget unmarshals an instance of ExternalTarget from the specified map of raw messages.
func UnmarshalExternalTarget(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(ExternalTarget)
	err = core.UnmarshalPrimitive(m, "id", &obj.ID)
	if err != nil {
		err = core.SDKErrorf(err, "", "id-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "name", &obj.Name)
	if err != nil {
		err = core.SDKErrorf(err, "", "name-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "systemId", &obj.SystemID)
	if err != nil {
		err = core.SDKErrorf(err, "", "systemId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "systemName", &obj.SystemName)
	if err != nil {
		err = core.SDKErrorf(err, "", "systemName-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "targetType", &obj.TargetType)
	if err != nil {
		err = core.SDKErrorf(err, "", "targetType-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// GetComponentByIdOptions : The GetComponentByID options.
type GetComponentByIdOptions struct {
	// Specifies the id of the report component.
	ID *string `json:"id" validate:"required,ne="`

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// NewGetComponentByIdOptions : Instantiate GetComponentByIdOptions
func (*BackupRecoveryManagementReportingApiV1) NewGetComponentByIdOptions(id string) *GetComponentByIdOptions {
	return &GetComponentByIdOptions{
		ID: core.StringPtr(id),
	}
}

// SetID : Allow user to set ID
func (_options *GetComponentByIdOptions) SetID(id string) *GetComponentByIdOptions {
	_options.ID = core.StringPtr(id)
	return _options
}

// SetHeaders : Allow user to set Headers
func (options *GetComponentByIdOptions) SetHeaders(param map[string]string) *GetComponentByIdOptions {
	options.Headers = param
	return options
}

// GetComponentPreviewOptions : The GetComponentPreview options.
type GetComponentPreviewOptions struct {
	// Specifies the id of the component.
	ID *string `json:"id" validate:"required,ne="`

	// Specifies list of global filters that are applicable to given components in the report.
	Filters []AttributeFilter `json:"filters,omitempty"`

	// Specifies the parameters to limit the resulting dataset.
	Limit *LimitParams `json:"limit,omitempty"`

	// Specifies the sorting (ordering) parameters to be applied to the resulting data.
	Sort []AttributeSort `json:"sort,omitempty"`

	// Specifies timezone of the user. If nil, defaults to UTC. The time specified should be a location name in the IANA
	// Time Zone database, for example, 'America/Los_Angeles'.
	Timezone *string `json:"timezone,omitempty"`

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// NewGetComponentPreviewOptions : Instantiate GetComponentPreviewOptions
func (*BackupRecoveryManagementReportingApiV1) NewGetComponentPreviewOptions(id string) *GetComponentPreviewOptions {
	return &GetComponentPreviewOptions{
		ID: core.StringPtr(id),
	}
}

// SetID : Allow user to set ID
func (_options *GetComponentPreviewOptions) SetID(id string) *GetComponentPreviewOptions {
	_options.ID = core.StringPtr(id)
	return _options
}

// SetFilters : Allow user to set Filters
func (_options *GetComponentPreviewOptions) SetFilters(filters []AttributeFilter) *GetComponentPreviewOptions {
	_options.Filters = filters
	return _options
}

// SetLimit : Allow user to set Limit
func (_options *GetComponentPreviewOptions) SetLimit(limit *LimitParams) *GetComponentPreviewOptions {
	_options.Limit = limit
	return _options
}

// SetSort : Allow user to set Sort
func (_options *GetComponentPreviewOptions) SetSort(sort []AttributeSort) *GetComponentPreviewOptions {
	_options.Sort = sort
	return _options
}

// SetTimezone : Allow user to set Timezone
func (_options *GetComponentPreviewOptions) SetTimezone(timezone string) *GetComponentPreviewOptions {
	_options.Timezone = core.StringPtr(timezone)
	return _options
}

// SetHeaders : Allow user to set Headers
func (options *GetComponentPreviewOptions) SetHeaders(param map[string]string) *GetComponentPreviewOptions {
	options.Headers = param
	return options
}

// GetComponentsOptions : The GetComponents options.
type GetComponentsOptions struct {
	// Specifies the ids of the report component to fetch.
	Ids []string `json:"ids,omitempty"`

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// NewGetComponentsOptions : Instantiate GetComponentsOptions
func (*BackupRecoveryManagementReportingApiV1) NewGetComponentsOptions() *GetComponentsOptions {
	return &GetComponentsOptions{}
}

// SetIds : Allow user to set Ids
func (_options *GetComponentsOptions) SetIds(ids []string) *GetComponentsOptions {
	_options.Ids = ids
	return _options
}

// SetHeaders : Allow user to set Headers
func (options *GetComponentsOptions) SetHeaders(param map[string]string) *GetComponentsOptions {
	options.Headers = param
	return options
}

// GetProviderInstancesOptions : The GetProviderInstances options.
type GetProviderInstancesOptions struct {
	// List of Provider Instance IDs to filter on.
	InstanceIds []string `json:"instanceIds,omitempty"`

	// List of regions to filter Provider Instances on.
	Regions []string `json:"regions,omitempty"`

	// Bool flag to check whether to include other details like cluster details and software version along with status for
	// the service instances.
	IncludeServiceInstanceStatus *bool `json:"includeServiceInstanceStatus,omitempty"`

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// NewGetProviderInstancesOptions : Instantiate GetProviderInstancesOptions
func (*BackupRecoveryManagementReportingApiV1) NewGetProviderInstancesOptions() *GetProviderInstancesOptions {
	return &GetProviderInstancesOptions{}
}

// SetInstanceIds : Allow user to set InstanceIds
func (_options *GetProviderInstancesOptions) SetInstanceIds(instanceIds []string) *GetProviderInstancesOptions {
	_options.InstanceIds = instanceIds
	return _options
}

// SetRegions : Allow user to set Regions
func (_options *GetProviderInstancesOptions) SetRegions(regions []string) *GetProviderInstancesOptions {
	_options.Regions = regions
	return _options
}

// SetIncludeServiceInstanceStatus : Allow user to set IncludeServiceInstanceStatus
func (_options *GetProviderInstancesOptions) SetIncludeServiceInstanceStatus(includeServiceInstanceStatus bool) *GetProviderInstancesOptions {
	_options.IncludeServiceInstanceStatus = core.BoolPtr(includeServiceInstanceStatus)
	return _options
}

// SetHeaders : Allow user to set Headers
func (options *GetProviderInstancesOptions) SetHeaders(param map[string]string) *GetProviderInstancesOptions {
	options.Headers = param
	return options
}

// GetReportByIdOptions : The GetReportByID options.
type GetReportByIdOptions struct {
	// Specifies the id of the report.
	ID *string `json:"id" validate:"required,ne="`

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// NewGetReportByIdOptions : Instantiate GetReportByIdOptions
func (*BackupRecoveryManagementReportingApiV1) NewGetReportByIdOptions(id string) *GetReportByIdOptions {
	return &GetReportByIdOptions{
		ID: core.StringPtr(id),
	}
}

// SetID : Allow user to set ID
func (_options *GetReportByIdOptions) SetID(id string) *GetReportByIdOptions {
	_options.ID = core.StringPtr(id)
	return _options
}

// SetHeaders : Allow user to set Headers
func (options *GetReportByIdOptions) SetHeaders(param map[string]string) *GetReportByIdOptions {
	options.Headers = param
	return options
}

// GetReportPreviewOptions : The GetReportPreview options.
type GetReportPreviewOptions struct {
	// Specifies the id of the report.
	ID *string `json:"id" validate:"required,ne="`

	// Specifies list of components ids to be evaluated for the given report. If not specified, then all the components are
	// evaluated.
	ComponentIds []string `json:"componentIds,omitempty"`

	// Specifies list of global filters that are applicable to given components in the report.
	Filters []AttributeFilter `json:"filters,omitempty"`

	// Specifies timezone of the user. If nil, defaults to UTC. The time specified should be a location name in the IANA
	// Time Zone database, for example, 'America/Los_Angeles'.
	Timezone *string `json:"timezone,omitempty"`

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// NewGetReportPreviewOptions : Instantiate GetReportPreviewOptions
func (*BackupRecoveryManagementReportingApiV1) NewGetReportPreviewOptions(id string) *GetReportPreviewOptions {
	return &GetReportPreviewOptions{
		ID: core.StringPtr(id),
	}
}

// SetID : Allow user to set ID
func (_options *GetReportPreviewOptions) SetID(id string) *GetReportPreviewOptions {
	_options.ID = core.StringPtr(id)
	return _options
}

// SetComponentIds : Allow user to set ComponentIds
func (_options *GetReportPreviewOptions) SetComponentIds(componentIds []string) *GetReportPreviewOptions {
	_options.ComponentIds = componentIds
	return _options
}

// SetFilters : Allow user to set Filters
func (_options *GetReportPreviewOptions) SetFilters(filters []AttributeFilter) *GetReportPreviewOptions {
	_options.Filters = filters
	return _options
}

// SetTimezone : Allow user to set Timezone
func (_options *GetReportPreviewOptions) SetTimezone(timezone string) *GetReportPreviewOptions {
	_options.Timezone = core.StringPtr(timezone)
	return _options
}

// SetHeaders : Allow user to set Headers
func (options *GetReportPreviewOptions) SetHeaders(param map[string]string) *GetReportPreviewOptions {
	options.Headers = param
	return options
}

// GetReportTypeOptions : The GetReportType options.
type GetReportTypeOptions struct {
	// Specifies the report type.
	ReportType *string `json:"reportType" validate:"required,ne="`

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// Constants associated with the GetReportTypeOptions.ReportType property.
// Specifies the report type.
const (
	GetReportTypeOptions_ReportType_Failures                    = "Failures"
	GetReportTypeOptions_ReportType_Protectedobjects            = "ProtectedObjects"
	GetReportTypeOptions_ReportType_Protectedunprotectedobjects = "ProtectedUnprotectedObjects"
	GetReportTypeOptions_ReportType_Protectionactivity          = "ProtectionActivity"
	GetReportTypeOptions_ReportType_Protectiongroupsummary      = "ProtectionGroupSummary"
	GetReportTypeOptions_ReportType_Protectionruns              = "ProtectionRuns"
	GetReportTypeOptions_ReportType_Protectionrunstrend         = "ProtectionRunsTrend"
	GetReportTypeOptions_ReportType_Recovery                    = "Recovery"
)

// NewGetReportTypeOptions : Instantiate GetReportTypeOptions
func (*BackupRecoveryManagementReportingApiV1) NewGetReportTypeOptions(reportType string) *GetReportTypeOptions {
	return &GetReportTypeOptions{
		ReportType: core.StringPtr(reportType),
	}
}

// SetReportType : Allow user to set ReportType
func (_options *GetReportTypeOptions) SetReportType(reportType string) *GetReportTypeOptions {
	_options.ReportType = core.StringPtr(reportType)
	return _options
}

// SetHeaders : Allow user to set Headers
func (options *GetReportTypeOptions) SetHeaders(param map[string]string) *GetReportTypeOptions {
	options.Headers = param
	return options
}

// GetReportsOptions : The GetReports options.
type GetReportsOptions struct {
	// Specifies the ids of reports to fetch.
	Ids []string `json:"ids,omitempty"`

	// Specifies the user context to filter reports.
	UserContext *string `json:"userContext,omitempty"`

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// Constants associated with the GetReportsOptions.UserContext property.
// Specifies the user context to filter reports.
const (
	GetReportsOptions_UserContext_Ibmbaas = "IBMBaaS"
)

// NewGetReportsOptions : Instantiate GetReportsOptions
func (*BackupRecoveryManagementReportingApiV1) NewGetReportsOptions() *GetReportsOptions {
	return &GetReportsOptions{}
}

// SetIds : Allow user to set Ids
func (_options *GetReportsOptions) SetIds(ids []string) *GetReportsOptions {
	_options.Ids = ids
	return _options
}

// SetUserContext : Allow user to set UserContext
func (_options *GetReportsOptions) SetUserContext(userContext string) *GetReportsOptions {
	_options.UserContext = core.StringPtr(userContext)
	return _options
}

// SetHeaders : Allow user to set Headers
func (options *GetReportsOptions) SetHeaders(param map[string]string) *GetReportsOptions {
	options.Headers = param
	return options
}

// GetResourcesOptions : The GetResources options.
type GetResourcesOptions struct {
	// Specifies the type of the resource.
	ResourceType *string `json:"resourceType" validate:"required"`

	// Allows users to set headers on API requests.
	Headers map[string]string
}

// Constants associated with the GetResourcesOptions.ResourceType property.
// Specifies the type of the resource.
const (
	GetResourcesOptions_ResourceType_Messagecodemappings = "MessageCodeMappings"
	GetResourcesOptions_ResourceType_Policies            = "Policies"
	GetResourcesOptions_ResourceType_Protectiongroups    = "ProtectionGroups"
	GetResourcesOptions_ResourceType_Registeredsources   = "RegisteredSources"
)

// NewGetResourcesOptions : Instantiate GetResourcesOptions
func (*BackupRecoveryManagementReportingApiV1) NewGetResourcesOptions(resourceType string) *GetResourcesOptions {
	return &GetResourcesOptions{
		ResourceType: core.StringPtr(resourceType),
	}
}

// SetResourceType : Allow user to set ResourceType
func (_options *GetResourcesOptions) SetResourceType(resourceType string) *GetResourcesOptions {
	_options.ResourceType = core.StringPtr(resourceType)
	return _options
}

// SetHeaders : Allow user to set Headers
func (options *GetResourcesOptions) SetHeaders(param map[string]string) *GetResourcesOptions {
	options.Headers = param
	return options
}

// IbmServiceInstance : Specifies the IBM Service Instance Details like instance id, name, region and other metadata associated with the
// service instance.
type IbmServiceInstance struct {
	// Specifies the cluster Id of the cluster where the provider instance is located.
	ClusterID *interface{} `json:"clusterId,omitempty"`

	// Specifies the Incarnation id associated with the cluster.
	ClusterIncarnationID *interface{} `json:"clusterIncarnationId,omitempty"`

	// Specifies the provider instance uuid used to uniquely identify the instance.
	InstanceID *string `json:"instanceId,omitempty"`

	// Specifies the latitude of provider instance location.
	Lat *float64 `json:"lat,omitempty"`

	// Specifies the longitude of provider instance location.
	Lon *float64 `json:"lon,omitempty"`

	// Specifies the provider instance name.
	Name *string `json:"name,omitempty"`

	// Specifies the region where the provider instance is present.
	Region *string `json:"region,omitempty"`

	// Specifies the software Version of the cluster.
	SoftwareVersion *string `json:"softwareVersion,omitempty"`

	// Species the status of the service instance. Can be one of Active, Standby, InMigration/.
	Status *string `json:"status,omitempty"`
}

// Constants associated with the IbmServiceInstance.Status property.
// Species the status of the service instance. Can be one of Active, Standby, InMigration/.
const (
	IbmServiceInstance_Status_Active      = "Active"
	IbmServiceInstance_Status_Inmigration = "InMigration"
	IbmServiceInstance_Status_Standby     = "StandBy"
)

// UnmarshalIbmServiceInstance unmarshals an instance of IbmServiceInstance from the specified map of raw messages.
func UnmarshalIbmServiceInstance(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(IbmServiceInstance)
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
	err = core.UnmarshalPrimitive(m, "instanceId", &obj.InstanceID)
	if err != nil {
		err = core.SDKErrorf(err, "", "instanceId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "lat", &obj.Lat)
	if err != nil {
		err = core.SDKErrorf(err, "", "lat-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "lon", &obj.Lon)
	if err != nil {
		err = core.SDKErrorf(err, "", "lon-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "name", &obj.Name)
	if err != nil {
		err = core.SDKErrorf(err, "", "name-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "region", &obj.Region)
	if err != nil {
		err = core.SDKErrorf(err, "", "region-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "softwareVersion", &obj.SoftwareVersion)
	if err != nil {
		err = core.SDKErrorf(err, "", "softwareVersion-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "status", &obj.Status)
	if err != nil {
		err = core.SDKErrorf(err, "", "status-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// InFilterParams : Specifies the in filter that are applied on attributes.
type InFilterParams struct {
	// Specifies the data type of the attribute.
	AttributeDataType *string `json:"attributeDataType" validate:"required"`

	// Specifies the optional label values for the attribute.
	AttributeLabels []string `json:"attributeLabels,omitempty"`

	// Specifies list of boolean values to filter results on.
	BoolFilterValues []bool `json:"boolFilterValues,omitempty"`

	// Specifies list of int32 values to filter results on.
	Int32FilterValues []int64 `json:"int32FilterValues,omitempty"`

	// Specifies list of int64 values to filter results on.
	Int64FilterValues []int64 `json:"int64FilterValues,omitempty"`

	// Specifies list of string values to filter results on.
	StringFilterValues []string `json:"stringFilterValues,omitempty"`
}

// Constants associated with the InFilterParams.AttributeDataType property.
// Specifies the data type of the attribute.
const (
	InFilterParams_AttributeDataType_Bool        = "Bool"
	InFilterParams_AttributeDataType_Float64     = "Float64"
	InFilterParams_AttributeDataType_Int32       = "Int32"
	InFilterParams_AttributeDataType_Int64       = "Int64"
	InFilterParams_AttributeDataType_Int64array  = "Int64Array"
	InFilterParams_AttributeDataType_String      = "String"
	InFilterParams_AttributeDataType_Stringarray = "StringArray"
)

// NewInFilterParams : Instantiate InFilterParams (Generic Model Constructor)
func (*BackupRecoveryManagementReportingApiV1) NewInFilterParams(attributeDataType string) (_model *InFilterParams, err error) {
	_model = &InFilterParams{
		AttributeDataType: core.StringPtr(attributeDataType),
	}
	err = core.ValidateStruct(_model, "required parameters")
	if err != nil {
		err = core.SDKErrorf(err, "", "model-missing-required", common.GetComponentInfo())
	}
	return
}

// UnmarshalInFilterParams unmarshals an instance of InFilterParams from the specified map of raw messages.
func UnmarshalInFilterParams(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(InFilterParams)
	err = core.UnmarshalPrimitive(m, "attributeDataType", &obj.AttributeDataType)
	if err != nil {
		err = core.SDKErrorf(err, "", "attributeDataType-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "attributeLabels", &obj.AttributeLabels)
	if err != nil {
		err = core.SDKErrorf(err, "", "attributeLabels-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "boolFilterValues", &obj.BoolFilterValues)
	if err != nil {
		err = core.SDKErrorf(err, "", "boolFilterValues-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "int32FilterValues", &obj.Int32FilterValues)
	if err != nil {
		err = core.SDKErrorf(err, "", "int32FilterValues-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "int64FilterValues", &obj.Int64FilterValues)
	if err != nil {
		err = core.SDKErrorf(err, "", "int64FilterValues-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "stringFilterValues", &obj.StringFilterValues)
	if err != nil {
		err = core.SDKErrorf(err, "", "stringFilterValues-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// LimitParams : Specifies the parameters to limit the resulting dataset.
type LimitParams struct {
	// Specifies the offset to which resulting data will be skipped before applying the size parameter. For example if
	// dataset size is 10 objects, from=2 and size=5, then from 10 objects only 5 objects are returned starting from offset
	// 2 i.e., 2 to 7. If not specified, then none of the objects are skipped.
	From *int64 `json:"from,omitempty"`

	// Specifies the number of objects to be returned from the offset specified.
	Size *int64 `json:"size" validate:"required"`
}

// NewLimitParams : Instantiate LimitParams (Generic Model Constructor)
func (*BackupRecoveryManagementReportingApiV1) NewLimitParams(size int64) (_model *LimitParams, err error) {
	_model = &LimitParams{
		Size: core.Int64Ptr(size),
	}
	err = core.ValidateStruct(_model, "required parameters")
	if err != nil {
		err = core.SDKErrorf(err, "", "model-missing-required", common.GetComponentInfo())
	}
	return
}

// UnmarshalLimitParams unmarshals an instance of LimitParams from the specified map of raw messages.
func UnmarshalLimitParams(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(LimitParams)
	err = core.UnmarshalPrimitive(m, "from", &obj.From)
	if err != nil {
		err = core.SDKErrorf(err, "", "from-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "size", &obj.Size)
	if err != nil {
		err = core.SDKErrorf(err, "", "size-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// MessageCodeMapping : Specifies an error code and GUID.
type MessageCodeMapping struct {
	// Specifies the message code.
	MessageCode *string `json:"messageCode,omitempty"`

	// Specifies the message GUID.
	MessageGuid *string `json:"messageGuid,omitempty"`
}

// UnmarshalMessageCodeMapping unmarshals an instance of MessageCodeMapping from the specified map of raw messages.
func UnmarshalMessageCodeMapping(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(MessageCodeMapping)
	err = core.UnmarshalPrimitive(m, "messageCode", &obj.MessageCode)
	if err != nil {
		err = core.SDKErrorf(err, "", "messageCode-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "messageGuid", &obj.MessageGuid)
	if err != nil {
		err = core.SDKErrorf(err, "", "messageGuid-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// Policy : Specifies the Protection Policy resource.
type Policy struct {
	// Specifies the id of the Protection Policy.
	ID *string `json:"id,omitempty"`

	// Specifies whether this is a Global Policy.
	IsGlobalPolicy *bool `json:"isGlobalPolicy,omitempty"`

	// Specifies the name of the Protection Policy.
	Name *string `json:"name,omitempty"`

	// Specifies the id of the System.
	SystemID *string `json:"systemId,omitempty"`

	// Specifies the name of the System.
	SystemName *string `json:"systemName,omitempty"`
}

// UnmarshalPolicy unmarshals an instance of Policy from the specified map of raw messages.
func UnmarshalPolicy(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(Policy)
	err = core.UnmarshalPrimitive(m, "id", &obj.ID)
	if err != nil {
		err = core.SDKErrorf(err, "", "id-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "isGlobalPolicy", &obj.IsGlobalPolicy)
	if err != nil {
		err = core.SDKErrorf(err, "", "isGlobalPolicy-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "name", &obj.Name)
	if err != nil {
		err = core.SDKErrorf(err, "", "name-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "systemId", &obj.SystemID)
	if err != nil {
		err = core.SDKErrorf(err, "", "systemId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "systemName", &obj.SystemName)
	if err != nil {
		err = core.SDKErrorf(err, "", "systemName-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// ProtectionGroup : Specifies the Protection Group resource.
type ProtectionGroup struct {
	// Specifies the id of the Protection Group.
	ID *string `json:"id,omitempty"`

	// Specifies the name of the Protection Group.
	Name *string `json:"name,omitempty"`

	// Specifies the id of the System.
	SystemID *string `json:"systemId,omitempty"`

	// Specifies the name of the System.
	SystemName *string `json:"systemName,omitempty"`
}

// UnmarshalProtectionGroup unmarshals an instance of ProtectionGroup from the specified map of raw messages.
func UnmarshalProtectionGroup(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(ProtectionGroup)
	err = core.UnmarshalPrimitive(m, "id", &obj.ID)
	if err != nil {
		err = core.SDKErrorf(err, "", "id-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "name", &obj.Name)
	if err != nil {
		err = core.SDKErrorf(err, "", "name-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "systemId", &obj.SystemID)
	if err != nil {
		err = core.SDKErrorf(err, "", "systemId-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "systemName", &obj.SystemName)
	if err != nil {
		err = core.SDKErrorf(err, "", "systemName-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// ProviderInstancesList : Specifies the list of Provider Instances along with their metadata.
type ProviderInstancesList struct {
	// Specifies the list of IBM Service Instances.
	IbmServiceInstances []IbmServiceInstance `json:"ibmServiceInstances,omitempty"`
}

// UnmarshalProviderInstancesList unmarshals an instance of ProviderInstancesList from the specified map of raw messages.
func UnmarshalProviderInstancesList(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(ProviderInstancesList)
	err = core.UnmarshalModel(m, "ibmServiceInstances", &obj.IbmServiceInstances, UnmarshalIbmServiceInstance)
	if err != nil {
		err = core.SDKErrorf(err, "", "ibmServiceInstances-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// RangeFilterParams : Specifies the filters that are applied on attributes.
type RangeFilterParams struct {
	// Specifies the lower bound value. If specified, all the results which are greater than this value will be returned.
	LowerBound *int64 `json:"lowerBound,omitempty"`

	// Specifies the upper bound value. If specified, all the results which are lesser than this value will be returned.
	UpperBound *int64 `json:"upperBound,omitempty"`
}

// UnmarshalRangeFilterParams unmarshals an instance of RangeFilterParams from the specified map of raw messages.
func UnmarshalRangeFilterParams(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(RangeFilterParams)
	err = core.UnmarshalPrimitive(m, "lowerBound", &obj.LowerBound)
	if err != nil {
		err = core.SDKErrorf(err, "", "lowerBound-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "upperBound", &obj.UpperBound)
	if err != nil {
		err = core.SDKErrorf(err, "", "upperBound-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// RegisteredSource : Specifies the Registered Source resource.
type RegisteredSource struct {
	// Specifies list of all environments discovered as part of this source.
	Environments []string `json:"environments,omitempty"`

	// Specifies the name of the registered source.
	Name *string `json:"name,omitempty"`

	// Specifies the unique identifier of registered source.
	UUID *string `json:"uuid,omitempty"`
}

// UnmarshalRegisteredSource unmarshals an instance of RegisteredSource from the specified map of raw messages.
func UnmarshalRegisteredSource(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(RegisteredSource)
	err = core.UnmarshalPrimitive(m, "environments", &obj.Environments)
	if err != nil {
		err = core.SDKErrorf(err, "", "environments-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "name", &obj.Name)
	if err != nil {
		err = core.SDKErrorf(err, "", "name-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "uuid", &obj.UUID)
	if err != nil {
		err = core.SDKErrorf(err, "", "uuid-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// Report : Specifies a Report.
type Report struct {
	// Specifies categoty of the Report.
	Category *string `json:"category,omitempty"`

	// Specifies the list of component ids in the Report.
	ComponentIds []string `json:"componentIds,omitempty"`

	// Specifies description of the Report.
	Description *string `json:"description,omitempty"`

	// Specifies the id of the report.
	ID *string `json:"id,omitempty"`

	// Specifies all the supported user contexts for this report.
	SupportedUserContexts []string `json:"supportedUserContexts,omitempty"`

	// Specifies the title of the report.
	Title *string `json:"title,omitempty"`
}

// Constants associated with the Report.Category property.
// Specifies categoty of the Report.
const (
	Report_Category_Compliance = "Compliance"
	Report_Category_Protection = "Protection"
	Report_Category_Storage    = "Storage"
)

// Constants associated with the Report.SupportedUserContexts property.
const (
	Report_SupportedUserContexts_Ibmbaas = "IBMBaaS"
)

// UnmarshalReport unmarshals an instance of Report from the specified map of raw messages.
func UnmarshalReport(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(Report)
	err = core.UnmarshalPrimitive(m, "category", &obj.Category)
	if err != nil {
		err = core.SDKErrorf(err, "", "category-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "componentIds", &obj.ComponentIds)
	if err != nil {
		err = core.SDKErrorf(err, "", "componentIds-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "description", &obj.Description)
	if err != nil {
		err = core.SDKErrorf(err, "", "description-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "id", &obj.ID)
	if err != nil {
		err = core.SDKErrorf(err, "", "id-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "supportedUserContexts", &obj.SupportedUserContexts)
	if err != nil {
		err = core.SDKErrorf(err, "", "supportedUserContexts-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "title", &obj.Title)
	if err != nil {
		err = core.SDKErrorf(err, "", "title-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// ReportPreview : Specifies preview of a Report.
type ReportPreview struct {
	// Specifies the component params and data.
	Components []Component `json:"components,omitempty"`

	// Specifies list of global filters that are applicable to given components in the report.
	Filters []AttributeFilter `json:"filters,omitempty"`

	// Specifies the epoch timestamp in UTC in microseconds.
	GeneratedTimestampUsecs *int64 `json:"generatedTimestampUsecs,omitempty"`

	// Specifies the id of the report.
	ID *string `json:"id,omitempty"`

	// Specifies the last refresh timestamp of data used to evaluate the component. If this parameter is not returned, then
	// 'generatedTimestampUsecs' can be used for last refreshed timestamp of the data. This is epoch timestamp in UTC in
	// microseconds.
	LastRefreshTimestampUsecs *int64 `json:"lastRefreshTimestampUsecs,omitempty"`

	// Specifies timezone of the user. If nil, defaults to UTC. The time specified should be a location name in the IANA
	// Time Zone database, for example, 'America/Los_Angeles'.
	Timezone *string `json:"timezone,omitempty"`

	// Specifies the title of the report.
	Title *string `json:"title,omitempty"`
}

// UnmarshalReportPreview unmarshals an instance of ReportPreview from the specified map of raw messages.
func UnmarshalReportPreview(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(ReportPreview)
	err = core.UnmarshalModel(m, "components", &obj.Components, UnmarshalComponent)
	if err != nil {
		err = core.SDKErrorf(err, "", "components-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "filters", &obj.Filters, UnmarshalAttributeFilter)
	if err != nil {
		err = core.SDKErrorf(err, "", "filters-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "generatedTimestampUsecs", &obj.GeneratedTimestampUsecs)
	if err != nil {
		err = core.SDKErrorf(err, "", "generatedTimestampUsecs-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "id", &obj.ID)
	if err != nil {
		err = core.SDKErrorf(err, "", "id-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "lastRefreshTimestampUsecs", &obj.LastRefreshTimestampUsecs)
	if err != nil {
		err = core.SDKErrorf(err, "", "lastRefreshTimestampUsecs-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "timezone", &obj.Timezone)
	if err != nil {
		err = core.SDKErrorf(err, "", "timezone-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "title", &obj.Title)
	if err != nil {
		err = core.SDKErrorf(err, "", "title-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// ReportTypeAttribute : Specifies an attribute and its properties for a given report type.
type ReportTypeAttribute struct {
	// Specifies the data type of the attribute.
	DataType *string `json:"dataType,omitempty"`

	// Specifies the attribute name.
	Name *string `json:"name,omitempty"`
}

// Constants associated with the ReportTypeAttribute.DataType property.
// Specifies the data type of the attribute.
const (
	ReportTypeAttribute_DataType_Bool        = "Bool"
	ReportTypeAttribute_DataType_Float64     = "Float64"
	ReportTypeAttribute_DataType_Int32       = "Int32"
	ReportTypeAttribute_DataType_Int64       = "Int64"
	ReportTypeAttribute_DataType_Int64array  = "Int64Array"
	ReportTypeAttribute_DataType_String      = "String"
	ReportTypeAttribute_DataType_Stringarray = "StringArray"
)

// UnmarshalReportTypeAttribute unmarshals an instance of ReportTypeAttribute from the specified map of raw messages.
func UnmarshalReportTypeAttribute(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(ReportTypeAttribute)
	err = core.UnmarshalPrimitive(m, "dataType", &obj.DataType)
	if err != nil {
		err = core.SDKErrorf(err, "", "dataType-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "name", &obj.Name)
	if err != nil {
		err = core.SDKErrorf(err, "", "name-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// ReportTypeAttributes : Specifies list of attributes and properties for a given report type.
type ReportTypeAttributes struct {
	// Specifies the attribute name.
	Attributes []ReportTypeAttribute `json:"attributes,omitempty"`
}

// UnmarshalReportTypeAttributes unmarshals an instance of ReportTypeAttributes from the specified map of raw messages.
func UnmarshalReportTypeAttributes(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(ReportTypeAttributes)
	err = core.UnmarshalModel(m, "attributes", &obj.Attributes, UnmarshalReportTypeAttribute)
	if err != nil {
		err = core.SDKErrorf(err, "", "attributes-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// Reports : Specifies a list of Reports.
type Reports struct {
	// Specifies list of reports.
	Reports []Report `json:"reports,omitempty"`
}

// UnmarshalReports unmarshals an instance of Reports from the specified map of raw messages.
func UnmarshalReports(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(Reports)
	err = core.UnmarshalModel(m, "reports", &obj.Reports, UnmarshalReport)
	if err != nil {
		err = core.SDKErrorf(err, "", "reports-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// Resources : Specifies the resources response.
type Resources struct {
	// Specifies the list of External Targets.
	ExternalTargets []ExternalTarget `json:"externalTargets,omitempty"`

	// Specifies the list of Message codes.
	MessageCodeMappings []MessageCodeMapping `json:"messageCodeMappings,omitempty"`

	// Specifies the list of Protection Groups.
	Policies []Policy `json:"policies,omitempty"`

	// Specifies the list of Protection Groups.
	ProtectionGroups []ProtectionGroup `json:"protectionGroups,omitempty"`

	// Specifies the type of the resource.
	ResourceType *string `json:"resourceType,omitempty"`

	// Specifies the list of Registered sources.
	Sources []RegisteredSource `json:"sources,omitempty"`

	// Specifies the list of Tenants.
	Tenants []Tenant `json:"tenants,omitempty"`
}

// Constants associated with the Resources.ResourceType property.
// Specifies the type of the resource.
const (
	Resources_ResourceType_Messagecodemappings = "MessageCodeMappings"
	Resources_ResourceType_Policies            = "Policies"
	Resources_ResourceType_Protectiongroups    = "ProtectionGroups"
	Resources_ResourceType_Registeredsources   = "RegisteredSources"
)

// UnmarshalResources unmarshals an instance of Resources from the specified map of raw messages.
func UnmarshalResources(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(Resources)
	err = core.UnmarshalModel(m, "externalTargets", &obj.ExternalTargets, UnmarshalExternalTarget)
	if err != nil {
		err = core.SDKErrorf(err, "", "externalTargets-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "messageCodeMappings", &obj.MessageCodeMappings, UnmarshalMessageCodeMapping)
	if err != nil {
		err = core.SDKErrorf(err, "", "messageCodeMappings-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "policies", &obj.Policies, UnmarshalPolicy)
	if err != nil {
		err = core.SDKErrorf(err, "", "policies-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "protectionGroups", &obj.ProtectionGroups, UnmarshalProtectionGroup)
	if err != nil {
		err = core.SDKErrorf(err, "", "protectionGroups-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "resourceType", &obj.ResourceType)
	if err != nil {
		err = core.SDKErrorf(err, "", "resourceType-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "sources", &obj.Sources, UnmarshalRegisteredSource)
	if err != nil {
		err = core.SDKErrorf(err, "", "sources-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalModel(m, "tenants", &obj.Tenants, UnmarshalTenant)
	if err != nil {
		err = core.SDKErrorf(err, "", "tenants-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// SystemsFilterParams : Specifies the systems filter. Specifying this will pre filter all the results provided list of system identifier
// before applying aggregations.
type SystemsFilterParams struct {
	// Specifies an array of system identifiers. System identifiers may be of format clusterid:clusterincarnationid or a
	// regionid (applicable only in case of DMaaS).
	SystemIds []string `json:"systemIds" validate:"required"`

	// Specifies the optional system names labels.
	SystemNames []string `json:"systemNames,omitempty"`
}

// NewSystemsFilterParams : Instantiate SystemsFilterParams (Generic Model Constructor)
func (*BackupRecoveryManagementReportingApiV1) NewSystemsFilterParams(systemIds []string) (_model *SystemsFilterParams, err error) {
	_model = &SystemsFilterParams{
		SystemIds: systemIds,
	}
	err = core.ValidateStruct(_model, "required parameters")
	if err != nil {
		err = core.SDKErrorf(err, "", "model-missing-required", common.GetComponentInfo())
	}
	return
}

// UnmarshalSystemsFilterParams unmarshals an instance of SystemsFilterParams from the specified map of raw messages.
func UnmarshalSystemsFilterParams(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(SystemsFilterParams)
	err = core.UnmarshalPrimitive(m, "systemIds", &obj.SystemIds)
	if err != nil {
		err = core.SDKErrorf(err, "", "systemIds-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "systemNames", &obj.SystemNames)
	if err != nil {
		err = core.SDKErrorf(err, "", "systemNames-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// // Tenant : Specifies the Tenant resource.
// type Tenant struct {
// 	// Specifies the id of the Tenant.
// 	ID *string `json:"id,omitempty"`

// 	// Specifies the name of the Tenant.
// 	Name *string `json:"name,omitempty"`
// }

// // UnmarshalTenant unmarshals an instance of Tenant from the specified map of raw messages.
// func UnmarshalTenant(m map[string]json.RawMessage, result interface{}) (err error) {
// 	obj := new(Tenant)
// 	err = core.UnmarshalPrimitive(m, "id", &obj.ID)
// 	if err != nil {
// 		err = core.SDKErrorf(err, "", "id-error", common.GetComponentInfo())
// 		return
// 	}
// 	err = core.UnmarshalPrimitive(m, "name", &obj.Name)
// 	if err != nil {
// 		err = core.SDKErrorf(err, "", "name-error", common.GetComponentInfo())
// 		return
// 	}
// 	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
// 	return
// }

// TimeRangeFilterParams : Specifies the time range filter. Specifying this will pre filter all the results on necessary resources like
// Protection Runs etc before applying aggregations. Currently, maximum allowed time range is 60 days.
type TimeRangeFilterParams struct {
	// Enum value for specifying the date range for a time filter. Considered only if lowerBound and upperBound are empty.
	DateRange *string `json:"dateRange,omitempty"`

	// Specifies the duration preceding the current time for which the data must be fetch i.e fetch data between
	// currentTime and currentTime - durationHours. This filter is only considered if neither upperBound, lowerBound or
	// dateRange is specified.
	DurationHours *int64 `json:"durationHours,omitempty"`

	// Specifies the lower bound value. If specified, all the results which are greater than this value will be returned.
	LowerBound *int64 `json:"lowerBound,omitempty"`

	// Specifies the upper bound value. If specified, all the results which are lesser than this value will be returned.
	UpperBound *int64 `json:"upperBound,omitempty"`
}

// Constants associated with the TimeRangeFilterParams.DateRange property.
// Enum value for specifying the date range for a time filter. Considered only if lowerBound and upperBound are empty.
const (
	TimeRangeFilterParams_DateRange_Currentmonth = "CurrentMonth"
	TimeRangeFilterParams_DateRange_Currentyear  = "CurrentYear"
	TimeRangeFilterParams_DateRange_Last180days  = "Last180Days"
	TimeRangeFilterParams_DateRange_Last1hour    = "Last1Hour"
	TimeRangeFilterParams_DateRange_Last24hours  = "Last24Hours"
	TimeRangeFilterParams_DateRange_Last30days   = "Last30Days"
	TimeRangeFilterParams_DateRange_Last365days  = "Last365Days"
	TimeRangeFilterParams_DateRange_Last3months  = "Last3Months"
	TimeRangeFilterParams_DateRange_Last6months  = "Last6Months"
	TimeRangeFilterParams_DateRange_Last7days    = "Last7Days"
	TimeRangeFilterParams_DateRange_Last90days   = "Last90Days"
	TimeRangeFilterParams_DateRange_Lastmonth    = "LastMonth"
	TimeRangeFilterParams_DateRange_Lastyear     = "LastYear"
)

// UnmarshalTimeRangeFilterParams unmarshals an instance of TimeRangeFilterParams from the specified map of raw messages.
func UnmarshalTimeRangeFilterParams(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(TimeRangeFilterParams)
	err = core.UnmarshalPrimitive(m, "dateRange", &obj.DateRange)
	if err != nil {
		err = core.SDKErrorf(err, "", "dateRange-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "durationHours", &obj.DurationHours)
	if err != nil {
		err = core.SDKErrorf(err, "", "durationHours-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "lowerBound", &obj.LowerBound)
	if err != nil {
		err = core.SDKErrorf(err, "", "lowerBound-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "upperBound", &obj.UpperBound)
	if err != nil {
		err = core.SDKErrorf(err, "", "upperBound-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// XlsxAttributeCustomConfigParams : Specifies the configuration parameters to customize the columns in excel report for a component. The ordering of the
// columns in the resulting xlsx will be done according to the order they are configured.
type XlsxAttributeCustomConfigParams struct {
	// Specifies the name of the attribute.
	AttributeName *string `json:"attributeName" validate:"required"`

	// Specifies a custom label for attribute to appear in the xlsx report. If not specified, default attribute name will
	// be used.
	CustomLabel *string `json:"customLabel,omitempty"`

	// Specifies a custom format for attribute to appear in the xlsx report. If not specified, the attribute value is sent
	// as-is.
	Format *string `json:"format,omitempty"`
}

// Constants associated with the XlsxAttributeCustomConfigParams.Format property.
// Specifies a custom format for attribute to appear in the xlsx report. If not specified, the attribute value is sent
// as-is.
const (
	XlsxAttributeCustomConfigParams_Format_Duration  = "Duration"
	XlsxAttributeCustomConfigParams_Format_Timestamp = "Timestamp"
)

// UnmarshalXlsxAttributeCustomConfigParams unmarshals an instance of XlsxAttributeCustomConfigParams from the specified map of raw messages.
func UnmarshalXlsxAttributeCustomConfigParams(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(XlsxAttributeCustomConfigParams)
	err = core.UnmarshalPrimitive(m, "attributeName", &obj.AttributeName)
	if err != nil {
		err = core.SDKErrorf(err, "", "attributeName-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "customLabel", &obj.CustomLabel)
	if err != nil {
		err = core.SDKErrorf(err, "", "customLabel-error", common.GetComponentInfo())
		return
	}
	err = core.UnmarshalPrimitive(m, "format", &obj.Format)
	if err != nil {
		err = core.SDKErrorf(err, "", "format-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// XlsxCustomConfigParams : Specifies the configuration parameters to customize a component in excel report.
type XlsxCustomConfigParams struct {
	// Specifies customized configuration for the attributes in the report. If not specified, all the attributes will be
	// sent as-is to the report without any formatting.
	AttributeConfig []XlsxAttributeCustomConfigParams `json:"attributeConfig,omitempty"`
}

// UnmarshalXlsxCustomConfigParams unmarshals an instance of XlsxCustomConfigParams from the specified map of raw messages.
func UnmarshalXlsxCustomConfigParams(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(XlsxCustomConfigParams)
	err = core.UnmarshalModel(m, "attributeConfig", &obj.AttributeConfig, UnmarshalXlsxAttributeCustomConfigParams)
	if err != nil {
		err = core.SDKErrorf(err, "", "attributeConfig-error", common.GetComponentInfo())
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}
