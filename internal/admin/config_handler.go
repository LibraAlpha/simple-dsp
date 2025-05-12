package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"simple-dsp/internal/config"
)

// ConfigHandler 配置管理处理器
type ConfigHandler struct {
	configService *config.Service
}

// NewConfigHandler 创建配置管理处理器
func NewConfigHandler(configService *config.Service) *ConfigHandler {
	return &ConfigHandler{
		configService: configService,
	}
}

// RegisterRoutes 注册路由
func (h *ConfigHandler) RegisterRoutes(router *gin.Engine) {
	group := router.Group("/api/v1/configs")
	{
		group.GET("", h.ListConfigs)
		group.GET("/:key", h.GetConfig)
		group.POST("/:key", h.SetConfig)
		group.DELETE("/:key", h.DeleteConfig)
		group.GET("/:key/history/:version", h.GetConfigHistory)
	}
}

// ListConfigs 列出所有配置
func (h *ConfigHandler) ListConfigs(c *gin.Context) {
	configs, err := h.configService.ListConfigs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, configs)
}

// GetConfig 获取配置
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	key := c.Param("key")
	config, err := h.configService.GetConfig(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, config)
}

// SetConfig 设置配置
func (h *ConfigHandler) SetConfig(c *gin.Context) {
	key := c.Param("key")
	var value interface{}
	if err := c.ShouldBindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置值"})
		return
	}

	// 获取更新者信息
	updatedBy := c.GetString("user_id") // 假设已经通过中间件设置了用户信息
	if updatedBy == "" {
		updatedBy = "system"
	}

	if err := h.configService.SetConfig(c.Request.Context(), key, value, updatedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "配置已更新"})
}

// DeleteConfig 删除配置
func (h *ConfigHandler) DeleteConfig(c *gin.Context) {
	key := c.Param("key")
	if err := h.configService.DeleteConfig(c.Request.Context(), key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "配置已删除"})
}

// GetConfigHistory 获取配置历史版本
func (h *ConfigHandler) GetConfigHistory(c *gin.Context) {
	key := c.Param("key")
	versionStr := c.Param("version")
	version, err := strconv.ParseInt(versionStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的版本号"})
		return
	}

	config, err := h.configService.GetConfigHistory(c.Request.Context(), key, version)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, config)
} 