package monitor

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"oracle-monitor/internal/config"
	"oracle-monitor/internal/storage"

	_ "github.com/godror/godror"
)

type Service struct {
    cfg         *config.Config
    store       *storage.Storage
    workerPool  chan struct{}
    wg          sync.WaitGroup
}


func NewService(cfg *config.Config) *Service {
    return &Service{
        cfg:        cfg,
        store:      storage.NewStorage(),
        workerPool: make(chan struct{}, cfg.Monitor.MaxConcurrent),
    }
}

func (s *Service) Start(ctx context.Context) error {
    ticker := time.NewTicker(time.Duration(s.cfg.Monitor.Interval) * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return nil
        case <-ticker.C:
            s.runMonitorCycle(ctx)
        }
    }
}

func (s *Service) runMonitorCycle(ctx context.Context) {
    enabledDBs := s.getEnabledDatabases()
    resultChan := make(chan storage.MonitorResult, len(enabledDBs))

    // 启动监控任务
    for _, db := range enabledDBs {
        s.wg.Add(1)
        go func(db config.DatabaseConfig) {
            defer s.wg.Done()
            s.workerPool <- struct{}{} // 获取工作槽
            defer func() { <-s.workerPool }() // 释放工作槽

            result := s.checkDatabase(ctx, db)
            resultChan <- result
        }(db)
    }

    // 等待所有任务完成
    go func() {
        s.wg.Wait()
        close(resultChan)
    }()

    // 处理结果
    for result := range resultChan {
        s.handleResult(result)
    }
}

func (s *Service) checkDatabase(ctx context.Context, db config.DatabaseConfig) storage.MonitorResult {
    start := time.Now()
    result := storage.MonitorResult{
        DBName:    db.Name,
        CheckTime: start,
    }

    connStr := fmt.Sprintf(
        "user=%s password=%s connectString=%s:%d/%s",
        db.Username,
        db.Password,
        db.Host,
        db.Port,
        db.ServiceName,
    )

    // 设置连接超时
    timeoutCtx, cancel := context.WithTimeout(ctx, 
        time.Duration(s.cfg.Monitor.Timeout)*time.Second)
    defer cancel()

    conn, err := sql.Open("godror", connStr)
    if err != nil {
        result.Error = err.Error()
        return result
    }
    defer conn.Close()

    // 执行测试查询
    err = conn.PingContext(timeoutCtx)
    result.ResponseTime = time.Since(start).Seconds()
    
    if err != nil {
        result.Error = err.Error()
        return result
    }

    result.Status = true
    return result
}

func (s *Service) handleResult(result storage.MonitorResult) {
    // 保存结果
    if err := s.store.SaveResult(result); err != nil {
        log.Printf("Failed to save result for %s: %v", result.DBName, err)
    }

    // 处理告警
    if !result.Status {
        log.Printf("Database %s is down: %s", result.DBName, result.Error)
        // 这里可以添加告警逻辑
    }
}

func (s *Service) getEnabledDatabases() []config.DatabaseConfig {
    var enabled []config.DatabaseConfig
    for _, db := range s.cfg.Databases {
        if db.Enable {
            enabled = append(enabled, db)
        }
    }
    return enabled
}

func (s *Service) Shutdown() {
    s.wg.Wait() // 等待所有任务完成
}
