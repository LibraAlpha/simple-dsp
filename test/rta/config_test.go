package rta_test

import (
    "testing"
    "time"
    
    "simple-dsp/internal/rta"
    "github.com/stretchr/testify/assert"
)

func TestConfigManager(t *testing.T) {
    mgr := rta.NewConfigManager()
    
    // 测试设置和获取配置
    t.Run("设置和获取配置", func(t *testing.T) {
        config := &rta.TaskConfig{
            TaskID:            "test_task_1",
            Channel:           "test_channel",
            AdvertisingSpaceID: "test_ad_space",
            Timeout:           time.Millisecond * 100,
            Enabled:           true,
            Priority:          1,
            RetryCount:        2,
            RetryInterval:     time.Millisecond * 50,
            CacheExpiration:   time.Minute * 5,
        }
        
        mgr.SetConfig(config)
        
        got, exists := mgr.GetConfig("test_task_1")
        assert.True(t, exists)
        assert.Equal(t, config, got)
    })
    
    // 测试移除配置
    t.Run("移除配置", func(t *testing.T) {
        config := &rta.TaskConfig{
            TaskID: "test_task_2",
        }
        
        mgr.SetConfig(config)
        mgr.RemoveConfig("test_task_2")
        
        _, exists := mgr.GetConfig("test_task_2")
        assert.False(t, exists)
    })
    
    // 测试列出所有配置
    t.Run("列出所有配置", func(t *testing.T) {
        mgr := rta.NewConfigManager()
        
        configs := []*rta.TaskConfig{
            {TaskID: "task1"},
            {TaskID: "task2"},
            {TaskID: "task3"},
        }
        
        for _, config := range configs {
            mgr.SetConfig(config)
        }
        
        list := mgr.ListConfigs()
        assert.Equal(t, len(configs), len(list))
        
        // 验证所有配置都在列表中
        found := make(map[string]bool)
        for _, config := range list {
            found[config.TaskID] = true
        }
        
        for _, config := range configs {
            assert.True(t, found[config.TaskID])
        }
    })
    
    // 测试并发安全性
    t.Run("并发安全性", func(t *testing.T) {
        mgr := rta.NewConfigManager()
        done := make(chan bool)
        
        // 并发写入
        go func() {
            for i := 0; i < 100; i++ {
                mgr.SetConfig(&rta.TaskConfig{
                    TaskID: fmt.Sprintf("task_%d", i),
                })
            }
            done <- true
        }()
        
        // 并发读取
        go func() {
            for i := 0; i < 100; i++ {
                mgr.GetConfig(fmt.Sprintf("task_%d", i))
            }
            done <- true
        }()
        
        // 等待完成
        <-done
        <-done
    })
} 