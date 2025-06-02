package api

import (
	"log"

	db "github.com/aronreisx/bubblebank/db/sqlc"
	"github.com/gin-gonic/gin"
)

// Server serves HTTP requests for banking service
type Server struct {
	store   db.Store
	router  *gin.Engine
	isReady bool
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// Set trusted proxies to nil to not trust any proxy
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Printf("Warning: Failed to set trusted proxies: %v", err)
	}

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

// Start runs the HTTP server on a specific address.
func (server *Server) Start(address string) error {
	return server.router.Run(address)
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
		"status": "up",
	})
}

// readinessCheck handles the GET /ready endpoint
// This is used by Kubernetes/Docker to determine if the application is ready to receive traffic
func (server *Server) readinessCheck(ctx *gin.Context) {
	if server.isReady {
		ctx.JSON(200, gin.H{
			"status":     "ready",
			"migrations": "complete",
		})
		return
	}

	ctx.JSON(503, gin.H{
		"status":     "not ready",
		"migrations": "pending",
	})
}
