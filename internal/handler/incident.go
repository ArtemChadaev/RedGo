package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ArtemChadaev/RedGo/internal/domain"
	"github.com/gin-gonic/gin"
)

// POST /api/v1/incidents/
func (h *Handler) createIncident(c *gin.Context) {
	var input domain.Incident

	// 1. Валидация JSON (x, y и корректность статуса, если он передан)
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid input: x, y are required; status must be 'active' or 'inactive'",
		})
		return
	}

	// 2. Гарантируем наличие статуса: если пусто — ставим active
	if input.Status == "" {
		input.Status = domain.StatusActive
	}

	// 3. Сохранение в базу через сервис
	if err := h.services.IncidentService.CreateIncident(c.Request.Context(), &input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, input)
}

// GET /api/v1/incidents/
func (h *Handler) getIncidents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "0"))

	incidents, err := h.services.IncidentService.GetIncidents(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, incidents)
}

// GET /api/v1/incidents/:id
func (h *Handler) getIncidentByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	inc, err := h.services.IncidentService.GetIncidentByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, inc)
}

// PUT /api/v1/incidents/:id
func (h *Handler) updateIncident(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var input domain.UpdateIncidentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input format or status"})
		return
	}

	// Проверка: прислано ли хотя бы одно поле
	if input.X == nil && input.Y == nil && input.Description == nil && input.Status == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one field (x, y, description, or status) must be provided"})
		return
	}

	if err := h.services.IncidentService.Update(c.Request.Context(), id, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// DELETE /api/v1/incidents/:id
func (h *Handler) deleteIncident(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.services.IncidentService.DeleteIncident(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GET /api/v1/incidents/stats
func (h *Handler) getStats(c *gin.Context) {
	count, err := h.services.IncidentService.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_count": count})
}

// POST /api/v1/location/check
func (h *Handler) checkLocation(c *gin.Context) {
	var input struct {
		UserID int     `json:"user_id" binding:"required"`
		X      float64 `json:"x" binding:"required"`
		Y      float64 `json:"y" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nearby, err := h.services.IncidentService.CheckLocation(c.Request.Context(), input.UserID, input.X, input.Y)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, nearby)
}

// GET /api/v1/system/health
func (h *Handler) healthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	var status = "ok"
	details := make(map[string]interface{})

	// 1. Проверка PostgreSQL
	// Предполагаем, что в репозитории есть метод Ping
	if err := h.services.HealthCheckDB(ctx); err != nil {
		status = "unhealthy"
		details["postgres"] = "down: " + err.Error()
	} else {
		details["postgres"] = "up"
	}

	// 2. Проверка Redis
	if err := h.services.HealthCheckRedis(ctx); err != nil {
		status = "unhealthy"
		details["redis"] = "down: " + err.Error()
	} else {
		details["redis"] = "up"
	}

	// 3. Сбор статистики воркеров
	workerStats, err := h.worker.GetStats(ctx)
	if err != nil {
		details["worker_stats"] = "error: " + err.Error()
	} else {
		details["worker_stats"] = workerStats
	}

	// Определяем HTTP статус
	httpStatus := http.StatusOK
	if status != "ok" {
		httpStatus = http.StatusServiceUnavailable // 503
	}

	c.JSON(httpStatus, gin.H{
		"status":    status,
		"details":   details,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
