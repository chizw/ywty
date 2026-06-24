// Package queue 封装 Asynq 客户端 + Server
// 队列基于 Redis，支持：critical/default/low 三个队列优先级
package queue

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/ywty/server/internal/config"
	"github.com/ywty/server/internal/logger"
)

// 队列名（与 config.Queue.Queues 对应）
const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)

// 任务类型（全局唯一）
const (
	TypeThumbnail    = "photo:thumbnail"   // 缩略图生成
	TypeProcess      = "photo:process"     // 图片处理（resize/watermark）
	TypeAutoDelete   = "photo:auto_delete" // 过期图片自动删除
	TypeSendMail     = "notify:send_mail"  // 异步发邮件
	TypeSendSMS      = "notify:send_sms"   // 异步发短信
	TypeSendCode     = "notify:send_code"  // 异步发验证码
	TypeTicketNotify = "ticket:notify"     // 工单通知
	TypeOrderCancel  = "order:cancel"      // 订单超时取消
	TypeOrderPaid    = "order:paid"        // 订单完成处理
)

// Client 队列客户端
type Client struct {
	c *asynq.Client
}

// NewClient 构造客户端
func NewClient(cfg config.Redis) *Client {
	c := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return &Client{c: c}
}

// Enqueue 入队一个任务
// queue 留空则使用 default
func (cl *Client) Enqueue(task *asynq.Task, opts ...asynq.Option) error {
	if cl == nil || cl.c == nil {
		return fmt.Errorf("queue client not initialized")
	}
	defaults := []asynq.Option{
		asynq.Queue(QueueDefault),
		asynq.MaxRetry(3),
	}
	opts = append(defaults, opts...)
	_, err := cl.c.Enqueue(task, opts...)
	return err
}

// EnqueueCritical 入队关键任务
func (cl *Client) EnqueueCritical(task *asynq.Task, opts ...asynq.Option) error {
	opts = append(opts, asynq.Queue(QueueCritical), asynq.MaxRetry(5))
	return cl.Enqueue(task, opts...)
}

// EnqueueLow 入队低优任务
func (cl *Client) EnqueueLow(task *asynq.Task, opts ...asynq.Option) error {
	opts = append(opts, asynq.Queue(QueueLow), asynq.MaxRetry(1))
	return cl.Enqueue(task, opts...)
}

// Close 关闭客户端
func (cl *Client) Close() error {
	if cl == nil || cl.c == nil {
		return nil
	}
	return cl.c.Close()
}

// NewServer 构造 asynq.Server（worker 进程使用）
func NewServer(cfg config.Redis, concurrency int, queues []string) *asynq.Server {
	if concurrency <= 0 {
		concurrency = 10
	}
	if len(queues) == 0 {
		queues = []string{QueueCritical, QueueDefault, QueueLow}
	}
	qMap := make(map[string]int, len(queues))
	for i, q := range queues {
		// 第一个最高优先级
		prio := len(queues) - i
		if prio < 1 {
			prio = 1
		}
		qMap[q] = prio
	}
	return asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.Addr, Password: cfg.Password, DB: cfg.DB},
		asynq.Config{
			Concurrency:    concurrency,
			Queues:         qMap,
			StrictPriority: true,
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				if task != nil {
					logger.L.Warn("task failed",
						zap.String("type", task.Type()),
						zap.Error(err),
					)
				}
			}),
		},
	)
}

// NewMux 构造任务路由 mux
func NewMux() *asynq.ServeMux {
	return asynq.NewServeMux()
}
