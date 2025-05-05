package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)


const (
	//RequestIDHeader is the header key for the request ID
	RequestIDHeader = "X-Request-ID"
	//RequestIDContextKey is the context key for the request ID
	RequestIDContextKey = "request_id"
)

// correlationID is a middleware that adds a correlation ID to the request
// if the request has already got a correlation ID, it is passed through
func CorrelationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// check if we already have a correlation ID, it is passed through
		requestID := c.GetHeader(RequestIDHeader)

		// if not, generate a new one
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// add the correlation ID to the context
		c.Set(RequestIDContextKey, requestID)

		// set response header
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}

// GetRequestID retrieves the request ID from the Gin context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDContextKey); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

