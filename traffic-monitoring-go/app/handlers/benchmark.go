package handlers

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"traffic-monitoring-go/app/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BenchmarkHandler struct {
	DB        *gorm.DB
	StartTime time.Time
}

func NewBenchmarkHandler(db *gorm.DB) *BenchmarkHandler {
	return &BenchmarkHandler{
		DB:        db,
		StartTime: time.Now(),
	}
}

// Get real time metrics
func (h *BenchmarkHandler) GetRealTimeMetrics(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Get actual database counts
	var totalEvents, totalAlerts, totalRules, totalV2XMessages int64
	h.DB.Model(&models.SecurityEvent{}).Count(&totalEvents)
	h.DB.Model(&models.Alert{}).Count(&totalAlerts)
	h.DB.Model(&models.Rule{}).Count(&totalRules)
	h.DB.Model(&models.V2XMessage{}).Count(&totalV2XMessages)

	// calculate recent activity
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	var recentEvents, recentAlerts, recentV2X int64
	h.DB.Model(&models.SecurityEvent{}).Where("created_at > ?", fiveMinutesAgo).Count(&recentEvents)
	h.DB.Model(&models.Alert{}).Where("created_at > ?", fiveMinutesAgo).Count(&recentAlerts)
	h.DB.Model(&models.V2XMessage{}).Where("created_at > ?", fiveMinutesAgo).Count(&recentV2X)

	// Calculate throughput
	uptimeMinutes := time.Since(h.StartTime).Minutes()
	eventsPerMinute := float64(totalEvents) / uptimeMinutes
	alertsPerMinute := float64(totalAlerts) / uptimeMinutes

	metrics := map[string]interface{}{
		"system_resources": map[string]interface{}{
			"memory_allocated_mb": bToMb(m.Alloc),
			"memory_total_mb":     bToMb(m.Sys),
			"memory_heap_mb":      bToMb(m.HeapAlloc),
			"gc_cycles":           m.NumGC,
			"goroutines_active":   runtime.NumGoroutine(),
		},
		"application_metrics": map[string]interface{}{
			"uptime_minutes":        uptimeMinutes,
			"total_events":          totalEvents,
			"total_alerts":          totalAlerts,
			"total_rules":           totalRules,
			"total_v2x_messages":    totalV2XMessages,
			"events_per_minute_avg": eventsPerMinute,
			"alerts_per_minute_avg": alertsPerMinute,
		},
		"recent_activity_5min": map[string]interface{}{
			"events_processed":  recentEvents,
			"alerts_generated":  recentAlerts,
			"v2x_messages":      recentV2X,
			"events_per_second": float64(recentEvents) / 300.0,
			"alerts_per_second": float64(recentAlerts) / 300.0,
		},
	}

	c.JSON(http.StatusOK, metrics)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// GetPerformanceBenchmark handles GET /benchmark/performance
func (h *BenchmarkHandler) GetPerformanceBenchmark(c *gin.Context) {
	start := time.Now()

	// Test 1: Database query performance
	dbStartTime := time.Now()
	var events []models.SecurityEvent
	h.DB.Limit(100).Find(&events)
	dbQueryTime := time.Since(dbStartTime)

	// Test 2: Rule query performance
	ruleStartTime := time.Now()
	var rules []models.Rule
	h.DB.Where("status = ?", models.RuleStatusEnabled).Find(&rules)
	ruleQueryTime := time.Since(ruleStartTime)

	// Test 3: Complex join query performance
	joinStartTime := time.Now()
	var alerts []models.Alert
	h.DB.Preload("Rule").Preload("SecurityEvent").Limit(50).Find(&alerts)
	joinQueryTime := time.Since(joinStartTime)

	// Test 4: Aggregation query performance
	aggStartTime := time.Now()
	var severityCounts []struct {
		Severity string
		Count    int64
	}
	h.DB.Model(&models.SecurityEvent{}).
		Select("severity, count(*) as count").
		Group("severity").
		Find(&severityCounts)
	aggQueryTime := time.Since(aggStartTime)

	// Get rule evaluation efficiency
	var alertRuleRatio float64
	var totalEvents, totalAlerts int64
	h.DB.Model(&models.SecurityEvent{}).Count(&totalEvents)
	h.DB.Model(&models.Alert{}).Count(&totalAlerts)

	if totalEvents > 0 {
		alertRuleRatio = float64(totalAlerts) / float64(totalEvents) * 100
	}

	totalBenchmarkTime := time.Since(start)

	benchmark := map[string]interface{}{
		"database_performance": map[string]interface{}{
			"simple_query_ms":    dbQueryTime.Milliseconds(),
			"rule_query_ms":      ruleQueryTime.Milliseconds(),
			"join_query_ms":      joinQueryTime.Milliseconds(),
			"aggregation_ms":     aggQueryTime.Milliseconds(),
			"total_benchmark_ms": totalBenchmarkTime.Milliseconds(),
		},
		"rule_engine_efficiency": map[string]interface{}{
			"total_events":          totalEvents,
			"total_alerts":          totalAlerts,
			"alert_generation_rate": alertRuleRatio,
			"active_rules":          len(rules),
		},
		"severity_distribution": severityCounts,
		"performance_rating": map[string]interface{}{
			"query_performance":  getPerformanceRating(dbQueryTime.Milliseconds()),
			"rule_performance":   getPerformanceRating(ruleQueryTime.Milliseconds()),
			"overall_efficiency": getEfficiencyRating(alertRuleRatio),
		},
	}

	c.JSON(http.StatusOK, benchmark)
}

func getPerformanceRating(ms int64) string {
	if ms < 10 {
		return "excellent"
	} else if ms < 50 {
		return "good"
	} else if ms < 100 {
		return "acceptable"
	}
	return "needs_optimization"
}

func getEfficiencyRating(ratio float64) string {
	if ratio < 1 {
		return "low_alert_noise"
	} else if ratio < 5 {
		return "balanced"
	} else if ratio < 15 {
		return "high_sensitivity"
	}
	return "potential_alert_fatigue"
}

// Stress test the app
func (h *BenchmarkHandler) RunStressTest(c *gin.Context) {
	// test parameters from request
	var params struct {
		EventCount     int `json:"event_count" default:"100"`
		ConcurrentReqs int `json:"concurrent_requests" default:"10"`
		Duration       int `json:"duration_seconds" default:"30"`
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		// use defaults if no body is provided in request
		params.EventCount = 100
		params.ConcurrentReqs = 10
		params.Duration = 30
	}

	// validate params
	if params.EventCount > 10000 {
		params.EventCount = 10000
	}
	if params.ConcurrentReqs > 50 {
		params.ConcurrentReqs = 50
	}
	if params.Duration > 120 {
		params.Duration = 120
	}

	startTime := time.Now()
	var totalLatency time.Duration
	var successCount, errorCount int64
	var minLatency, maxLatency time.Duration = time.Hour, 0

	// channel to collect results
	results := make(chan struct {
		latency time.Duration
		success bool
	}, params.EventCount)

	// record initial memory
	var initialMem runtime.MemStats
	runtime.ReadMemStats(&initialMem)

	// generate concurrent load
	semaphore := make(chan struct{}, params.ConcurrentReqs)

	for i := 0; i < params.EventCount; i++ {
		go func(index int) {
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			reqStart := time.Now()

			// Single optimized insert - no ES during stress test
			event := models.SecurityEvent{
				Timestamp:   time.Now(),
				SourceIP:    fmt.Sprintf("192.168.1.%d", index%254+1),
				Severity:    models.SeverityInfo,
				Category:    models.CategorySystem,
				Message:     fmt.Sprintf("Stress test event %d", index),
				LogSourceID: 1,
				RawData:     "", // Minimal data
			}

			// Use transaction for speed
			err := h.DB.Transaction(func(tx *gorm.DB) error {
				return tx.Create(&event).Error
			})

			latency := time.Since(reqStart)

			results <- struct {
				latency time.Duration
				success bool
			}{latency, err == nil}
		}(i)
	}

	// collect results
	for i := 0; i < params.EventCount; i++ {
		result := <-results
		totalLatency += result.latency

		if result.success {
			successCount++
		} else {
			errorCount++
		}

		if result.latency < minLatency {
			minLatency = result.latency
		}
		if result.latency > maxLatency {
			maxLatency = result.latency
		}
	}

	totalDuration := time.Since(startTime)

	// record final memory
	var finalMem runtime.MemStats
	runtime.ReadMemStats(&finalMem)

	// calculate metrics
	avgLatency := totalLatency / time.Duration(params.EventCount)
	throughputPerSec := float64(successCount) / totalDuration.Seconds()
	successRate := float64(successCount) / float64(params.EventCount) * 100
	memoryIncrease := bToMb(finalMem.Alloc - initialMem.Alloc)

	stressResults := map[string]interface{}{
		"test_parameters": map[string]interface{}{
			"events_generated":    params.EventCount,
			"concurrent_requests": params.ConcurrentReqs,
			"total_duration_ms":   totalDuration.Milliseconds(),
		},
		"performance_metrics": map[string]interface{}{
			"throughput_events_per_sec": throughputPerSec,
			"avg_latency_ms":            avgLatency.Milliseconds(),
			"min_latency_ms":            minLatency.Milliseconds(),
			"max_latency_ms":            maxLatency.Milliseconds(),
			"success_rate_percent":      successRate,
			"total_requests":            successCount + errorCount,
			"successful_requests":       successCount,
			"failed_requests":           errorCount,
		},
		"resource_impact": map[string]interface{}{
			"memory_increase_mb":    memoryIncrease,
			"gc_cycles_during_test": finalMem.NumGC - initialMem.NumGC,
			"final_goroutines":      runtime.NumGoroutine(),
		},
		"performance_rating": map[string]interface{}{
			"latency_rating":    getLatencyRating(avgLatency.Milliseconds()),
			"throughput_rating": getThroughputRating(throughputPerSec),
			"stability_rating":  getStabilityRating(successRate),
		},
	}

	c.JSON(http.StatusOK, stressResults)
}

func getLatencyRating(avgMs int64) string {
	if avgMs < 5 {
		return "excellent"
	}
	if avgMs < 20 {
		return "good"
	}
	if avgMs < 100 {
		return "acceptable"
	}
	return "poor"
}

func getThroughputRating(tps float64) string {
	if tps > 100 {
		return "excellent"
	}
	if tps > 50 {
		return "good"
	}
	if tps > 20 {
		return "acceptable"
	}
	return "poor"
}

func getStabilityRating(successRate float64) string {
	if successRate >= 99 {
		return "excellent"
	}
	if successRate >= 95 {
		return "good"
	}
	if successRate >= 90 {
		return "acceptable"
	}
	return "poor"
}
