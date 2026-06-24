package model

// 用户状态
const (
	UserStatusNormal   = "normal"   // 正常
	UserStatusDisabled = "disabled" // 停用
	UserStatusPending  = "pending"  // 待激活
)

// 用户相关
const (
	CapacityFromSystem = "system" // 系统
	CapacityFromPlan   = "plan"   // 套餐
	CapacityFromAdmin  = "admin"  // 管理员调整
	CapacityFromOrder  = "order"  // 订单

	GroupFromSystem = "system"
	GroupFromPlan   = "plan"
	GroupFromAdmin  = "admin"
	GroupFromOrder  = "order"
)
