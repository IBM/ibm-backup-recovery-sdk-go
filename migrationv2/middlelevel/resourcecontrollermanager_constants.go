/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package middlelevel

// ResourceType represents the type of IBM Cloud resource
type ResourceType string

const (
	// ResourceTypeServiceInstance represents a service instance resource type
	ResourceTypeServiceInstance ResourceType = "service_instance"
)

// ResourceState represents the state of an IBM Cloud resource
type ResourceState string

const (
	// ResourceStateActive represents an active resource state
	ResourceStateActive ResourceState = "active"
	// ResourceStateInactive represents an inactive resource state
	ResourceStateInactive ResourceState = "inactive"
	// ResourceStatePending represents a pending resource state
	ResourceStatePending ResourceState = "pending"
	// ResourceStateFailed represents a failed resource state
	ResourceStateFailed ResourceState = "failed"
)

// CRN service identifiers for BRS instances
const (
	// CRNServiceBRS is the production BRS service identifier in CRN
	CRNServiceBRS = ":backup-recovery:"
	// CRNServiceBRSTest is the test BRS service identifier in CRN
	CRNServiceBRSTest = ":backup-recovery-tests:"
)

// Extension keys used in resource instance extensions
const (
	// ExtensionKeyTenantID is the key for tenant ID in extensions
	ExtensionKeyTenantID = "tenant-id"
	// ExtensionKeyEndpoints is the key for endpoints in extensions
	ExtensionKeyEndpoints = "endpoints"
	// ExtensionKeyEndpointPrivate is the key for private endpoint URL
	ExtensionKeyEndpointPrivate = "private"
	// ExtensionKeyEndpointPublic is the key for public endpoint URL
	ExtensionKeyEndpointPublic = "public"
)

// URL query parameters
const (
	// QueryParamStart is the pagination start token parameter
	QueryParamStart = "start"
)

// URL constants
const (
	// DummyBaseURL is used for parsing relative URLs
	BaseURL = "https://cloud.ibm.com"
	// URLPathPrefix is the prefix for relative URLs
	URLPathPrefix = "/"
)

// String returns the string representation of ResourceType
func (r ResourceType) String() string {
	return string(r)
}

// String returns the string representation of ResourceState
func (r ResourceState) String() string {
	return string(r)
}

// IsValid checks if the ResourceType is valid
func (r ResourceType) IsValid() bool {
	switch r {
	case ResourceTypeServiceInstance:
		return true
	default:
		return false
	}
}

// IsValid checks if the ResourceState is valid
func (r ResourceState) IsValid() bool {
	switch r {
	case ResourceStateActive, ResourceStateInactive, ResourceStatePending, ResourceStateFailed:
		return true
	default:
		return false
	}
}

// Made with Bob
