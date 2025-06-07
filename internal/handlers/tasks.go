package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"taskqueue/internal/database"
	"taskqueue/internal/models"
	"taskqueue/internal/queue"
	"taskqueue/pkg/logger"
)

// TaskHandler provides HTTP handlers for task operations.
type TaskHandler struct {
	DB *pgxpool.Pool
	Q  *queue.Client
}

// Create handles POST /api/tasks to create and enqueue a task.
func (h *TaskHandler) Create(c *gin.Context) {
	var req struct {
		Name     string          `json:"name" form:"name" binding:"required"`
		Type     string          `json:"type" form:"type" binding:"required"`
		Priority string          `json:"priority" form:"priority" binding:"required"`
		Payload  json.RawMessage `json:"payload" form:"payload"`
	}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// In a real app the user ID would come from auth context.
	userID := int64(1)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	msgID, err := h.Q.Enqueue(c.Request.Context(), string(req.Payload))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "queue error"})
		return
	}
	task.MessageID = msgID
	if c.GetHeader("HX-Request") != "" {
		c.HTML(http.StatusAccepted, "partials/row.html", task)
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"task_id": task.ID, "message_id": msgID})
}

// List handles GET /api/tasks to list tasks for the user.
func (h *TaskHandler) List(c *gin.Context) {
	userID := int64(1)
	tasks, err := database.ListTasks(c.Request.Context(), h.DB, userID, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	if c.GetHeader("HX-Request") != "" {
		c.HTML(http.StatusOK, "partials/rows.html", tasks)
		return
	}
	c.JSON(http.StatusOK, tasks)
}
