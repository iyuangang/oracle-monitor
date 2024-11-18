package config

import (
	"encoding/json"
	"os"
)

type Config struct {
    Databases []DatabaseConfig `json:"databases"`
    Monitor   MonitorConfig   `json:"monitor"`
    Alert     AlertConfig    `json:"alert"`
}

type DatabaseConfig struct {
    Name        string `json:"db_name"`
    Host        string `json:"host"`
    Port        int    `json:"port"`
    ServiceName string `json:"service_name"`
    Username    string `json:"username"`
    Password    string `json:"password"`
    Enable      bool   `json:"enable"`
}

type MonitorConfig struct {
    Interval       int `json:"interval"`        // 监控间隔(秒)
    Timeout        int `json:"timeout"`         // 连接超时(秒)
    MaxConcurrent  int `json:"max_concurrent"`  // 最大并发数
    BatchSize      int `json:"batch_size"`      // 批次大小
}

type AlertConfig struct {
    InitialInterval int `json:"initial_interval"` // 初始告警间隔(秒)
    MaxInterval     int `json:"max_interval"`     // 最大告警间隔(秒)
}

func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}
