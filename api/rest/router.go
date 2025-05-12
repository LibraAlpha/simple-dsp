package rest

import "github.com/gin-gonic/gin"

func NewRouter() *gin.Engine {
	r := gin.Default()

	// 竞价接口
	r.POST("/api/v1/bid", func(c *gin.Context) {
		var req BidRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		resp, err := biddingEngine.ProcessBid(req)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, resp)
	})

	// 数据报表接口
	r.GET("/api/v1/reports", reportHandler.GetDailyReport)

	return r
}
