package brutespray

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pterm/pterm"
	"github.com/x90skysn3k/brutespray/banner"
	"github.com/x90skysn3k/brutespray/brute"
	"github.com/x90skysn3k/brutespray/modules"
	"github.com/x90skysn3k/brutespray/fingerprint"
)

// [Previous constants and struct definitions remain the same...]

// Enhanced HostWorkerPool with stealth capabilities
type HostWorkerPool struct {
	host           modules.Host
	workers        int
	jobQueue       chan Credential
	progressCh     chan int
	wg             sync.WaitGroup
	stopChan       chan struct{}
	// Performance tracking
	avgResponseTime time.Duration
	successRate     float64
	totalAttempts   int64
	mutex           sync.RWMutex
	// Stealth features
	requestDelay    time.Duration
	randomDelay     bool
	jitterFactor    float64 // 0.0 to 1.0 for percentage jitter
	// Service fingerprint
	serviceInfo     fingerprint.ServiceInfo
}

// Enhanced WorkerPool with intelligence capabilities
type WorkerPool struct {
	globalWorkers     int
	threadsPerHost    int
	hostPools         map[string]*HostWorkerPool
	hostPoolsMutex    sync.RWMutex
	progressCh        chan int
	globalStopChan    chan struct{}
	hostParallelism   int
	hostSem           chan struct{}
	// Dynamic thread allocation
	dynamicAllocation bool
	minThreadsPerHost int
	maxThreadsPerHost int
	// Statistics control
	noStats bool
	// Intelligence features
	fingerprintEnabled bool
	// Adaptive attack
	adaptiveMode      bool
}

// NewHostWorkerPool creates a new host-specific worker pool with stealth options
func NewHostWorkerPool(host modules.Host, workers int, progressCh chan int, 
	requestDelay time.Duration, randomDelay bool, jitterFactor float64) *HostWorkerPool {
	
	return &HostWorkerPool{
		host:          host,
		workers:       workers,
		jobQueue:      make(chan Credential, workers*10),
		progressCh:    progressCh,
		stopChan:      make(chan struct{}),
		requestDelay:  requestDelay,
		randomDelay:   randomDelay,
		jitterFactor:  jitterFactor,
	}
}

// NewWorkerPool creates a new worker pool with enhanced capabilities
func NewWorkerPool(threadsPerHost int, progressCh chan int, hostParallelism int, 
	hostCount int, fingerprintEnabled bool, adaptiveMode bool) *WorkerPool {
	
	totalWorkers := threadsPerHost * hostCount

	return &WorkerPool{
		globalWorkers:      totalWorkers,
		threadsPerHost:     threadsPerHost,
		hostPools:          make(map[string]*HostWorkerPool),
		progressCh:         progressCh,
		globalStopChan:     make(chan struct{}),
		hostParallelism:    hostParallelism,
		hostSem:            make(chan struct{}, hostParallelism),
		dynamicAllocation:  adaptiveMode,
		minThreadsPerHost:  1,
		maxThreadsPerHost:  threadsPerHost * 2, // Allow scaling up
		fingerprintEnabled: fingerprintEnabled,
		adaptiveMode:       adaptiveMode,
	}
}

// Enhanced worker with stealth capabilities
func (hwp *HostWorkerPool) worker(timeout time.Duration, retry int, output string, 
	socksProxy string, netInterface string, domain string, noStats bool) {
	
	defer hwp.wg.Done()

	for {
		select {
		case <-hwp.stopChan:
			return
		case cred, ok := <-hwp.jobQueue:
			if !ok {
				return
			}

			// Apply stealth delay if configured
			if hwp.requestDelay > 0 {
				delay := hwp.requestDelay
				if hwp.randomDelay {
					// Add jitter to delay
					jitter := time.Duration(float64(delay) * hwp.jitterFactor * rand.Float64())
					if rand.Intn(2) == 0 {
						delay += jitter
					} else {
						delay -= jitter
					}
					if delay < 0 {
						delay = 0
					}
				}
				time.Sleep(delay)
			}

			startTime := time.Now()

			// Use service-specific brute force if fingerprint info is available
			var success bool
			if hwp.serviceInfo.Version != "" && hwp.adaptiveMode {
				success = brute.AdaptiveBrute(cred.Host, cred.User, cred.Password, 
					hwp.progressCh, timeout, retry, output, socksProxy, 
					netInterface, domain, hwp.serviceInfo)
			} else {
				success = brute.RunBrute(cred.Host, cred.User, cred.Password, 
					hwp.progressCh, timeout, retry, output, socksProxy, 
					netInterface, domain)
			}

			duration := time.Since(startTime)
			if !noStats {
				if !success {
					modules.RecordConnectionError(cred.Host.Host)
				}
			}

			hwp.updatePerformanceMetrics(success, duration)
			hwp.progressCh <- 1
		}
	}
}

// Enhanced ProcessHost with service fingerprinting
func (wp *WorkerPool) ProcessHost(host modules.Host, service string, combo string, 
	user string, password string, version string, timeout time.Duration, 
	retry int, output string, socksProxy string, netInterface string, 
	domain string, requestDelay time.Duration, randomDelay bool, jitterFactor float64) {
	
	// [Previous host processing logic...]

	// Service fingerprinting before starting attacks
	var serviceInfo fingerprint.ServiceInfo
	if wp.fingerprintEnabled {
		if !NoColorMode {
			modules.PrintfColored(pterm.FgLightBlue, "[*] Fingerprinting service on %s:%d...\n", 
				host.Host, host.Port)
		}
		
		serviceInfo = fingerprint.FingerprintService(host, timeout)
		
		if !NoColorMode {
			if serviceInfo.Version != "" {
				modules.PrintfColored(pterm.FgLightGreen, 
					"[*] Detected: %s %s on %s:%d\n", 
					serviceInfo.Name, serviceInfo.Version, host.Host, host.Port)
			} else {
				modules.PrintfColored(pterm.FgLightYellow, 
					"[*] Service detected but version unknown on %s:%d\n", 
					host.Host, host.Port)
			}
		}
	}

	// Get or create host-specific worker pool with stealth options
	hostPool := wp.getOrCreateHostPool(host, requestDelay, randomDelay, jitterFactor)
	hostPool.serviceInfo = serviceInfo

	// [Rest of the host processing logic...]
}

// Enhanced Execute function with new options
func Execute() {
	// [Previous flag definitions...]
	
	// New stealth and intelligence flags
	requestDelay := flag.Duration("delay", 0, "Delay between requests to avoid detection")
	randomDelay := flag.Bool("random-delay", false, "Add random jitter to delay times")
	jitterFactor := flag.Float64("jitter", 0.3, "Jitter factor (0.0-1.0) for random delays")
	fingerprint := flag.Bool("fingerprint", true, "Fingerprint services before attacking")
	adaptive := flag.Bool("adaptive", true, "Use adaptive attacks based on service fingerprinting")
	rateLimit := flag.Int("rate-limit", 0, "Maximum requests per minute per host (0 for unlimited)")
	userAgent := flag.String("user-agent", "", "Custom user agent for HTTP services")
	stealthMode := flag.Bool("stealth", false, "Enable stealth mode (randomized delays, etc.)")
	
	flag.Parse()

	// Set stealth options if stealth mode is enabled
	if *stealthMode {
		if *requestDelay == 0 {
			*requestDelay = 100 * time.Millisecond
		}
		*randomDelay = true
		*fingerprint = true
		*adaptive = true
	}

	// [Previous initialization code...]

	// Create enhanced worker pool with intelligence features
	progressCh := make(chan int, (*threads)*totalHosts*10)
	workerPool := NewWorkerPool(*threads, progressCh, *hostParallelism, totalHosts, 
		*fingerprint, *adaptive)

	// [Rest of the execution logic with enhanced host processing...]
	
	// Process hosts with enhanced capabilities
	for _, service := range supportedServices {
		for _, h := range hostsList {
			if h.Service == service {
				hostWg.Add(1)
				go func(host modules.Host, svc string) {
					defer hostWg.Done()
					select {
					case <-workerPool.globalStopChan:
						return
					default:
						workerPool.ProcessHost(host, svc, *combo, *user, *password, 
							version, *timeout, *retry, *output, *socksProxy, 
							*netInterface, *domain, *requestDelay, *randomDelay, 
							*jitterFactor)
					}
				}(h, service)
			}
		}
	}

	
}
