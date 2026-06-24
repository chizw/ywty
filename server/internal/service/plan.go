// Package service 套餐服务
package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	bizerr "github.com/ywty/server/internal/errors"
	"github.com/ywty/server/internal/model"
)

// PlanService 套餐服务
type PlanService struct {
	db *gorm.DB
}

// NewPlanService 创建套餐服务
func NewPlanService(db *gorm.DB) *PlanService {
	return &PlanService{db: db}
}

// AdminPlanReq 创建/更新套餐请求
type AdminPlanReq struct {
	Type     string   `json:"type" binding:"required,oneof=vip default"`
	Name     string   `json:"name" binding:"required"`
	Intro    string   `json:"intro"`
	Features []string `json:"features"`
	Badge    string   `json:"badge"`
	Sort     int      `json:"sort"`
	IsUp     bool     `json:"is_up"`
	Prices   []struct {
		Name     string `json:"name" binding:"required"`
		Duration int    `json:"duration" binding:"required"`
		Price    int    `json:"price" binding:"required,min=0"`
	} `json:"prices"`
	Capacities []int64 `json:"capacities"` // 容量(字节)
	GroupIDs   []uint64  `json:"group_ids"`  // 空=全部可用
}

// PlanDetail 套餐详情（含关联数据）
type PlanDetail struct {
	Plan       model.Plan
	Prices     []model.PlanPrice
	Capacities []int64
	GroupIDs   []uint64
}

type planCtxKey int

const (
	planCtxKeyIsAdmin planCtxKey = iota
	planCtxKeyGroupIDs
)

// WithPlanAdmin 在上下文中标记是否为管理员
func WithPlanAdmin(ctx context.Context, isAdmin bool) context.Context {
	return context.WithValue(ctx, planCtxKeyIsAdmin, isAdmin)
}

// WithPlanGroupIDs 在上下文中写入当前浏览者的角色组 ID 列表
func WithPlanGroupIDs(ctx context.Context, groupIDs []uint64) context.Context {
	return context.WithValue(ctx, planCtxKeyGroupIDs, groupIDs)
}

func planIsAdmin(ctx context.Context) bool {
	v, _ := ctx.Value(planCtxKeyIsAdmin).(bool)
	return v
}

func planGroupIDs(ctx context.Context) []uint64 {
	v, _ := ctx.Value(planCtxKeyGroupIDs).([]uint64)
	return v
}

// ResolveViewerGroupIDs 解析浏览者所属的角色组
// userID == 0 时取游客组
func (s *PlanService) ResolveViewerGroupIDs(ctx context.Context, userID uint64) ([]uint64, error) {
	if userID == 0 {
		var groups []model.Group
		if err := s.db.WithContext(ctx).Where("is_guest = ?", true).Find(&groups).Error; err != nil {
			return nil, err
		}
		out := make([]uint64, len(groups))
		for i, g := range groups {
			out[i] = g.ID
		}
		return out, nil
	}

	var rows []model.UserGroup
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]uint64, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.GroupID)
	}
	return out, nil
}

// List 套餐列表
// isAdmin=true 返回全部；否则仅返回上架且对当前用户组可见的套餐
func (s *PlanService) List(ctx context.Context, isAdmin bool) ([]model.Plan, error) {
	if isAdmin {
		var rows []model.Plan
		if err := s.db.WithContext(ctx).Order("sort ASC, id ASC").Find(&rows).Error; err != nil {
			return nil, err
		}
		return rows, nil
	}

	q := s.db.WithContext(ctx).Model(&model.Plan{}).Where("is_up = ?", true)
	groupIDs := planGroupIDs(ctx)
	if len(groupIDs) > 0 {
		hasAny := s.db.Select("1").Table("plan_groups").Where("plan_groups.plan_id = plans.id")
		inGroups := s.db.Select("1").Table("plan_groups").Where("plan_groups.plan_id = plans.id").Where("group_id IN ?", groupIDs)
		q = q.Where("NOT EXISTS (?) OR EXISTS (?)", hasAny, inGroups)
	} else {
		q = q.Where("NOT EXISTS (SELECT 1 FROM plan_groups pg WHERE pg.plan_id = plans.id)")
	}

	var rows []model.Plan
	if err := q.Order("sort ASC, id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// Get 套餐详情（含价格、容量、可用组）
// 非管理员会校验上架状态及用户组可见性
func (s *PlanService) Get(ctx context.Context, id uint64) (*PlanDetail, error) {
	var plan model.Plan
	q := s.db.WithContext(ctx).Where("id = ?", id)
	if !planIsAdmin(ctx) {
		q = q.Where("is_up = ?", true)
		groupIDs := planGroupIDs(ctx)
		if len(groupIDs) > 0 {
			hasAny := s.db.Select("1").Table("plan_groups").Where("plan_groups.plan_id = plans.id")
			inGroups := s.db.Select("1").Table("plan_groups").Where("plan_groups.plan_id = plans.id").Where("group_id IN ?", groupIDs)
			q = q.Where("NOT EXISTS (?) OR EXISTS (?)", hasAny, inGroups)
		} else {
			q = q.Where("NOT EXISTS (SELECT 1 FROM plan_groups pg WHERE pg.plan_id = plans.id)")
		}
	}
	if err := q.First(&plan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.ResourceNotFound
		}
		return nil, err
	}

	prices, _ := s.GetPrices(ctx, id)
	capacities, _ := s.GetCapacities(ctx, id)
	groupIDs, _ := s.GetGroupIDs(ctx, id)
	return &PlanDetail{
		Plan:       plan,
		Prices:     prices,
		Capacities: capacities,
		GroupIDs:   groupIDs,
	}, nil
}

// Create 创建套餐及关联数据
func (s *PlanService) Create(ctx context.Context, req AdminPlanReq) (*model.Plan, error) {
	plan := &model.Plan{
		Type:     req.Type,
		Name:     req.Name,
		Intro:    req.Intro,
		Features: toJSONSlice(req.Features),
		Badge:    req.Badge,
		Sort:     req.Sort,
		IsUp:     req.IsUp,
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(plan).Error; err != nil {
			return fmt.Errorf("create plan: %w", err)
		}
		if err := createPlanRelations(ctx, tx, plan.ID, req); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return plan, nil
}

// Update 更新套餐及关联数据
func (s *PlanService) Update(ctx context.Context, id uint64, req AdminPlanReq) (*model.Plan, error) {
	var plan model.Plan
	if err := s.db.WithContext(ctx).First(&plan, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.ResourceNotFound
		}
		return nil, err
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{
			"type":     req.Type,
			"name":     req.Name,
			"intro":    req.Intro,
			"features": toJSONSlice(req.Features),
			"badge":    req.Badge,
			"sort":     req.Sort,
			"is_up":    req.IsUp,
		}
		if err := tx.Model(&model.Plan{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return fmt.Errorf("update plan: %w", err)
		}

		if err := tx.Where("plan_id = ?", id).Delete(&model.PlanPrice{}).Error; err != nil {
			return fmt.Errorf("delete plan prices: %w", err)
		}
		if err := tx.Where("plan_id = ?", id).Delete(&model.PlanCapacity{}).Error; err != nil {
			return fmt.Errorf("delete plan capacities: %w", err)
		}
		if err := tx.Where("plan_id = ?", id).Delete(&model.PlanGroup{}).Error; err != nil {
			return fmt.Errorf("delete plan groups: %w", err)
		}

		if err := createPlanRelations(ctx, tx, id, req); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, id)
}

// Delete 删除套餐及关联数据
func (s *PlanService) Delete(ctx context.Context, id uint64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Delete(&model.Plan{}, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return bizerr.ResourceNotFound
		}
		_ = tx.Where("plan_id = ?", id).Delete(&model.PlanPrice{}).Error
		_ = tx.Where("plan_id = ?", id).Delete(&model.PlanCapacity{}).Error
		_ = tx.Where("plan_id = ?", id).Delete(&model.PlanGroup{}).Error
		return nil
	})
}

// ToggleUp 上下架切换
func (s *PlanService) ToggleUp(ctx context.Context, id uint64) error {
	var plan model.Plan
	if err := s.db.WithContext(ctx).First(&plan, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return bizerr.ResourceNotFound
		}
		return err
	}
	return s.db.WithContext(ctx).Model(&model.Plan{}).Where("id = ?", id).Update("is_up", !plan.IsUp).Error
}

// GetByID 根据 ID 查询（管理后台内部使用，不限制上下架）
func (s *PlanService) GetByID(ctx context.Context, id uint64) (*model.Plan, error) {
	var p model.Plan
	if err := s.db.WithContext(ctx).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.ResourceNotFound
		}
		return nil, err
	}
	return &p, nil
}

// GetPrices 获取套餐价格列表
func (s *PlanService) GetPrices(ctx context.Context, planID uint64) ([]model.PlanPrice, error) {
	var rows []model.PlanPrice
	if err := s.db.WithContext(ctx).Where("plan_id = ?", planID).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// GetCapacities 获取套餐容量列表（字节）
func (s *PlanService) GetCapacities(ctx context.Context, planID uint64) ([]int64, error) {
	var rows []model.PlanCapacity
	if err := s.db.WithContext(ctx).Where("plan_id = ?", planID).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]int64, len(rows))
	for i, r := range rows {
		out[i] = r.Capacity
	}
	return out, nil
}

// GetGroupIDs 获取套餐适用的群组 ID 列表
func (s *PlanService) GetGroupIDs(ctx context.Context, planID uint64) ([]uint64, error) {
	var rows []model.PlanGroup
	if err := s.db.WithContext(ctx).Where("plan_id = ?", planID).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]uint64, 0, len(rows))
	for _, r := range rows {
		if r.GroupID != 0 {
			out = append(out, r.GroupID)
		}
	}
	return out, nil
}

func createPlanRelations(ctx context.Context, tx *gorm.DB, planID uint64, req AdminPlanReq) error {
	if len(req.Prices) > 0 {
		prices := make([]model.PlanPrice, 0, len(req.Prices))
		for _, p := range req.Prices {
			prices = append(prices, model.PlanPrice{
				PlanID:   planID,
				Name:     p.Name,
				Duration: p.Duration,
				Price:    p.Price,
			})
		}
		if err := tx.WithContext(ctx).Create(&prices).Error; err != nil {
			return fmt.Errorf("create plan prices: %w", err)
		}
	}

	if len(req.Capacities) > 0 {
		caps := make([]model.PlanCapacity, 0, len(req.Capacities))
		for _, c := range req.Capacities {
			caps = append(caps, model.PlanCapacity{
				PlanID:   planID,
				Capacity: int64(c),
			})
		}
		if err := tx.WithContext(ctx).Create(&caps).Error; err != nil {
			return fmt.Errorf("create plan capacities: %w", err)
		}
	}

	if len(req.GroupIDs) > 0 {
		groups := make([]model.PlanGroup, 0, len(req.GroupIDs))
		for _, gid := range req.GroupIDs {
			groups = append(groups, model.PlanGroup{
				PlanID:  planID,
				GroupID: gid,
			})
		}
		if err := tx.WithContext(ctx).Create(&groups).Error; err != nil {
			return fmt.Errorf("create plan groups: %w", err)
		}
	}

	return nil
}

func toJSONSlice(s []string) model.JSONSlice {
	if len(s) == 0 {
		return nil
	}
	out := make(model.JSONSlice, len(s))
	for i, v := range s {
		out[i] = v
	}
	return out
}
