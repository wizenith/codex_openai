package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"taskqueue/internal/config"
	"taskqueue/internal/database"
	"taskqueue/internal/queue"
	"taskqueue/pkg/logger"
)

func main() {
	cfg := config.Load()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

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

	r.GET("/healthz", func(c *gin.Context) {
		if err := db.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db unreachable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/api/tasks", func(c *gin.Context) {
		// simplified task enqueue example
		var req struct {
			Payload string `json:"payload"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		id, err := q.Enqueue(ctx, req.Payload)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue"})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"message_id": id})
	})

	logger.Info("starting server on", cfg.Port)
	r.Run(":" + cfg.Port)
}
