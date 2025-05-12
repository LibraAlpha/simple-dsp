import { defineStore } from 'pinia'
import axios from 'axios'

export const useConfigStore = defineStore('config', {
  state: () => ({
    configs: [],
    loading: false,
    error: null
  }),

  actions: {
    // 获取配置列表
    async getConfigList() {
      this.loading = true
      this.error = null
      try {
        const response = await axios.get('/api/v1/configs')
        this.configs = response.data
        return response.data
      } catch (error) {
        this.error = error.message
        throw error
      } finally {
        this.loading = false
      }
    },

    // 获取单个配置
    async getConfig(key) {
      try {
        const response = await axios.get(`/api/v1/configs/${key}`)
        return response.data
      } catch (error) {
        throw error
      }
    },

    // 获取配置历史
    async getConfigHistory(key) {
      try {
        const response = await axios.get(`/api/v1/configs/${key}/history`)
        return response.data
      } catch (error) {
        throw error
      }
    },

    // 添加配置
    async addConfig(key, value) {
      try {
        await axios.post(`/api/v1/configs/${key}`, value)
      } catch (error) {
        throw error
      }
    },

    // 更新配置
    async updateConfig(key, value) {
      try {
        await axios.put(`/api/v1/configs/${key}`, value)
      } catch (error) {
        throw error
      }
    },

    // 删除配置
    async deleteConfig(key) {
      try {
        await axios.delete(`/api/v1/configs/${key}`)
      } catch (error) {
        throw error
      }
    }
  }
}) 