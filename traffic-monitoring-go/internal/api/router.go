package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"traffic-monitoring-go/internal/api/handlers"
	"traffic-monitoring-go/internal/api/middleware"
)

// Router sets up all API routes
type Router struct {
	engine	*gin.Engine
	log	*logrus.Logger

	// handlers
	ruleHandler *handlers.RuleHandler
	// TODO: add more handlers here
}

// newRouter creates a new router
func NewRouter(log *logrus.Logger, ruleHandler *handlers.RuleHandler) *Router {
	return &Router{
		engine:		gin.New(),
		log:		log,
		ruleHandler:	ruleHandler,
	}
}

// setup configures all routes and middleware
func (r *Router) Setup() {
	// set up middleware
	r.engine.Use(middleware.CorrelationID())
	r.engine.Use(middleware.Logger(r.log))
	r.engine.Use(middleware.Recovery(r.log))

	// health check endpoint
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := r.engine.Group("/api/v1")
	{
		// rules endpoints
		rules := v1.Group("/rules")
		{
			rules.GET("", r.ruleHandler.List)
			rules.GET("/:id", r.ruleHandler.Get)
			rules.POST("", r.ruleHandler.Create)
			rules.PUT("/:id", r.ruleHandler.Update)
			rules.DELETE("/:id", r.ruleHandler.Delete)
		}

		//TODO: alerts endpoints will go here

		//TODO:  security events endpoints here

		// additional endpoints here
	}
}


// engine returns the configured gin engine
func (r *Router) Engine() *gin.Engine {
	return r.engine
}

