import { defineStore } from 'pinia'
import request from '@/utils/request'

export const useBidStore = defineStore('bid', {
  state: () => ({
    bids: [],
    loading: false,
    error: null
  }),

  actions: {
    // 获取出价列表
    async getBidList(params) {
      this.loading = true
      this.error = null
      try {
        const response = await request.get('/api/v1/bids', { params })
        this.bids = response.data.list
        return response.data
      } catch (error) {
        this.error = error.message
        throw error
      } finally {
        this.loading = false
      }
    },

    // 获取单个出价策略
    async getBid(id) {
      try {
        const response = await request.get(`/api/v1/bids/${id}`)
        return response.data
      } catch (error) {
        throw error
      }
    },

    // 添加出价策略
    async addBid(data) {
      try {
        const response = await request.post('/api/v1/bids', data)
        return response.data
      } catch (error) {
        throw error
      }
    },

    // 更新出价策略
    async updateBid(data) {
      try {
        const response = await request.put(`/api/v1/bids/${data.id}`, data)
        return response.data
      } catch (error) {
        throw error
      }
    },

    // 删除出价策略
    async deleteBid(id) {
      try {
        const response = await request.delete(`/api/v1/bids/${id}`)
        return response.data
      } catch (error) {
        throw error
      }
    },

    // 更新出价策略状态
    async updateBidStatus(id, status) {
      try {
        const response = await request.patch(`/api/v1/bids/${id}/status`, { status })
        return response.data
      } catch (error) {
        throw error
      }
    },

    // 添加素材
    async addCreative(strategyId, creativeId) {
      try {
        const response = await request.post(`/api/v1/bids/${strategyId}/creatives`, {
          creativeId
        })
        return response.data
      } catch (error) {
        throw error
      }
    },

    // 移除素材
    async removeCreative(strategyId, creativeId) {
      try {
        const response = await request.delete(
          `/api/v1/bids/${strategyId}/creatives/${creativeId}`
        )
        return response.data
      } catch (error) {
        throw error
      }
    },

    // 获取策略统计数据
    async getStrategyStats(strategyId, startDate, endDate) {
      try {
        const response = await request.get(`/api/v1/bids/${strategyId}/stats`, {
          params: { startDate, endDate }
        })
        return response.data
      } catch (error) {
        throw error
      }
    }
  }
}) 