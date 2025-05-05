package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// logger is a middleware that logs request information
func Logger(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		requestID := GetRequestID(c)


		// create a buffer to read the request body
		var requestBodyBuffer bytes.Buffer
		if method == "POST" || method == "PUT" || method == "PATCH" {
			requestBody := io.TeeReader(c.Request.Body, &requestBodyBuffer)
			bodyBytes, _ := io.ReadAll(requestBody)

			// replace the request body
			c.Request.Body = io.NopCloser(&requestBodyBuffer)

			// log the request details
			log.WithFields(logrus.Fields{
				"request_id":	requestID,
				"method":	method,
				"path":		path,
				"body":		string(bodyBytes),
			}).Info("Request received")
		} else {
			// log the request details without the body
			log.WithFields(logrus.Fields{
				"request_id":	requestID,
				"method":	method,
				"path":		path,
			}).Info("Request received")
		}

		// process request
		c.Next()

		// log the response details

		duration := time.Since(start)
		status := c.Writer.Status()

		// determine log level based on status code
		var logEntry *logrus.Entry
		if status >= 500 {
			logEntry = log.WithFields(logrus.Fields{
				"request_id":	requestID,
				"method":	method,
				"path":		path,
				"status":	status,
				"duration":	duration.String(),
				"errors":	c.Errors.Errors(),
			})
			logEntry.Error("Request completed with server error")
		} else if stauts >= 400 {
			logEntry = log.WithFields(logrus.Fields{
				"request_id":	requestID,
				"method":	method,
				"path":		path,
				"status":	status,
				"duration":	duration.String(),
			})
			logEntry.Warn("Request completed with client error")
		} else {
			logEntry = log.WithFields(logrus.Fields{
				"request_id":	requestID,
				"method":	method,
				"path":		path,
				"status":	status,
				"duration":	duration.String(),
			})
			logEntry.Info("Request completed successfully")
		}
	}
}

