package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"taskqueue/internal/config"
	"taskqueue/internal/database"
	"taskqueue/internal/models"
	"taskqueue/internal/queue"
	"taskqueue/pkg/logger"
)

func main() {
	cfg := config.Load()

	logger.Init(cfg.LogLevel)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	ctx := context.Background()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("db connect:", err)
		return
	}
	defer db.Close()

	if err := database.Migrate(ctx, db); err != nil {
		logger.Error("migrate:", err)
		return
	}

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
		if err := q.HealthCheck(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "queue unreachable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")

	api.POST("/tasks", func(c *gin.Context) {
		var req struct {
			Name    string `json:"name"`
			Payload string `json:"payload"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		msgID, err := q.Enqueue(ctx, req.Payload)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue"})
			return
		}
		task := &models.Task{ // simplified model
			Name:      req.Name,
			Payload:   []byte(req.Payload),
			Status:    "queued",
			MessageID: msgID,
		}
		id, err := models.InsertTask(ctx, db, task)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db insert failed"})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"id": id, "message_id": msgID})
	})

	api.GET("/tasks", func(c *gin.Context) {
		tasks, err := models.ListTasks(ctx, db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		c.JSON(http.StatusOK, tasks)
	})

	api.GET("/tasks/:id", func(c *gin.Context) {
		idParam := c.Param("id")
		tid, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		task, err := models.GetTask(ctx, db, tid)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		c.JSON(http.StatusOK, task)
	})

	logger.Info("starting server on", cfg.Port)
	r.Run(":" + cfg.Port)
}
