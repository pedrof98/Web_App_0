package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"traffic-monitoring-go/internal/dto"
)

// recovery is a middleware that recovers from any panis and writes a 500 response
func Recovery(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// get stack trace
				stack := string(debug.Stack())
				requestID := GetRequestID(c)

				// log the error with a stack trace
				log.WithFields(logrus.Fields{
					"request_id":	requestID,
					"error":	err,
					"stack":	stack,
				}).Error("Panic recovered in API request")

				// respond with a 500 error
				errorResponse := dto.Error{}
				errorResponse.Error.Code = "INTERNAL_ERROR"
				errorResponse.Error.Message = "An internal server error occurred"

				c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse)
			}
		}()

		c.Next()
	}
}

