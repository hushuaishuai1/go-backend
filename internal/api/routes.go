package api

import (
	"net/http"
	"strconv"
	"github.com/hushuaishuai1/go-backend/internal/models"
	"github.com/hushuaishuai1/go-backend/internal/services"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// SetupRouter 现在返回一个 *echo.Echo 实例
func SetupRouter(am *services.AlertManager) *echo.Echo {
	// 创建一个新的Echo实例
	e := echo.New()

	// 配置CORS中间件 (这是Echo内置的、更简洁的方式)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		// 在生产环境中，为了安全，应将 "*" 替换为您的前端域名，例如 "https://jiagejiankong.store"
		AllowOrigins: []string{"*"}, 
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodOptions},
	}))

	// API路由分组
	api := e.Group("/api")

	// --- 路由定义 (语法已从Gin改为Echo) ---

	// POST /api/alerts: 创建新警报
	api.POST("/alerts", func(c echo.Context) error {
		var req models.CreateAlertRequest
		// Echo使用 c.Bind() 来绑定JSON
		if err := c.Bind(&req); err != nil {
			// Echo的处理器返回一个error
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		alert, err := am.CreateAlert(req)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create alert"})
		}
		return c.JSON(http.StatusCreated, alert)
	})

	// GET /api/alerts: 根据邮箱获取警报
	api.GET("/alerts", func(c echo.Context) error {
		// Echo使用 c.QueryParam() 来获取查询参数
		email := c.QueryParam("email")
		if email == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email query parameter is required"})
		}
		alerts, err := am.GetAlertsByEmail(email)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve alerts"})
		}
		return c.JSON(http.StatusOK, alerts)
	})

	// DELETE /api/alerts/:id: 删除警报
	api.DELETE("/alerts/:id", func(c echo.Context) error {
		// Echo使用 c.Param() 来获取路径参数
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid alert ID"})
		}
		if err := am.DeleteAlert(uint(id)); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete alert"})
		}
		return c.JSON(http.StatusOK, map[string]bool{"success": true})
	})

	return e
}