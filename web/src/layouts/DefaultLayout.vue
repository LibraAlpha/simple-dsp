<template>
  <el-container class="layout-container">
    <el-aside width="200px" class="aside">
      <div class="logo">
        <img src="@/assets/logo.png" alt="DSP Admin">
        <span>DSP 管理系统</span>
      </div>
      <el-menu
        :router="true"
        :default-active="route.path"
        class="menu"
      >
        <el-menu-item index="/">
          <el-icon><Monitor /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/configs">
          <el-icon><Setting /></el-icon>
          <span>配置管理</span>
        </el-menu-item>
        <el-menu-item index="/stats">
          <el-icon><DataLine /></el-icon>
          <span>数据统计</span>
        </el-menu-item>
        <el-menu-item index="/frequency">
          <el-icon><Timer /></el-icon>
          <span>频次控制</span>
        </el-menu-item>
        <el-menu-item index="/budget">
          <el-icon><Money /></el-icon>
          <span>预算管理</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    
    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-button
            type="text"
            @click="toggleCollapse"
          >
            <el-icon>
              <Fold v-if="!isCollapse" />
              <Expand v-else />
            </el-icon>
          </el-button>
          <el-breadcrumb>
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item>{{ route.name }}</el-breadcrumb-item>
          </el-breadcrumb>
        </div>
        <div class="header-right">
          <el-dropdown>
            <span class="user-info">
              <el-avatar :size="32" src="https://avatars.githubusercontent.com/u/1?v=4" />
              <span>管理员</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item>个人设置</el-dropdown-item>
                <el-dropdown-item divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      
      <el-main class="main">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import {
  Monitor,
  Setting,
  DataLine,
  Timer,
  Money,
  Fold,
  Expand
} from '@element-plus/icons-vue'

const route = useRoute()
const isCollapse = ref(false)

const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}
</script>

<style lang="scss" scoped>
.layout-container {
  height: 100vh;
  
  .aside {
    background-color: #304156;
    
    .logo {
      height: 60px;
      display: flex;
      align-items: center;
      padding: 0 20px;
      color: #fff;
      
      img {
        width: 32px;
        height: 32px;
        margin-right: 12px;
      }
      
      span {
        font-size: 16px;
        font-weight: 600;
      }
    }
    
    .menu {
      border-right: none;
      background-color: transparent;
    }
  }
  
  .header {
    background-color: #fff;
    border-bottom: 1px solid #dcdfe6;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 20px;
    
    .header-left {
      display: flex;
      align-items: center;
      
      .el-button {
        margin-right: 20px;
      }
    }
    
    .header-right {
      .user-info {
        display: flex;
        align-items: center;
        cursor: pointer;
        
        span {
          margin-left: 8px;
        }
      }
    }
  }
  
  .main {
    background-color: #f0f2f5;
    padding: 20px;
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style> 