package api

import (
	"traffic-monitoring-go/internal/api/handlers"
	"traffic-monitoring-go/internal/api/middleware"
	"traffic-monitoring-go/internal/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Router sets up all API routes
type Router struct {
	engine     *gin.Engine
	log        *logrus.Logger
	jwtManager *auth.JWTManager

	// handlers
	ruleHandler          *handlers.RuleHandler
	alertHandler         *handlers.AlertHandler
	securityEventHandler *handlers.SecurityEventHandler
	authHandler          *handlers.AuthHandler
	userHandler          *handlers.UserHandler
	v2xHandler           *handlers.V2XHandler
	// TODO: add more handlers here
}

// newRouter creates a new router
func NewRouter(log *logrus.Logger,
	jwtManager *auth.JWTManager,
	ruleHandler *handlers.RuleHandler,
	alertHandler *handlers.AlertHandler,
	securityEventHandler *handlers.SecurityEventHandler,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	v2xHandler *handlers.V2XHandler,
) *Router {
	return &Router{
		engine:               gin.New(),
		log:                  log,
		jwtManager:           jwtManager,
		ruleHandler:          ruleHandler,
		alertHandler:         alertHandler,
		securityEventHandler: securityEventHandler,
		authHandler:          authHandler,
		userHandler:          userHandler,
		v2xHandler:           v2xHandler,
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

		// alerts endpoints will go here
		alerts := v1.Group("/alerts")
		{
			alerts.GET("", r.alertHandler.List)
			alerts.GET("/:id", r.alertHandler.Get)
			alerts.POST("", r.alertHandler.Create)
			alerts.PUT("/:id", r.alertHandler.Update)
			alerts.DELETE("/:id", r.alertHandler.Delete)
			alerts.POST("/:id/assign", r.alertHandler.Assign)
		}
		//security events endpoints here
		securityEvents := v1.Group("/security-events")
		{
			securityEvents.GET("", r.securityEventHandler.List)
			securityEvents.GET("/:id", r.securityEventHandler.Get)
			securityEvents.POST("", r.securityEventHandler.Create)
			securityEvents.POST("/batch", r.securityEventHandler.BatchCreate)
			securityEvents.DELETE("/:id", r.securityEventHandler.Delete)

			// additional endpoints here
		}

		// V2X endpoints
		v2x := v1.Group("/v2x")
		{
			v2x.POST("/messages", r.v2xHandler.ProcessMessage)
			v2x.POST("/messages/batch", r.v2xHandler.ProcessBatch)
			v2x.GET("/metrics", r.v2xHandler.GetMetrics)
			v2x.GET("/config", r.v2xHandler.GetConfiguration)
			v2x.PUT("/config", middleware.AuthMiddleware(r.jwtManager), middleware.RequireAdminOrAnalyst(), r.v2xHandler.UpdateConfiguration)
			v2x.GET("/vehicles", r.v2xHandler.GetVehicleStates)
		}

	}
}

// engine returns the configured gin engine
func (r *Router) Engine() *gin.Engine {
	return r.engine
}
