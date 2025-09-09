package main

import (
	"flag"
	"fmt"
	"math"
	"net/http"
	"sort"
	"sync"
	"time"
)

type TestConfig struct {
	URL         string
	Requests    int
	Concurrency int
}

type TestResults struct {
	TotalTime     time.Duration
	Success       int
	Failures      int
	AllResponses  []time.Duration
	Sum           time.Duration
}

type ColdStartAnalysis struct {
	Count     int
	Sum       time.Duration
	Threshold time.Duration
	Median    time.Duration
	StdDev    time.Duration
}

func parseFlags() TestConfig {
	url := flag.String("url", "http://localhost:8080", "API endpoint to test")
	requests := flag.Int("n", 100, "Number of requests to send")
	concurrency := flag.Int("c", 10, "Number of concurrent workers")
	flag.Parse()

	return TestConfig{
		URL:         *url,
		Requests:    *requests,
		Concurrency: *concurrency,
	}
}

func runLoadTest(config TestConfig) TestResults {
	var wg sync.WaitGroup
	jobs := make(chan int, config.Requests)
	results := make(chan time.Duration, config.Requests)
	errors := make(chan error, config.Requests)
	allResponses := make([]time.Duration, 0, config.Requests)

	start := time.Now()

	worker := func(id int) {
		defer wg.Done()
		for job := range jobs {
			fmt.Printf("Worker %d sending request #%d to %s\n", id, job+1, config.URL)
			begin := time.Now()
			resp, err := http.Get(config.URL)
			if err != nil {
				errors <- err
				continue
			}
			resp.Body.Close()
			duration := time.Since(begin)
			results <- duration
		}
	}

	// Start workers
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go worker(i)
	}

	// Send jobs
	for i := 0; i < config.Requests; i++ {
		jobs <- i
	}
	close(jobs)

	// Wait for completion
	wg.Wait()
	close(results)
	close(errors)
	totalTime := time.Since(start)

	// Collect results
	success := 0
	failures := 0
	var sum time.Duration
	for d := range results {
		sum += d
		success++
		allResponses = append(allResponses, d)
	}
	for range errors {
		failures++
	}

	return TestResults{
		TotalTime:    totalTime,
		Success:      success,
		Failures:     failures,
		AllResponses: allResponses,
		Sum:          sum,
	}
}

func analyzeColdStarts(results TestResults) ColdStartAnalysis {
	if results.Success == 0 {
		return ColdStartAnalysis{}
	}

	// Sort responses to calculate median
	sortedResponses := make([]time.Duration, len(results.AllResponses))
	copy(sortedResponses, results.AllResponses)
	sort.Slice(sortedResponses, func(i, j int) bool {
		return sortedResponses[i] < sortedResponses[j]
	})

	median := sortedResponses[results.Success/2]

	// Calculate mean
	meanNanos := float64(results.Sum.Nanoseconds()) / float64(results.Success)

	// Calculate standard deviation
	var variance float64
	for _, d := range results.AllResponses {
		diff := float64(d.Nanoseconds()) - meanNanos
		variance += diff * diff
	}
	variance /= float64(results.Success)
	stdDev := time.Duration(math.Sqrt(variance))

	// Cold start threshold: median + 2*stddev (captures ~95% of normal responses)
	threshold := median + 2*stdDev

	// Alternative: use 3x median as threshold (simpler approach)
	simpleThreshold := time.Duration(float64(median) * 3)

	// Use the more conservative threshold
	finalThreshold := threshold
	if simpleThreshold > threshold {
		finalThreshold = simpleThreshold
	}

	// Identify cold starts
	coldStartCount := 0
	var coldStartSum time.Duration
	for _, d := range results.AllResponses {
		if d > finalThreshold {
			coldStartSum += d
			coldStartCount++
		}
	}

	fmt.Printf("Debug: Median=%v, StdDev=%v, Threshold=%v\n", median, stdDev, finalThreshold)

	return ColdStartAnalysis{
		Count:     coldStartCount,
		Sum:       coldStartSum,
		Threshold: finalThreshold,
		Median:    median,
		StdDev:    stdDev,
	}
}

func printReport(config TestConfig, results TestResults, coldStart ColdStartAnalysis) {
	// Color codes
	const (
		colorReset  = "\033[0m"
		colorRed    = "\033[31m"
		colorGreen  = "\033[32m"
		colorYellow = "\033[33m"
		colorBlue   = "\033[34m"
		colorPurple = "\033[35m"
		colorCyan   = "\033[36m"
		colorWhite  = "\033[37m"
		colorBold   = "\033[1m"
	)

	// Calculate averages
	avg := time.Duration(0)
	if results.Success > 0 {
		avg = results.Sum / time.Duration(results.Success)
	}

	avgColdStart := time.Duration(0)
	if coldStart.Count > 0 {
		avgColdStart = coldStart.Sum / time.Duration(coldStart.Count)
	}

	// Print professional report
	fmt.Printf("\n%s%s════════════════════════════════════════%s\n", colorBold, colorCyan, colorReset)
	fmt.Printf("%s%s           LOAD TEST REPORT%s\n", colorBold, colorCyan, colorReset)
	fmt.Printf("%s%s════════════════════════════════════════%s\n", colorBold, colorCyan, colorReset)

	fmt.Printf("\n%s%sTest Configuration:%s\n", colorBold, colorBlue, colorReset)
	fmt.Printf("  Target URL:      %s\n", config.URL)
	fmt.Printf("  Total Requests:  %d\n", config.Requests)
	fmt.Printf("  Concurrency:     %d\n", config.Concurrency)

	fmt.Printf("\n%s%sExecution Results:%s\n", colorBold, colorBlue, colorReset)

	// Success rate with color
	successRate := float64(results.Success) / float64(config.Requests) * 100
	successColor := colorGreen
	if successRate < 95 {
		successColor = colorYellow
	}
	if successRate < 90 {
		successColor = colorRed
	}

	fmt.Printf("  %sSuccessful:      %d (%.1f%%)%s\n", successColor, results.Success, successRate, colorReset)
	if results.Failures > 0 {
		fmt.Printf("  %sFailed:          %d (%.1f%%)%s\n", colorRed, results.Failures, float64(results.Failures)/float64(config.Requests)*100, colorReset)
	} else {
		fmt.Printf("  %sFailed:          %d%s\n", colorGreen, results.Failures, colorReset)
	}

	fmt.Printf("\n%s%sPerformance Metrics:%s\n", colorBold, colorBlue, colorReset)
	fmt.Printf("  Total Duration:  %s%v%s\n", colorWhite, results.TotalTime, colorReset)
	fmt.Printf("  Average Response: %s%v%s\n", colorWhite, avg, colorReset)

	// Cold start analysis with color
	fmt.Printf("\n%s%sCold Start Analysis:%s\n", colorBold, colorBlue, colorReset)
	if coldStart.Count > 0 {
		coldStartRate := float64(coldStart.Count) / float64(results.Success) * 100
		coldStartColor := colorYellow
		if coldStartRate > 20 {
			coldStartColor = colorRed
		}
		fmt.Printf("  %sCold Starts:     %d (%.1f%% of successful requests)%s\n", coldStartColor, coldStart.Count, coldStartRate, colorReset)
		fmt.Printf("  %sAvg Cold Start:  %v%s\n", coldStartColor, avgColdStart, colorReset)
	} else {
		fmt.Printf("  %sCold Starts:     None detected%s\n", colorGreen, colorReset)
		fmt.Printf("  %sAvg Cold Start:  N/A%s\n", colorGreen, colorReset)
	}

	fmt.Printf("\n%s%s════════════════════════════════════════%s\n", colorBold, colorCyan, colorReset)
}

func main() {
	config := parseFlags()
	results := runLoadTest(config)
	coldStart := analyzeColdStarts(results)
	printReport(config, results, coldStart)
}
