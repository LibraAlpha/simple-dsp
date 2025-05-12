package admin

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册管理后台路由
func (s *Service) RegisterRoutes(r *gin.Engine) {
	// API 版本分组
	v1 := r.Group("/api/v1")
	{
		// 广告管理
		ads := v1.Group("/ads")
		{
			ads.POST("", s.CreateAd)           // 创建广告
			ads.PUT("/:id", s.UpdateAd)        // 更新广告
			ads.DELETE("/:id", s.DeleteAd)     // 删除广告
			ads.GET("/:id", s.GetAd)           // 获取广告信息
			ads.GET("", s.ListAds)             // 获取广告列表
			ads.GET("/:id/stats", s.GetAdStats) // 获取广告统计

			// 频次控制配置
			ads.PUT("/:id/frequency", s.UpdateFrequencyConfig)  // 更新频次控制配置
			ads.GET("/:id/frequency", s.GetFrequencyConfig)     // 获取频次控制配置
		}

		// 预算管理
		budgets := v1.Group("/budgets")
		{
			budgets.POST("", s.CreateBudget)           // 创建预算
			budgets.PUT("/:id", s.UpdateBudget)        // 更新预算
			budgets.GET("/:id", s.GetBudget)           // 获取预算信息
			budgets.GET("", s.ListBudgets)             // 获取预算列表
			budgets.POST("/:id/renew", s.RenewBudget)  // 续费预算
			budgets.GET("/:id/stats", s.GetBudgetStats) // 获取预算统计
		}

		// 数据统计
		stats := v1.Group("/stats")
		{
			stats.GET("/overview", s.GetStatsOverview) // 获取统计概览
			stats.GET("/daily", s.GetDailyStats)       // 获取每日统计
			stats.GET("/hourly", s.GetHourlyStats)     // 获取每小时统计
		}

		// 系统管理
		system := v1.Group("/system")
		{
			system.GET("/status", s.GetSystemStatus)   // 获取系统状态
			system.GET("/metrics", s.GetSystemMetrics) // 获取系统指标
		}
	}
} 