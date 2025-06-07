package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"taskqueue/internal/config"
	"taskqueue/internal/database"
	"taskqueue/internal/handlers"
	"taskqueue/internal/queue"
	"taskqueue/pkg/logger"
)

func main() {
	cfg := config.Load()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.LoadHTMLGlob("web/templates/**/*.html")
	r.Static("/static", "web/static")

	ctx := context.Background()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("db connect:", err)
		return
	}
	defer db.Close()

	q, err := queue.New(ctx, cfg.AWSRegion, cfg.SQSQueueURL)
	if err != nil {
		logger.Error("sqs:", err)
		return
	}

	taskHandler := &handlers.TaskHandler{DB: db, Q: q}
	webHandler := &handlers.WebHandler{}

	r.GET("/healthz", func(c *gin.Context) {
		if err := db.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db unreachable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/api/tasks", taskHandler.Create)
	r.GET("/api/tasks", taskHandler.List)

	r.GET("/login", webHandler.Login)
	r.GET("/", webHandler.Dashboard)

	logger.Info("starting server on", cfg.Port)
	r.Run(":" + cfg.Port)
}
