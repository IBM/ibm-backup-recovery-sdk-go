package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/config"
	"github.com/IBM/ibm-backup-recovery-sdk-go/migrationv2/common/metrics"
)

func main() {
	fmt.Println("IBM BRS SDK - Metrics Example")
	fmt.Println("==============================")

	// Get required configuration from environment
	apiKey := os.Getenv("BRS_API_KEY")
	instanceCRN := os.Getenv("BRS_INSTANCE_CRN")
	resourceGroupID := os.Getenv("BRS_RESOURCE_GROUP_ID")

	if apiKey == "" || instanceCRN == "" || resourceGroupID == "" {
		log.Fatal("Error: Missing required environment variables\n" +
			"Please set: BRS_API_KEY, BRS_INSTANCE_CRN, BRS_RESOURCE_GROUP_ID")
	}

	// Create client with Prometheus metrics enabled
	cfg := config.DefaultConfig().
		WithRegion("us-south").
		WithAPIKey(apiKey).
		WithResourceGroupID(resourceGroupID).
		WithBRSInstanceCRN(instanceCRN).
		WithMetricsConfig(&metrics.PrometheusConfig{
			Namespace: "ibm_brs",
		})

	ctx := context.Background()
	client, err := migrationv2.NewClient(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close(ctx)

	fmt.Println("✓ Client initialized successfully")
	fmt.Println("✓ Metrics collection enabled")
	fmt.Println()

	// Start HTTP server for metrics endpoint
	metricsCollector := cfg.GetMetrics()
	http.Handle("/metrics", metricsCollector.Handler())

	go func() {
		fmt.Println("Starting metrics server on http://localhost:8080")
		fmt.Println("Metrics available at: http://localhost:8080/metrics")
		fmt.Println()
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(1 * time.Second)

	// Perform some operations to generate metrics
	fmt.Println("Executing operations to generate metrics...")
	fmt.Println()

	// List connections
	fmt.Print("→ Listing connections... ")
	connections, err := client.TaskAPI.ListConnections(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Found %d connections\n", len(connections))
	}

	// List registrations
	fmt.Print("→ Listing registrations... ")
	registrations, err := client.TaskAPI.ListRegistrations(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Found %d registrations\n", len(registrations))
	}

	// List policies
	fmt.Print("→ Listing policies... ")
	policies, err := client.TaskAPI.ListPolicies(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Found %d policies\n", len(policies))
	}

	fmt.Println()
	fmt.Println("✓ Operations completed")
	fmt.Println()
	fmt.Println("View metrics at: http://localhost:8080/metrics")
	fmt.Println("Press Ctrl+C to exit")

	// Keep the server running
	select {}
}

