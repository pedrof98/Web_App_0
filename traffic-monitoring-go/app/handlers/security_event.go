package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"traffic-monitoring-go/app/models"
	"traffic-monitoring-go/app/siem/elasticsearch"
)

// SecurityEventHandler handles security event-related endpoints
type SecurityEventHandler struct {
	DB        *gorm.DB
	ESService *elasticsearch.Service
}

// NewSecurityEventHandler creates a new SecurityEventHandler
func NewSecurityEventHandler(db *gorm.DB, esService *elasticsearch.Service) *SecurityEventHandler {
	return &SecurityEventHandler{
		DB:        db,
		ESService: esService,
	}
}

// GetSecurityEvents handles GET /security-events
func (h *SecurityEventHandler) GetSecurityEvents(c *gin.Context) {
	var events []models.SecurityEvent

	// Basic pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
	offset := (page - 1) * pageSize

	// Basic filtering by severity and category
	severity := c.Query("severity")
	category := c.Query("category")

	// Create a query builder
	query := h.DB.Model(&models.SecurityEvent{})

	if severity != "" {
		query = query.Where("severity = ?", severity)
	}

	if category != "" {
		query = query.Where("category = ?", category)
	}

	// Order by timestamp descending (most recent first)
	query = query.Order("timestamp DESC")

	// Count total for pagination info
	var total int64
	query.Count(&total)

	// Execute the query with pagination
	if err := query.Offset(offset).Limit(pageSize).Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": events,
		"pagination": gin.H{
			"page":     page,
			"pageSize": pageSize,
			"total":    total,
			"pages":    (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// GetSecurityEvent handles GET /security-events/:id
func (h *SecurityEventHandler) GetSecurityEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var event models.SecurityEvent
	if err := h.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Security event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// CreateSecurityEvent handles POST /security-events
func (h *SecurityEventHandler) CreateSecurityEvent(c *gin.Context) {
	var event models.SecurityEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Save to database
	if err := h.DB.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Index in Elasticsearch if available
	if h.ESService != nil {
		if err := h.ESService.IndexSecurityEvent(&event); err != nil {
			// Log the error but don't fail the request
			// The event is already in the database
			c.JSON(http.StatusCreated, gin.H{
				"event": event,
				"warning": "Event created in database but could not be indexed in Elasticsearch: " + err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusCreated, event)
}

// CreateBatchSecurityEvents handles POST /security-events/batch
func (h *SecurityEventHandler) CreateBatchSecurityEvents(c *gin.Context) {
	var events []models.SecurityEvent
	if err := c.ShouldBindJSON(&events); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use a transaction for batch insert
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		for i := range events {
			if err := tx.Create(&events[i]).Error; err != nil {
				return err
			}
			
			// Index in Elasticsearch if available
			if h.ESService != nil {
				if err := h.ESService.IndexSecurityEvent(&events[i]); err != nil {
					// Log the error but continue with other events
					// We don't want to fail the entire batch
					c.Error(err) // Add to Gin's error list
				}
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if there were any Elasticsearch indexing errors
	if len(c.Errors) > 0 {
		c.JSON(http.StatusCreated, gin.H{
			"message": "Batch security events created with some Elasticsearch indexing errors",
			"count": len(events),
			"warnings": c.Errors.Errors(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Batch security events created successfully",
		"count": len(events),
	})
}

// SearchSecurityEvents handles GET /security-events/search
func (h *SecurityEventHandler) SearchSecurityEvents(c *gin.Context) {
	// Check if Elasticsearch is available
	if h.ESService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Elasticsearch service not available"})
		return
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	// Build query from query parameters
	var query map[string]interface{}
	
	// If a raw query is provided, use it
	rawQuery := c.Query("query")
	if rawQuery != "" {
		if err := json.Unmarshal([]byte(rawQuery), &query); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query JSON: " + err.Error()})
			return
		}
	} else {
		// Otherwise, build a query from individual parameters
		query = buildElasticsearchQuery(c)
	}

	// Execute search
	events, total, err := h.ESService.SearchSecurityEvents(query, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search events: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": events,
		"pagination": gin.H{
			"page":     page,
			"pageSize": pageSize,
			"total":    total,
			"pages":    (total + pageSize - 1) / pageSize,
		},
	})
}

// Helper function to build an Elasticsearch query from HTTP request params
func buildElasticsearchQuery(c *gin.Context) map[string]interface{} {
	// Start with a match_all query
	query := map[string]interface{}{
		"match_all": map[string]interface{}{},
	}
	
	// Add bool query if filters are provided
	var filters []map[string]interface{}
	
	// Add filters for common fields
	if severity := c.Query("severity"); severity != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"severity": severity,
			},
		})
	}
	
	if category := c.Query("category"); category != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"category": category,
			},
		})
	}
	
	if sourceIP := c.Query("source_ip"); sourceIP != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"source_ip": sourceIP,
			},
		})
	}
	
	if destIP := c.Query("destination_ip"); destIP != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"destination_ip": destIP,
			},
		})
	}
	
	// Add time range filter
	if from := c.Query("from"); from != "" {
		if to := c.Query("to"); to != "" {
			filters = append(filters, map[string]interface{}{
				"range": map[string]interface{}{
					"timestamp": map[string]interface{}{
						"gte": from,
						"lte": to,
					},
				},
			})
		} else {
			filters = append(filters, map[string]interface{}{
				"range": map[string]interface{}{
					"timestamp": map[string]interface{}{
						"gte": from,
					},
				},
			})
		}
	} else if to := c.Query("to"); to != "" {
		filters = append(filters, map[string]interface{}{
			"range": map[string]interface{}{
				"timestamp": map[string]interface{}{
					"lte": to,
				},
			},
		})
	}
	
	// Add text search if provided
	if searchText := c.Query("search"); searchText != "" {
		query = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": map[string]interface{}{
					"multi_match": map[string]interface{}{
						"query":  searchText,
						"fields": []string{"message", "source_ip", "destination_ip", "device_id"},
					},
				},
			},
		}
	}
	
	// If we have filters, add them to the query
	if len(filters) > 0 {
		if boolQuery, ok := query["bool"].(map[string]interface{}); ok {
			boolQuery["filter"] = filters
		} else {
			query = map[string]interface{}{
				"bool": map[string]interface{}{
					"must":   query,
					"filter": filters,
				},
			}
		}
	}
	
	return query
}