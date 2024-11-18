package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/robfig/cron/v3"
)

type PrometheusResponse struct {
    Status string `json:"status"`
    Data   struct {
        ResultType string `json:"resultType"`
        Result     []struct {
            Metric map[string]string `json:"metric"`
            Value  []interface{}     `json:"value"`
        } `json:"result"`
    } `json:"data"`
}

func main() {
    // Create a new cron instance
    c := cron.New()

    // Add a task to the cron scheduler to run every 15 minutes
    c.AddFunc("@every 1m", func() {
        fmt.Println("Collecting metrics and calculating billing...")
        queryPrometheusAndCalculateBilling()
    })

    // Start the cron scheduler
    c.Start()

    // Keep the program running
    select {}
}

func queryPrometheusAndCalculateBilling() {
    // URLs to query Prometheus for CPU, memory, storage, and network usage metrics
    cpuQueryURL := "http://localhost:9090/api/v1/query?query=container_cpu_usage_seconds_total"
    memoryQueryURL := "http://localhost:9090/api/v1/query?query=container_memory_usage_bytes"
    storageQueryURL := "http://localhost:9090/api/v1/query?query=container_fs_usage_bytes"
    networkReceiveURL := "http://localhost:9090/api/v1/query?query=container_network_receive_bytes_total"
    networkTransmitURL := "http://localhost:9090/api/v1/query?query=container_network_transmit_bytes_total"

    // Get CPU usage
    cpuUsage, err := queryPrometheus(cpuQueryURL)
    if err != nil {
        fmt.Println("Error querying CPU usage:", err)
        return
    }

    // Get Memory usage
    memoryUsage, err := queryPrometheus(memoryQueryURL)
    if err != nil {
        fmt.Println("Error querying Memory usage:", err)
        return
    }

    // Get Storage usage
    storageUsage, err := queryPrometheus(storageQueryURL)
    if err != nil {
        fmt.Println("Error querying Storage usage:", err)
        return
    }

    // Get Network Receive usage
    networkReceiveUsage, err := queryPrometheus(networkReceiveURL)
    if err != nil {
        fmt.Println("Error querying Network Receive usage:", err)
        return
    }

    // Get Network Transmit usage
    networkTransmitUsage, err := queryPrometheus(networkTransmitURL)
    if err != nil {
        fmt.Println("Error querying Network Transmit usage:", err)
        return
    }

    // Print the result and calculate billing
    for i, cpuResult := range cpuUsage.Data.Result {
        containerName := getContainerName(cpuResult.Metric)
        if containerName == "" {
            fmt.Printf("No recognizable container name found, available labels: %+v\n", cpuResult.Metric)
            continue
        }

        // Retrieve values from all the different metrics
        cpuUsageStr := cpuResult.Value[1].(string)
        memoryUsageStr := memoryUsage.Data.Result[i].Value[1].(string)
        storageUsageStr := storageUsage.Data.Result[i].Value[1].(string)
        networkReceiveStr := networkReceiveUsage.Data.Result[i].Value[1].(string)
        networkTransmitStr := networkTransmitUsage.Data.Result[i].Value[1].(string)

        // Convert metric strings to float64
        cpuUsageValue, _ := strconv.ParseFloat(cpuUsageStr, 64)
        memoryUsageValue, _ := strconv.ParseFloat(memoryUsageStr, 64)
        storageUsageValue, _ := strconv.ParseFloat(storageUsageStr, 64)
        networkReceiveValue, _ := strconv.ParseFloat(networkReceiveStr, 64)
        networkTransmitValue, _ := strconv.ParseFloat(networkTransmitStr, 64)

        // Calculate billing
        billingAmount := calculateBilling(cpuUsageValue, memoryUsageValue, storageUsageValue, networkReceiveValue, networkTransmitValue)
        fmt.Printf("Container: %s, CPU Usage: %f seconds, Memory Usage: %f bytes, Storage Usage: %f bytes, Network Usage: %f bytes (in/out), Billing Amount: $%.2f\n", containerName, cpuUsageValue, memoryUsageValue, storageUsageValue, networkReceiveValue+networkTransmitValue, billingAmount)
    }
}

func queryPrometheus(queryURL string) (PrometheusResponse, error) {
    // Making the HTTP GET request
    resp, err := http.Get(queryURL)
    if err != nil {
        return PrometheusResponse{}, err
    }
    defer resp.Body.Close()

    // Reading response body
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return PrometheusResponse{}, err
    }

    // Unmarshal JSON response
    var promResponse PrometheusResponse
    err = json.Unmarshal(body, &promResponse)
    if err != nil {
        return PrometheusResponse{}, err
    }

    return promResponse, nil
}

func getContainerName(metric map[string]string) string {
    // Try multiple potential keys for container identification
    if name, exists := metric["container"]; exists {
        return name
    }
    if name, exists := metric["container_name"]; exists {
        return name
    }
    if name, exists := metric["id"]; exists {
        return name
    }
    return ""
}

func calculateBilling(cpuSeconds float64, memoryBytes float64, storageBytes float64, networkReceiveBytes float64, networkTransmitBytes float64) float64 {
    // Define your pricing rates
    cpuRate := 0.02            // $0.02 per vCPU per hour
    memoryRate := 0.01         // $0.01 per GB per hour
    storageRate := 0.005       // $0.005 per GB per hour
    networkRate := 0.001       // $0.001 per GB transferred

    // Convert memory, storage, and network from bytes to gigabytes
    memoryGB := memoryBytes / (1024 * 1024 * 1024)
    storageGB := storageBytes / (1024 * 1024 * 1024)
    networkGB := (networkReceiveBytes + networkTransmitBytes) / (1024 * 1024 * 1024)

    // Calculate billing amount
    cpuBilling := (cpuSeconds / 3600) * cpuRate // Convert seconds to hours
    memoryBilling := memoryGB * memoryRate
    storageBilling := storageGB * storageRate
    networkBilling := networkGB * networkRate

    return cpuBilling + memoryBilling + storageBilling + networkBilling
}
