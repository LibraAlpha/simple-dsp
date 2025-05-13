package handlers

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "simple-dsp/internal/campaign"
    "simple-dsp/internal/models"
    "simple-dsp/pkg/logger"
    "gorm.io/gorm"
)

// CampaignHandler 广告计划处理器
type CampaignHandler struct {
    db        *gorm.DB
    logger    *logger.Logger
    configMgr *campaign.ConfigManager
}

// NewCampaignHandler 创建新的广告计划处理器
func NewCampaignHandler(db *gorm.DB, logger *logger.Logger, configMgr *campaign.ConfigManager) *CampaignHandler {
    return &CampaignHandler{
        db:        db,
        logger:    logger,
        configMgr: configMgr,
    }
}

// RegisterRoutes 注册路由
func (h *CampaignHandler) RegisterRoutes(r *gin.Engine) {
    g := r.Group("/api/v1/campaigns")
    {
        g.POST("", h.CreateCampaign)
        g.GET("", h.ListCampaigns)
        g.GET("/:id", h.GetCampaign)
        g.PUT("/:id", h.UpdateCampaign)
        g.DELETE("/:id", h.DeleteCampaign)
        g.PUT("/:id/tracking", h.UpdateTrackingConfig)
    }
}

// CreateCampaign 创建广告计划
func (h *CampaignHandler) CreateCampaign(c *gin.Context) {
    var config campaign.CampaignConfig
    if err := c.ShouldBindJSON(&config); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 验证配置
    if err := campaign.ValidateConfig(&config); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 创建数据库记录
    var model models.Campaign
    if err := model.FromCampaignConfig(&config); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    if err := h.db.Create(&model).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // 更新配置管理器
    h.configMgr.SetConfig(&config)

    c.JSON(http.StatusCreated, config)
}

// ListCampaigns 列出广告计划
func (h *CampaignHandler) ListCampaigns(c *gin.Context) {
    var campaigns []models.Campaign
    if err := h.db.Find(&campaigns).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    configs := make([]*campaign.CampaignConfig, 0, len(campaigns))
    for _, model := range campaigns {
        config, err := model.ToCampaignConfig()
        if err != nil {
            h.logger.Error("转换广告计划配置失败", "error", err)
            continue
        }
        configs = append(configs, config)
    }

    c.JSON(http.StatusOK, configs)
}

// GetCampaign 获取广告计划
func (h *CampaignHandler) GetCampaign(c *gin.Context) {
    id := c.Param("id")
    var model models.Campaign
    if err := h.db.First(&model, "id = ?", id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
        return
    }

    config, err := model.ToCampaignConfig()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, config)
}

// UpdateCampaign 更新广告计划
func (h *CampaignHandler) UpdateCampaign(c *gin.Context) {
    id := c.Param("id")
    var config campaign.CampaignConfig
    if err := c.ShouldBindJSON(&config); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 验证配置
    if err := campaign.ValidateConfig(&config); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 更新数据库记录
    var model models.Campaign
    if err := model.FromCampaignConfig(&config); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    if err := h.db.Where("id = ?", id).Updates(&model).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // 更新配置管理器
    h.configMgr.SetConfig(&config)

    c.JSON(http.StatusOK, config)
}

// DeleteCampaign 删除广告计划
func (h *CampaignHandler) DeleteCampaign(c *gin.Context) {
    id := c.Param("id")
    if err := h.db.Delete(&models.Campaign{}, "id = ?", id).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // 从配置管理器中移除
    h.configMgr.RemoveConfig(id)

    c.Status(http.StatusNoContent)
}

// UpdateTrackingConfig 更新跟踪配置
func (h *CampaignHandler) UpdateTrackingConfig(c *gin.Context) {
    id := c.Param("id")
    var trackingConfigs map[campaign.TrackingType]*campaign.TrackingConfig
    if err := c.ShouldBindJSON(&trackingConfigs); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 获取现有配置
    var model models.Campaign
    if err := h.db.First(&model, "id = ?", id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "campaign not found"})
        return
    }

    // 更新跟踪配置
    trackingConfigsJSON, err := json.Marshal(trackingConfigs)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    model.TrackingConfigs = trackingConfigsJSON
    model.UpdateTime = time.Now()

    if err := h.db.Save(&model).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // 更新配置管理器
    config, err := model.ToCampaignConfig()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    h.configMgr.SetConfig(config)

    c.JSON(http.StatusOK, trackingConfigs)
} 