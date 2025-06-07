package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"taskqueue/internal/auth"
	"taskqueue/internal/config"
	"taskqueue/internal/database"
	"taskqueue/internal/handlers"
	"taskqueue/internal/middleware"
	"taskqueue/internal/queue"
	ws "taskqueue/internal/websocket"
	"taskqueue/pkg/logger"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

func main() {
	cfg := config.Load()

	// Set Gin mode
	if cfg.Port == "8080" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router with custom middleware
	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// Load templates and static files
	r.LoadHTMLGlob("web/templates/**/*.html")
	r.Static("/static", "web/static")

	ctx := context.Background()

	// Initialize database
	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("db connect:", err)
		return
	}
	defer db.Close()

	// Initialize SQS
	q, err := queue.New(ctx, cfg.AWSRegion, cfg.SQSQueueURL)
	if err != nil {
		logger.Error("sqs:", err)
		return
	}

	// Initialize WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Initialize OAuth provider
	oauthProvider := auth.NewGoogleOAuth(cfg.GoogleClientID, cfg.GoogleSecret, cfg.GoogleRedirect)

	// Initialize handlers
	authHandler := &handlers.AuthHandler{
		DB:            db,
		OAuthProvider: oauthProvider,
		JWTSecret:     cfg.JWTSecret,
	}
	
	taskHandler := &handlers.TaskHandler{
		DB:  db,
		Q:   q,
		Hub: hub,
	}
	
	webHandler := &handlers.WebHandler{}

	// Public routes
	r.GET("/healthz", func(c *gin.Context) {
		if err := db.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db unreachable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Auth routes
	r.GET("/login", webHandler.Login)
	r.GET("/auth/google", authHandler.GoogleLogin)
	r.GET("/auth/google/callback", authHandler.GoogleCallback)
	r.GET("/logout", authHandler.Logout)

	// Protected web routes
	protected := r.Group("/")
	protected.Use(middleware.AuthRequired(cfg.JWTSecret))
	{
		protected.GET("/", webHandler.Dashboard)
		
		// WebSocket endpoint
		protected.GET("/ws", func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			
			conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
			if err != nil {
				logger.Error("websocket upgrade:", err)
				return
			}
			
			client := ws.NewClient(hub, conn, userID.(int64))
			hub.Register(client)
			
			go client.WritePump()
			go client.ReadPump()
		})
	}

	// API routes
	api := r.Group("/api")
	api.Use(middleware.APIAuthRequired(cfg.JWTSecret))
	{
		api.GET("/user", authHandler.GetCurrentUser)
		
		// Task endpoints
		api.POST("/tasks", taskHandler.Create)
		api.GET("/tasks", taskHandler.List)
		api.GET("/tasks/stats", taskHandler.Stats)
		api.GET("/tasks/:id", taskHandler.Get)
		api.DELETE("/tasks/:id", taskHandler.Cancel)
	}

	logger.Info("starting server on", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		logger.Error("server error:", err)
	}
}
