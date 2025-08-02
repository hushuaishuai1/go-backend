package api

import (
	"net/http"
	"strconv"
	"github.com/hushuaishuai1/go-backend/internal/models"
	"github.com/hushuaishuai1/go-backend/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(am *services.AlertManager) *gin.Engine {
	r := gin.Default()

	// 配置CORS（跨域资源共享）
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // 生产环境中应替换为您的前端域名
	r.Use(cors.New(config))

	// API路由分组
	api := r.Group("/api")
	{
		api.POST("/alerts", func(c *gin.Context) {
			var req models.CreateAlertRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			alert, err := am.CreateAlert(req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create alert"})
				return
			}
			c.JSON(http.StatusCreated, alert)
		})

		api.GET("/alerts", func(c *gin.Context) {
			email := c.Query("email")
			if email == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Email query parameter is required"})
				return
			}
			alerts, err := am.GetAlertsByEmail(email)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve alerts"})
				return
			}
			c.JSON(http.StatusOK, alerts)
		})

		api.DELETE("/alerts/:id", func(c *gin.Context) {
			idStr := c.Param("id")
			id, err := strconv.ParseUint(idStr, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
				return
			}
			if err := am.DeleteAlert(uint(id)); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete alert"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})
	}
	return r
}