// Package queue 基于 jobs 表的数据库队列（Laravel database queue 兼容）。
// Go 版任务 payload 为 {"job": "...", "data": {...}}；遇到 PHP 版遗留 payload
// （含 displayName 字段）时记入 failed_jobs 并跳过。
package queue

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Handler func(data []byte) error

type Queue struct {
	db         *gorm.DB
	handlers   map[string]Handler
	retryAfter int
	maxTries   int
	stop       chan struct{}
}

func New(gdb *gorm.DB) *Queue {
	return &Queue{
		db:         gdb,
		handlers:   map[string]Handler{},
		retryAfter: 90,
		maxTries:   1, // 与 PHP queue:work 默认一致
		stop:       make(chan struct{}),
	}
}

// Register 注册任务处理器。
func (q *Queue) Register(name string, h Handler) {
	q.handlers[name] = h
}

type payload struct {
	Job  string          `json:"job"`
	Data json.RawMessage `json:"data"`
	Name string          `json:"displayName"` // PHP 版遗留 payload 标记
}

// DispatchAt 写入队列并指定可执行时间（延迟任务，秒级时间戳）。
func (q *Queue) DispatchAt(name string, data any, availableAt int64) error {
	var raw json.RawMessage
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			return err
		}
		raw = b
	} else {
		raw = json.RawMessage("{}")
	}
	body, err := json.Marshal(payload{Job: name, Data: raw})
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	return q.db.Exec(
		"INSERT INTO `jobs` (`queue`, `payload`, `attempts`, `reserved_at`, `available_at`, `created_at`) "+
			"VALUES ('default', ?, 0, NULL, ?, ?)", body, availableAt, now,
	).Error
}

// Dispatch 写入队列。
func (q *Queue) Dispatch(name string, data any) error {
	var raw json.RawMessage
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			return err
		}
		raw = b
	} else {
		raw = json.RawMessage("{}")
	}
	body, err := json.Marshal(payload{Job: name, Data: raw})
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	return q.db.Exec(
		"INSERT INTO `jobs` (`queue`, `payload`, `attempts`, `reserved_at`, `available_at`, `created_at`) "+
			"VALUES ('default', ?, 0, NULL, ?, ?)", body, now, now,
	).Error
}

// Start 启动轮询 worker（阻塞直到 Stop）。
func (q *Queue) Start(pollInterval time.Duration) {
	go func() {
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-q.stop:
				return
			case <-ticker.C:
				for q.popAndRun() {
					// 连续消费直到队列为空
				}
			}
		}
	}()
}

func (q *Queue) Stop() {
	close(q.stop)
}

// popAndRun 取一个任务执行；返回是否确实执行了任务。
func (q *Queue) popAndRun() bool {
	now := time.Now().Unix()
	var job struct {
		ID       int64
		Payload  string
		Attempts int64
	}
	err := q.db.Raw(
		"SELECT `id`, `payload`, `attempts` FROM `jobs` "+
			"WHERE `queue` = 'default' AND `available_at` <= ? AND (`reserved_at` IS NULL OR `reserved_at` <= ?) "+
			"ORDER BY `id` LIMIT 1", now, now-int64(q.retryAfter),
	).Scan(&job).Error
	if err != nil || job.ID == 0 {
		return false
	}
	// 乐观锁抢占
	res := q.db.Exec(
		"UPDATE `jobs` SET `reserved_at` = ?, `attempts` = `attempts` + 1 WHERE `id` = ? AND `reserved_at` IS NULL",
		now, job.ID,
	)
	if res.Error != nil || res.RowsAffected == 0 {
		return false
	}

	var p payload
	if err := json.Unmarshal([]byte(job.Payload), &p); err != nil {
		q.fail(job.ID, job.Payload, "payload 解析失败: "+err.Error())
		return true
	}
	if p.Name != "" {
		// PHP 版遗留任务，Go 无法执行
		q.fail(job.ID, job.Payload, "PHP 版遗留队列任务，Go 版已跳过: "+p.Name)
		return true
	}
	h, ok := q.handlers[p.Job]
	if !ok {
		q.fail(job.ID, job.Payload, "未注册的任务类型: "+p.Job)
		return true
	}
	if err := h(p.Data); err != nil {
		slog.Warn("队列任务执行失败", "job", p.Job, "id", job.ID, "err", err)
		if job.Attempts+1 >= int64(q.maxTries) {
			q.fail(job.ID, job.Payload, err.Error())
		} else {
			_ = q.db.Exec(
				"UPDATE `jobs` SET `reserved_at` = NULL, `available_at` = ? WHERE `id` = ?",
				time.Now().Unix()+int64(q.retryAfter), job.ID,
			).Error
		}
		return true
	}
	_ = q.db.Exec("DELETE FROM `jobs` WHERE `id` = ?", job.ID).Error
	return true
}

// fail 记入 failed_jobs 并删除原任务。
func (q *Queue) fail(id int64, payloadStr, exception string) {
	slog.Error("队列任务失败", "id", id, "err", exception)
	_ = q.db.Exec(
		"INSERT INTO `failed_jobs` (`uuid`, `connection`, `queue`, `payload`, `exception`, `failed_at`) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)",
		uuid.NewString(), "database", "default", payloadStr, exception,
	).Error
	_ = q.db.Exec("DELETE FROM `jobs` WHERE `id` = ?", id).Error
}
