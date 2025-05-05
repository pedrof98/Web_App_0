
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"traffic-monitoring-go/app/models"
	)



// RuleHandler handles rule-related endpoints
type RuleHandler struct {
	DB *gorm.DB
}

// NewRuleHandler creates a new RuleHandler
func NewRuleHandler(db *gorm.DB) *RuleHandler {
	return &RuleHandler{DB: db}
}


// GetRules handles GET /rules
func (h *RuleHandler) GetRules(c *gin.Context) {
	var rules []models.Rule

	// basic filtering by status
	status := c.Query("status")
	category := c.Query("category")

	// Create a query builder
	query := h.DB.Model(&models.Rule{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if category != "" {
		query = query.Where("category = ?", category)
	}

	// Order by name ascending
	query = query.Order("name ASC")

	if err := query.Find(&rules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rules)
}


// GetRule handles GET /rules/:id
func (h *RuleHandler) GetRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID"})
		return
	}

	var rule models.Rule
	if err := h.DB.First(&rule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	c.JSON(http.StatusOK, rule)
}


// CreateRule handles POST /rules
func (h *RuleHandler) CreateRule(c *gin.Context) {
	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}


	// set default status if not provided
	if rule.Status == "" {
		rule.Status = models.RuleStatusDisabled
	}

	if err := h.DB.Create(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}


// UpdateRule handles PUT /rules/:id
func (h *RuleHandler) UpdateRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID"})
		return
	}

	var rule models.Rule
	if err := h.DB.First(&rule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.DB.Save(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}


// DeleteRule handles DELETE /rules/:id
func (h *RuleHandler) DeleteRule(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID"})
		return
	}


	// check if any alerts reference this rule before deletion
	var alertCount int64
	if err := h.DB.Model(&models.Alert{}).Where("rule_id = ?", id).Count(&alertCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if alertCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot delete rule with existing alerts",
			"alert_count": alertCount,
		})
		return
	}

	if err := h.DB.Delete(&models.Rule{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rule deleted successfully"})
}













