package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"taskqueue/internal/database"
	"taskqueue/internal/models"
	"taskqueue/internal/queue"
	"taskqueue/internal/websocket"
	"taskqueue/pkg/logger"
)

// TaskHandler provides HTTP handlers for task operations.
type TaskHandler struct {
	DB  *pgxpool.Pool
	Q   *queue.Client
	Hub *websocket.Hub
}

// Create handles POST /api/tasks to create and enqueue a task.
func (h *TaskHandler) Create(c *gin.Context) {
	var req struct {
		Name     string          `json:"name" form:"name" binding:"required"`
		Type     string          `json:"type" form:"type" binding:"required"`
		Priority string          `json:"priority" form:"priority" binding:"required,oneof=low medium high"`
		Payload  json.RawMessage `json:"payload" form:"payload"`
	}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(int64)

	task := &models.Task{
		UserID:   userID,
		Name:     req.Name,
		Type:     req.Type,
		Priority: req.Priority,
		Status:   "pending",
		Payload:  req.Payload,
	}
	
	if err := database.CreateTask(c.Request.Context(), h.DB, task); err != nil {
		logger.Error("create task:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create task"})
		return
	}

	// Enqueue task to SQS with priority
	queueMessage := map[string]interface{}{
		"task_id": task.ID,
		"type":    task.Type,
		"payload": req.Payload,
	}
	messageBody, _ := json.Marshal(queueMessage)
	
	msgID, err := h.Q.EnqueueWithPriority(c.Request.Context(), string(messageBody), req.Priority)
	if err != nil {
		logger.Error("queue error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "queue error"})
		return
	}
	
	// Update task with message ID
	if err := database.UpdateTaskStatus(c.Request.Context(), h.DB, task.ID, "queued", msgID); err != nil {
		logger.Error("update task status:", err)
	}
	task.MessageID = msgID
	task.Status = "queued"

	// Broadcast task creation via WebSocket
	if h.Hub != nil {
		h.Hub.BroadcastToUser(userID, "task_created", task)
	}

	if c.GetHeader("HX-Request") != "" {
		c.HTML(http.StatusAccepted, "partials/row.html", task)
		return
	}
	c.JSON(http.StatusAccepted, task)
}

// List handles GET /api/tasks to list tasks for the user.
func (h *TaskHandler) List(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(int64)

	// Parse query parameters
	filter := &database.TaskFilter{
		Status:   c.Query("status"),
		Type:     c.Query("type"),
		Priority: c.Query("priority"),
		Limit:    50,
		Offset:   0,
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	tasks, err := database.ListTasks(c.Request.Context(), h.DB, userID, filter)
	if err != nil {
		logger.Error("list tasks:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tasks"})
		return
	}

	if c.GetHeader("HX-Request") != "" {
		c.HTML(http.StatusOK, "partials/rows.html", tasks)
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// Get handles GET /api/tasks/:id to get a single task.
func (h *TaskHandler) Get(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(int64)

	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	task, err := database.GetTask(c.Request.Context(), h.DB, taskID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// Cancel handles DELETE /api/tasks/:id to cancel a task.
func (h *TaskHandler) Cancel(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(int64)

	taskIDStr := c.Param("id")
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	if err := database.CancelTask(c.Request.Context(), h.DB, taskID, userID); err != nil {
		logger.Error("cancel task:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Broadcast task cancellation via WebSocket
	if h.Hub != nil {
		h.Hub.BroadcastToUser(userID, "task_cancelled", gin.H{"task_id": taskID})
	}

	c.JSON(http.StatusOK, gin.H{"message": "task cancelled"})
}

// Stats handles GET /api/tasks/stats to get task statistics.
func (h *TaskHandler) Stats(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(int64)

	stats, err := database.GetTaskStats(c.Request.Context(), h.DB, userID)
	if err != nil {
		logger.Error("get task stats:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
