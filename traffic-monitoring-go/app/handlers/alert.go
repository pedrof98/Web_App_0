
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"traffic-monitoring-go/app/models"
	"traffic-monitoring-go/app/siem/elasticsearch"
	"traffic-monitoring-go/app/siem/notifications"
)

// Alert handler handles alert-related endpoints
type AlertHandler struct {
	DB 					*gorm.DB
	NotificationManager	*notifications.NotificationManager
	ESService			*elasticsearch.Service
}


// NewAlertHandler creates a new AlertHandler
func NewAlertHandler(db *gorm.DB, esService *elasticsearch.Service) *AlertHandler {
	// create a notification manager
	manager := notifications.NewNotificationManager(db)

	// register default notification channels
	// for demonstration, using placeholder config values
	emailChannel := notifications.NewEmailChannel(notifications.EmailConfig{
		BaseNotificationConfig: notifications.BaseNotificationConfig{
			Enabled: false, // disabled by default since it needs a real SMTP config
			Name:	"default-email",
		},
		SMTPServer:		"smtp.example.com",
		SMTPPort:		587,
		Username:		"username",
		Password:		"password",
		FromAddress:	"siem@example.com",
		ToAddresses:	[]string{"alerts@example.com"},
	})

	webhookChannel := notifications.NewWebhookChannel(notifications.WebhookConfig{
		BaseNotificationConfig:	notifications.BaseNotificationConfig{
			Enabled: false,
			Name:	 "default-webhook",
		},
		URL:	"https://example.com/webhook",
		Method:	"POST",
	})

	manager.RegisterChannel(emailChannel)
	manager.RegisterChannel(webhookChannel)


	return &AlertHandler{
		DB:		 				db,
		NotificationManager:	manager,
		ESService: 				esService,
	}
}


//GetAlerts handles GET /alerts
func (h *AlertHandler) GetAlerts(c *gin.Context) {
	var alerts []models.Alert

	// Basic pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pagesize", "50"))
	offset := (page - 1) * pageSize

	// Basic filtering by severity and status
	severity := c.Query("severity")
	status := c.Query("status")

	// Create a query builder
	query := h.DB.Model(&models.Alert{}).Preload("Rule")

	if severity != "" {
		query = query.Where("severity = ?", severity)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// order by timestamp descending (most recent first)
	query = query.Order("timestamp DESC")

	// Count total for pagination info
	var total int64
	query.Count(&total)

	//Execute the query with pagination
	if err:= query.Offset(offset).Limit(pageSize).Find(&alerts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": alerts,
		"pagination": gin.H{
			"page": page,
			"pageSize": pageSize,
			"total": total,
			"pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}


// GetAlert handles GET /alerts/:id
func (h *AlertHandler) GetAlert(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	var alert models.Alert
	if err := h.DB.Preload("Rule").Preload("SecurityEvent").First(&alert, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
		return
	}

	c.JSON(http.StatusOK, alert)
}



// UpdateAlert handles PUT /alerts/:id
func (h *AlertHandler) UpdateAlert(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Alert ID"})
		return
	}

	var alert models.Alert
	if err := h.DB.First(&alert, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
		return
	}

	// Only update specific fields, not the entire alert
	var updateData struct {
		Status		*models.AlertStatus	`json:"status,omitempty"`
		AssignedTo	*uint			`json:"assigned_to,omitempty"`
		Resolution	*string			`json:"resolution,omitempty"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Apply updates that were provided
	if updateData.Status != nil {
		alert.Status = *updateData.Status
	}
	if updateData.AssignedTo != nil {
		alert.AssignedTo = updateData.AssignedTo
	}
	if updateData.Resolution != nil {
		alert.Resolution = *updateData.Resolution
	}

	if err := h.DB.Save(&alert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//Update in elastisearch if available
	if h.ESService != nil {
		if err := h.ESService.IndexAlert(&alert); err != nil {
			// log error but dont fail the request
			c.JSON(http.StatusOK, gin.H{
				"alert": alert,
				"warning": "Alert updated in database but could not be indexed in Elasticsearch: " + err.Error(),
			})
			return
	}
}


	c.JSON(http.StatusOK, alert)
}



// SendNotitification handles POST /alerts/:id/notify
func (h *AlertHandler) SendNotification(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}


	// check if the alert exists
	var alert models.Alert
	if err := h.DB.First(&alert, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
		return
	}


	// send notifications
	err = h.NotificationManager.SendNotification(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification sent successfully"})
}

// GetNotificationChannels handles GET /notifications/channels
func (h *AlertHandler) GetNotificationChannels(c *gin.Context) {
	channelNames := h.NotificationManager.GetChannelNames()
	channels := make([]string, len(channelNames))
	copy(channels, channelNames)

	c.JSON(http.StatusOK, gin.H{
		"channels": channels,
	})
}






