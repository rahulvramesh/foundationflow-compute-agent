package main

import (
	"bytes"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

var (
	db            *sql.DB
	reportingURL  string
	authToken     string
	reportingFreq int
)

type SystemInfo struct {
	Timestamp   time.Time
	MemoryInfo  MemoryDetails
	SwapInfo    SwapDetails
	StorageInfo []StorageDetails
	CPUInfo     CPUDetails
	GPUUsage    float64
	LscpuJSON   map[string]interface{}
}

type MemoryDetails struct {
	Total     uint64
	Available uint64
	Used      uint64
}

type SwapDetails struct {
	Total uint64
	Used  uint64
	Free  uint64
}

type StorageDetails struct {
	Device string
	Total  uint64
	Used   uint64
	Free   uint64
}

type CPUDetails struct {
	TotalCores   int32
	UsagePerCore []float64
}

func main() {
	flag.StringVar(&reportingURL, "url", "", "URL to send reports to")
	flag.StringVar(&authToken, "token", "", "Authentication token")
	flag.IntVar(&reportingFreq, "freq", 5, "Reporting frequency in minutes")
	flag.Parse()

	if reportingURL == "" || authToken == "" {
		log.Fatal("Reporting URL and authentication token are required")
	}

	var err error
	db, err = sql.Open("sqlite3", "./server_monitor.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable()

	for {
		info := getSystemInfo()
		storeSystemInfo(info)
		sendReport(info)
		time.Sleep(time.Duration(reportingFreq) * time.Minute)
	}
}

func createTable() {
	query := `
	CREATE TABLE IF NOT EXISTS system_stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		memory_info TEXT,
		swap_info TEXT,
		storage_info TEXT,
		cpu_info TEXT,
		gpu_usage REAL,
		lscpu_json TEXT
	);`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func getSystemInfo() SystemInfo {
	return SystemInfo{
		Timestamp:   time.Now(),
		MemoryInfo:  getMemoryInfo(),
		SwapInfo:    getSwapInfo(),
		StorageInfo: getStorageInfo(),
		CPUInfo:     getCPUInfo(),
		GPUUsage:    getGPUUsage(),
		LscpuJSON:   getLscpuJSON(),
	}
}

func getMemoryInfo() MemoryDetails {
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("Error getting memory stats: %v", err)
		return MemoryDetails{}
	}
	return MemoryDetails{
		Total:     vmStat.Total,
		Available: vmStat.Available,
		Used:      vmStat.Used,
	}
}

func getSwapInfo() SwapDetails {
	swapStat, err := mem.SwapMemory()
	if err != nil {
		log.Printf("Error getting swap stats: %v", err)
		return SwapDetails{}
	}
	return SwapDetails{
		Total: swapStat.Total,
		Used:  swapStat.Used,
		Free:  swapStat.Free,
	}
}

func getStorageInfo() []StorageDetails {
	var storageDetails []StorageDetails
	partitions, err := disk.Partitions(false)
	if err != nil {
		log.Printf("Error getting disk partitions: %v", err)
		return storageDetails
	}

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			log.Printf("Error getting disk usage for %s: %v", partition.Mountpoint, err)
			continue
		}
		storageDetails = append(storageDetails, StorageDetails{
			Device: partition.Device,
			Total:  usage.Total,
			Used:   usage.Used,
			Free:   usage.Free,
		})
	}
	return storageDetails
}

func getCPUInfo() CPUDetails {
	cores, err := cpu.Counts(true)
	if err != nil {
		log.Printf("Error getting CPU core count: %v", err)
		return CPUDetails{}
	}

	perCPU, err := cpu.Percent(time.Second, true)
	if err != nil {
		log.Printf("Error getting CPU usage: %v", err)
		return CPUDetails{TotalCores: int32(cores)}
	}

	return CPUDetails{
		TotalCores:   int32(cores),
		UsagePerCore: perCPU,
	}
}

func getGPUUsage() float64 {
	if err := nvml.Init(); err != nil {
		log.Printf("Failed to initialize NVML: %v", err)
		return -1
	}
	defer nvml.Shutdown()

	count, err := nvml.GetDeviceCount()
	if err != nil {
		log.Printf("Error getting GPU count: %v", err)
		return -1
	}

	if count == 0 {
		return -1 // No GPUs found
	}

	var totalUsage uint
	for i := uint(0); i < count; i++ {
		device, err := nvml.NewDevice(i)
		if err != nil {
			log.Printf("Error getting GPU device: %v", err)
			continue
		}
		status, err := device.Status()
		if err != nil {
			log.Printf("Error getting GPU status: %v", err)
			continue
		}
		if status.Utilization.GPU != nil {
			totalUsage += *status.Utilization.GPU
		}
	}

	return float64(totalUsage) / float64(count)
}

func getLscpuJSON() map[string]interface{} {
	cmd := exec.Command("lscpu", "--json")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error executing lscpu command: %v", err)
		return nil
	}

	var result map[string]interface{}
	err = json.Unmarshal(output, &result)
	if err != nil {
		log.Printf("Error parsing lscpu JSON output: %v", err)
		return nil
	}

	return result
}

func storeSystemInfo(info SystemInfo) {
	memoryInfoJSON, _ := json.Marshal(info.MemoryInfo)
	swapInfoJSON, _ := json.Marshal(info.SwapInfo)
	storageInfoJSON, _ := json.Marshal(info.StorageInfo)
	cpuInfoJSON, _ := json.Marshal(info.CPUInfo)
	lscpuJSON, _ := json.Marshal(info.LscpuJSON)

	_, err := db.Exec(`INSERT INTO system_stats 
		(memory_info, swap_info, storage_info, cpu_info, gpu_usage, lscpu_json) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		string(memoryInfoJSON), string(swapInfoJSON), string(storageInfoJSON),
		string(cpuInfoJSON), info.GPUUsage, string(lscpuJSON))
	if err != nil {
		log.Printf("Error inserting into database: %v", err)
	}
}

func sendReport(info SystemInfo) {
	jsonData, err := json.Marshal(info)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return
	}

	// Create a custom HTTP client with TLS configuration
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("POST", reportingURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending report: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Unexpected status code: %d", resp.StatusCode)
	} else {
		log.Println("Report sent successfully")
	}
}
