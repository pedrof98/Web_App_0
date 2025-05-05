package respond

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gin-gonic/gin"
	"traffic-monitoring-go/internal/dto"
	"traffic-monitoring-go/internal/service"
)

// OK responds with a 200 OK and wraps the data in a success envelope
func OK(c *gin.Context, data interface{}, meta *dto.MetaInfo) {
	c.JSON(http.StatusOK, dto.Success[interface{}]{
		Data: data,
		Meta: meta,
	})
}

// created responds with a 201 created status and wraps the data in a success envelope
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, dto.Success[interface{}]{
		Data: data,
	})
}

// NoContent responds with a 204 no content status
func NoContent(c *gin.Context, err error) {
	c.Status(http.StatusNoContent)
}

// BadRequest responds with a 400 Bad Request status and error details
func BadRequest(c *gin.Context, err error) {
	code := "BAD_REQUEST"
	msg := "Invalid request"
	fields := make(map[string]string)

	var valErrs validator.ValidationErrors
	if errors.As(err, &valErrs) {
		for _, valErr := range valErrs {
			field := strings.ToLower(valErr.Field())
			fields[field] = validationErrorMessage(valErr)
		}
		msg = "Validation failed"
		code = "VALIDATION_ERRPR"
	} else {
		msg = err.Error()
	}

	c.JSON(http.StatusBadRequest, newErrorResponse(code, msg, fields))
}

// NotFound responds with a 404 not found status and error details
func NotFound(c *gin.Context, err error) {
	msg := "Respurce not found"
	if err != nil {
		msg = err.Error()
	}
	c.JSON(http.StatusNotFound, newErrorResponse("NOT_FOUND", msg.nil))
}

// Forbidden responds with a 403 forbidden status and error details
func Forbidden(c *gin.Context, err error) {
	msg := "Access denied"
	if err != nil {
		msg = err.Error()
	}
	c.JSON(http.StatusForbidden, newErrorResponse("FORBIDDEN", msg.nil))
}


// Conflict responds with a 409 conflict status and error details
func Conflict(c *gin.Context, err error) {
	msg := "Resource already exists"
	if err != nil {
		msg = err.Error()
	}
	c.JSON(http.StatusConflict, newErrorResponse("CONFLICT", msg.nil))
}


// Internal responds with a 500 internal server error status and error details
func Internal(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, newErrorResponse(
		"INTERNAL_ERROR",
		"An internal server error occurred",
		nil,
	))
}

// Error maps service errors to appropriate HTTP responses
func Error(c *gin.Context, err error) {
	if err == nil {
		return
	}

	switch {
	case errors.Is(err, service.ErrNotFound):
		NotFound(c, err)
	case errors.Is(err, service.ErrBadRequest):
		BadRequest(c, err)
	case errors.Is(err, service.ErrConflict):
		Conflict(c, err)
	case errors.Is(err, service.ErrForbidden):
		Forbidden(c, err)
	default:
		Internal(c, err)
	}
}


// newErrorResponse creates a new error response structure
func newErrorResponse(code, message string, fields map[string]string) dto.Error {
	response := dto.Error{}
	response.Error.Code = code
	response.Error.Message = message
	response.Error.Fields = fields
	return response
}

// validationErrorMessage returns a user-friendly message for a validation error
func validationErrorMessage(err validator.FieldError) string {
		switch err.Tag() {
		case "required":
			return "This field is required"
		case "min":
			return "This field must be at least " + err.Param() + " characters long"
		case "max":
			return "This field must be at most " + err.Param() + " characters long"
		case "oneof":
			return "This field must be one of: " + err.Param()
		default:
			return "This field is invalid"
		}
	}

																							
