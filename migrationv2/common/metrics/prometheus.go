/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * (C) Copyright IBM Corp. 2026 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/
package metrics

import (
	"context"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/logger"
)

// ----------------------------------------------------------------------------
// SDK Metric Names - Pre-registered metrics
// ----------------------------------------------------------------------------

const (
	MetricOperationTotal                     = "operation_total"
	MetricOperationDuration                  = "operation_duration_seconds"
	MetricRetryTotal                         = "retry_total"
	MetricIBMCloudContainerServiceCallsTotal = "ibm_cloud_container_service_calls_total"
	MetricBRSAPITimeoutsTotal                = "brs_api_timeouts_total"
)

// Operation name constants for consistent metric labeling
const (
	// BRS Manager operations
	OperationBRSInitialize = "brs_initialize"

	// IKS operations
	OperationIKSApplyRBAC      = "iks_apply_rbac_get_kubeconfig"
	OperationIKSGetKubeApi     = "iks_get_kube_api"
	OperationIKSGetBearerToken = "iks_get_bearer_token"

	// Task API operations (these track end-to-end operations including BRS API calls)
	// Connection operations
	OperationCreateConnection = "create_connection"
	OperationGetConnection    = "get_connection"
	OperationListConnections  = "list_connections"
	OperationDeleteConnection = "delete_connection"

	// Connector operations
	OperationDeployConnector = "deploy_connector"
	OperationGetConnector    = "get_connector"
	OperationDeleteConnector = "delete_connector"

	// Registration operations
	OperationRegisterSource    = "register_source"
	OperationGetRegistration   = "get_registration"
	OperationListRegistrations = "list_registrations"
	OperationUnregisterSource  = "unregister_source"

	// Protection Group operations
	OperationCreateProtectionGroup = "create_protection_group"
	OperationGetProtectionGroup    = "get_protection_group"
	OperationListProtectionGroups  = "list_protection_groups"
	OperationUpdateProtectionGroup = "update_protection_group"
	OperationDeleteProtectionGroup = "delete_protection_group"

	// Policy operations
	OperationCreatePolicy = "create_policy"
	OperationGetPolicy    = "get_policy"
	OperationListPolicies = "list_policies"
	OperationUpdatePolicy = "update_policy"
	OperationDeletePolicy = "delete_policy"

	// Backup operations
	OperationRunBackup     = "run_backup"
	OperationGetBackup     = "get_backup"
	OperationListBackups   = "list_backups"
	OperationWaitForBackup = "wait_for_backup"

	// Restore operations
	OperationRunRestore     = "run_restore"
	OperationGetRestore     = "get_restore"
	OperationListRestores   = "list_restores"
	OperationWaitForRestore = "wait_for_restore"

	// Control operations
	OperationPauseBackup  = "pause_backup"
	OperationResumeBackup = "resume_backup"
	OperationAbortBackup  = "abort_backup"
	OperationAbortRestore = "abort_restore"
)

// OperationStatus represents the status of an operation.
type OperationStatus string

const (
	StatusSuccess   OperationStatus = "success"
	StatusFailure   OperationStatus = "failure"
	StatusRetry     OperationStatus = "retry"
	StatusExhausted OperationStatus = "exhausted"
)

// ----------------------------------------------------------------------------
// Configuration
// ----------------------------------------------------------------------------

// PrometheusConfig holds configuration for creating a Prometheus metrics instance.
type PrometheusConfig struct {
	Namespace string
	Registry  *prometheus.Registry
	Logger    logger.Logger
}

// ----------------------------------------------------------------------------
// PrometheusMetrics
// ----------------------------------------------------------------------------

// PrometheusMetrics implements the Metrics interface with pre-registered SDK metrics.
//
// SDK metrics are pre-registered at construction time. Additional custom metrics
// can be added using the generic interface methods (IncCounter, SetGauge, etc.).
type PrometheusMetrics struct {
	namespace string
	reg       prometheus.Registerer
	gatherer  prometheus.Gatherer
	logger    logger.Logger

	mu         sync.RWMutex
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
	histograms map[string]*prometheus.HistogramVec

	// Pre-registered SDK metrics (for type-safe access)
	operationTotal                     *prometheus.CounterVec
	operationDuration                  *prometheus.HistogramVec
	retryTotal                         *prometheus.CounterVec
	ibmCloudContainerServiceCallsTotal *prometheus.CounterVec
	brsAPITimeouts                     *prometheus.CounterVec
}

// NewPrometheus creates a new PrometheusMetrics instance with SDK metrics pre-registered.
func NewPrometheus(cfg PrometheusConfig) *PrometheusMetrics {
	var reg prometheus.Registerer
	var gatherer prometheus.Gatherer

	if cfg.Registry != nil {
		reg = cfg.Registry
		gatherer = cfg.Registry
	} else {
		reg = prometheus.DefaultRegisterer
		gatherer = prometheus.DefaultGatherer
	}

	log := cfg.Logger
	if log == nil {
		log = logger.NewNoop()
	}

	m := &PrometheusMetrics{
		namespace:  cfg.Namespace,
		reg:        reg,
		gatherer:   gatherer,
		logger:     log,
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		histograms: make(map[string]*prometheus.HistogramVec),
	}

	m.registerSDKMetrics()
	return m
}

// registerSDKMetrics pre-registers all SDK metrics.
func (m *PrometheusMetrics) registerSDKMetrics() {
	m.operationTotal = m.registerCounter(
		MetricOperationTotal,
		"Total number of operations completed",
		[]string{"accountId", "operation", "status"},
	)

	m.operationDuration = m.registerHistogram(
		MetricOperationDuration,
		"Duration of operations in seconds",
		[]float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		[]string{"accountId", "operation"},
	)

	m.retryTotal = m.registerCounter(
		MetricRetryTotal,
		"Total number of retry attempts",
		[]string{"accountId", "operation", "status"},
	)

	m.ibmCloudContainerServiceCallsTotal = m.registerCounter(
		MetricIBMCloudContainerServiceCallsTotal,
		"Total number of IBM Cloud Container Service Calls",
		[]string{"accountId"},
	)

	m.brsAPITimeouts = m.registerCounter(
		MetricBRSAPITimeoutsTotal,
		"Total number of BRS API timeouts",
		[]string{"accountId"},
	)
}

// ----------------------------------------------------------------------------
// Generic interface implementation (Metrics interface)
// ----------------------------------------------------------------------------

// IncCounter increments a counter by 1.
func (m *PrometheusMetrics) IncCounter(_ context.Context, name string, labels ...Label) {
	m.AddCounter(context.Background(), name, 1, labels...)
}

// AddCounter adds a value to a counter.
func (m *PrometheusMetrics) AddCounter(_ context.Context, name string, value float64, labels ...Label) {
	keys, values := sortLabels(labels)
	counter := m.getOrCreateCounter(name, keys)
	if counter != nil {
		counter.WithLabelValues(values...).Add(value)
	}
}

// SetGauge sets a gauge to a value.
func (m *PrometheusMetrics) SetGauge(_ context.Context, name string, value float64, labels ...Label) {
	keys, values := sortLabels(labels)
	gauge := m.getOrCreateGauge(name, keys)
	if gauge != nil {
		gauge.WithLabelValues(values...).Set(value)
	}
}

// IncGauge increments a gauge by 1.
func (m *PrometheusMetrics) IncGauge(_ context.Context, name string, labels ...Label) {
	keys, values := sortLabels(labels)
	gauge := m.getOrCreateGauge(name, keys)
	if gauge != nil {
		gauge.WithLabelValues(values...).Inc()
	}
}

// DecGauge decrements a gauge by 1.
func (m *PrometheusMetrics) DecGauge(_ context.Context, name string, labels ...Label) {
	keys, values := sortLabels(labels)
	gauge := m.getOrCreateGauge(name, keys)
	if gauge != nil {
		gauge.WithLabelValues(values...).Dec()
	}
}

// RecordDuration records a duration in a histogram.
func (m *PrometheusMetrics) RecordDuration(_ context.Context, name string, duration time.Duration, labels ...Label) {
	keys, values := sortLabels(labels)
	hist := m.getOrCreateHistogram(name, keys)
	if hist != nil {
		hist.WithLabelValues(values...).Observe(duration.Seconds())
	}
}

// Handler returns an HTTP handler for Prometheus metrics.
func (m *PrometheusMetrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.gatherer, promhttp.HandlerOpts{
		ErrorHandling: promhttp.ContinueOnError,
	})
}

// RecordOperation records an operation completion with its duration.
// Label order matches registration: accountId, operation, status.
func (m *PrometheusMetrics) RecordOperation(operation string, status OperationStatus, accountID string, duration time.Duration) {
	m.operationTotal.WithLabelValues(accountID, operation, string(status)).Inc()
	m.operationDuration.WithLabelValues(accountID, operation).Observe(duration.Seconds())
}

// RecordOperationSuccess records a successful operation.
func (m *PrometheusMetrics) RecordOperationSuccess(operation, accountID string, duration time.Duration) {
	m.RecordOperation(operation, StatusSuccess, accountID, duration)
}

// RecordOperationFailure records a failed operation.
func (m *PrometheusMetrics) RecordOperationFailure(operation, accountID string, duration time.Duration) {
	m.RecordOperation(operation, StatusFailure, accountID, duration)
}

// RecordRetry records a retry event.
// Label order matches registration: accountId, operation, status.
func (m *PrometheusMetrics) RecordRetry(operation string, status OperationStatus, accountID string) {
	m.retryTotal.WithLabelValues(accountID, operation, string(status)).Inc()
}

// RecordRetryAttempt records a retry attempt.
func (m *PrometheusMetrics) RecordRetryAttempt(operation, accountID string) {
	m.RecordRetry(operation, StatusRetry, accountID)
}

// RecordRetrySuccess records a successful retry.
func (m *PrometheusMetrics) RecordRetrySuccess(operation, accountID string) {
	m.RecordRetry(operation, StatusSuccess, accountID)
}

// RecordRetryExhausted records when all retries are exhausted.
func (m *PrometheusMetrics) RecordRetryExhausted(operation, accountID string) {
	m.RecordRetry(operation, StatusExhausted, accountID)
}

// RecordIBMCloudContainerServiceCall records an IBM Cloud Container Service call.
func (m *PrometheusMetrics) RecordIBMCloudContainerServiceCall(accountID string) {
	m.ibmCloudContainerServiceCallsTotal.WithLabelValues(accountID).Inc()
}

// RecordBRSTimeout records a BRS API timeout.
func (m *PrometheusMetrics) RecordBRSTimeout(accountID string) {
	m.brsAPITimeouts.WithLabelValues(accountID).Inc()
}

// MetricInfo describes a metric.
type MetricInfo struct {
	Name     string
	FullName string
	Type     string
	Help     string
	Labels   []string
}

// ListMetrics returns all pre-registered SDK metrics.
func (m *PrometheusMetrics) ListMetrics() []MetricInfo {
	return []MetricInfo{
		{MetricOperationTotal, m.namespace + "_" + MetricOperationTotal, "counter", "Total operations completed", []string{"accountId", "operation", "status"}},
		{MetricOperationDuration, m.namespace + "_" + MetricOperationDuration, "histogram", "Operation duration in seconds", []string{"accountId", "operation"}},
		{MetricRetryTotal, m.namespace + "_" + MetricRetryTotal, "counter", "Total retry attempts", []string{"accountId", "operation", "status"}},
		{MetricIBMCloudContainerServiceCallsTotal, m.namespace + "_" + MetricIBMCloudContainerServiceCallsTotal, "counter", "Total IBM Cloud Container Service Calls", []string{"accountId"}},
		{MetricBRSAPITimeoutsTotal, m.namespace + "_" + MetricBRSAPITimeoutsTotal, "counter", "Total BRS API timeouts", []string{"accountId"}},
	}
}

// ----------------------------------------------------------------------------
// Internal helpers
// ----------------------------------------------------------------------------

func (m *PrometheusMetrics) registerCounter(name, help string, labels []string) *prometheus.CounterVec {
	cv := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: m.namespace,
		Name:      name,
		Help:      help,
	}, labels)

	if err := m.reg.Register(cv); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			cv = are.ExistingCollector.(*prometheus.CounterVec)
		} else {
			m.logger.Error(context.Background(), "failed to register counter", "error", err, "metric", name)
		}
	}

	m.mu.Lock()
	m.counters[name] = cv
	m.mu.Unlock()

	return cv
}

func (m *PrometheusMetrics) registerGauge(name, help string, labels []string) *prometheus.GaugeVec {
	gv := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: m.namespace,
		Name:      name,
		Help:      help,
	}, labels)

	if err := m.reg.Register(gv); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			gv = are.ExistingCollector.(*prometheus.GaugeVec)
		} else {
			m.logger.Error(context.Background(), "failed to register gauge", "error", err, "metric", name)
		}
	}

	m.mu.Lock()
	m.gauges[name] = gv
	m.mu.Unlock()

	return gv
}

func (m *PrometheusMetrics) registerHistogram(name, help string, buckets []float64, labels []string) *prometheus.HistogramVec {
	if buckets == nil {
		buckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
	}
	hv := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: m.namespace,
		Name:      name,
		Help:      help,
		Buckets:   buckets,
	}, labels)

	if err := m.reg.Register(hv); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			hv = are.ExistingCollector.(*prometheus.HistogramVec)
		} else {
			m.logger.Error(context.Background(), "failed to register histogram", "error", err, "metric", name)
		}
	}

	m.mu.Lock()
	m.histograms[name] = hv
	m.mu.Unlock()

	return hv
}

func (m *PrometheusMetrics) getOrCreateCounter(name string, labels []string) *prometheus.CounterVec {
	m.mu.RLock()
	cv, exists := m.counters[name]
	m.mu.RUnlock()

	if exists {
		return cv
	}

	return m.registerCounter(name, "Counter: "+name, labels)
}

func (m *PrometheusMetrics) getOrCreateGauge(name string, labels []string) *prometheus.GaugeVec {
	m.mu.RLock()
	gv, exists := m.gauges[name]
	m.mu.RUnlock()

	if exists {
		return gv
	}

	return m.registerGauge(name, "Gauge: "+name, labels)
}

func (m *PrometheusMetrics) getOrCreateHistogram(name string, labels []string) *prometheus.HistogramVec {
	m.mu.RLock()
	hv, exists := m.histograms[name]
	m.mu.RUnlock()

	if exists {
		return hv
	}

	return m.registerHistogram(name, "Histogram: "+name, prometheus.DefBuckets, labels)
}

func sortLabels(labels []Label) ([]string, []string) {
	if len(labels) == 0 {
		return nil, nil
	}

	sorted := make([]Label, len(labels))
	copy(sorted, labels)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Key < sorted[j].Key
	})

	keys := make([]string, len(sorted))
	values := make([]string, len(sorted))
	for i, l := range sorted {
		keys[i] = l.Key
		values[i] = l.Value
	}

	return keys, values
}
