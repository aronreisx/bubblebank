package api

import (
	"context"
	"log"
	"net/http"
	"time"

	db "github.com/aronreisx/bubblebank/db/sqlc"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Server serves HTTP requests for banking service
type Server struct {
	store            db.Store
	router           *gin.Engine
	isReady          bool
	httpServer       *http.Server
	telemetryManager interface{}
	tracer           trace.Tracer
	meter            metric.Meter
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(store db.Store, telemetryManager interface{}) *Server {
	server := &Server{
		store:            store,
		telemetryManager: telemetryManager,
		tracer:           otel.Tracer("bubblebank-api"),
		meter:            otel.Meter("bubblebank-api"),
	}

	router := gin.Default()

	// Set trusted proxies to nil to not trust any proxy
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Printf("Warning: Failed to set trusted proxies: %v", err)
	}

	// Add OpenTelemetry middleware
	router.Use(otelgin.Middleware("bubblebank"))

	// Add custom observability middleware
	router.Use(server.observabilityMiddleware())

	// Add health and readiness endpoints
	router.GET("/health", server.healthCheck)
	router.GET("/ready", server.readinessCheck)

	// Add API endpoints
	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccounts)

	server.router = router
	return server
}

// observabilityMiddleware adds custom telemetry to requests
func (server *Server) observabilityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Start custom span for business logic
		ctx, span := server.tracer.Start(c.Request.Context(), "http_request")
		defer span.End()

		// Add business context to span
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.user_agent", c.Request.UserAgent()),
		)

		// Update context
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Record metrics and span attributes
		duration := time.Since(start)
		status := c.Writer.Status()

		span.SetAttributes(
			attribute.Int("http.status_code", status),
			attribute.Float64("http.duration_ms", float64(duration.Nanoseconds())/1e6),
		)

		// Record error if status >= 400
		if status >= 400 {
			span.SetAttributes(attribute.String("error", "true"))
		}
	}
}

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	server.httpServer = &http.Server{
		Addr:    address,
		Handler: server.router,
	}
	return server.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (server *Server) Shutdown(ctx context.Context) error {
	if server.httpServer != nil {
		return server.httpServer.Shutdown(ctx)
	}
	return nil
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

// SetReady marks the server as ready, typically called after migrations complete
func (server *Server) SetReady() {
	server.isReady = true
}

// healthCheck handles the GET /health endpoint
func (server *Server) healthCheck(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"status":    "up",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// readinessCheck handles the GET /ready endpoint
// This is used by Kubernetes/Docker to determine if the application is ready to receive traffic
func (server *Server) readinessCheck(ctx *gin.Context) {
	if server.isReady {
		ctx.JSON(200, gin.H{
			"status":     "ready",
			"migrations": "complete",
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	ctx.JSON(503, gin.H{
		"status":     "not ready",
		"migrations": "pending",
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}
