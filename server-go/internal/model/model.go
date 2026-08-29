// Package model 定义 GORM 模型。列名与 PHP 版数据库完全一致（"完全兼容"约束）。
package model

import (
	"time"

	"github.com/chizw/ywty/server-go/internal/db/types"
	"gorm.io/gorm"
)

// ---------- users ----------

type User struct {
	ID              int64          `gorm:"primaryKey;column:id"`
	Avatar          string         `gorm:"column:avatar"`
	Name            string         `gorm:"column:name"`
	Username        string         `gorm:"column:username"`
	Phone           *string        `gorm:"column:phone"`
	Email           *string        `gorm:"column:email"`
	Password        string         `gorm:"column:password"`
	Location        string         `gorm:"column:location"`
	URL             string         `gorm:"column:url"`
	Company         string         `gorm:"column:company"`
	CompanyTitle    string         `gorm:"column:company_title"`
	Tagline         string         `gorm:"column:tagline"`
	Bio             string         `gorm:"column:bio"`
	Interests       types.JSON     `gorm:"column:interests"`
	Socials         types.JSON     `gorm:"column:socials"`
	PhoneVerifiedAt *time.Time     `gorm:"column:phone_verified_at"`
	EmailVerifiedAt *time.Time     `gorm:"column:email_verified_at"`
	RememberToken   *string        `gorm:"column:remember_token"`
	IsAdmin         bool           `gorm:"column:is_admin"`
	Options         types.JSON     `gorm:"column:options"`
	LoginIP         *string        `gorm:"column:login_ip"`
	RegisterIP      *string        `gorm:"column:register_ip"`
	CountryCode     *string        `gorm:"column:country_code"`
	Status          string         `gorm:"column:status"`
	CreatedAt       *time.Time     `gorm:"column:created_at"`
	UpdatedAt       *time.Time     `gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (User) TableName() string { return "users" }

// ---------- groups ----------

type Group struct {
	ID        int64          `gorm:"primaryKey;column:id"`
	Name      string         `gorm:"column:name"`
	Intro     string         `gorm:"column:intro"`
	Options   types.JSON     `gorm:"column:options"`
	IsDefault bool           `gorm:"column:is_default"`
	IsGuest   bool           `gorm:"column:is_guest"`
	CreatedAt *time.Time     `gorm:"column:created_at"`
	UpdatedAt *time.Time     `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Group) TableName() string { return "groups" }

// ---------- storages ----------

type Storage struct {
	ID        int64          `gorm:"primaryKey;column:id"`
	Name      string         `gorm:"column:name"`
	Intro     string         `gorm:"column:intro"`
	Prefix    string         `gorm:"column:prefix"`
	Provider  string         `gorm:"column:provider"`
	Options   types.JSON     `gorm:"column:options"`
	CreatedAt *time.Time     `gorm:"column:created_at"`
	UpdatedAt *time.Time     `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Storage) TableName() string { return "storages" }

// ---------- group_storage ----------

type GroupStorage struct {
	GroupID   int64 `gorm:"primaryKey;column:group_id"`
	StorageID int64 `gorm:"primaryKey;column:storage_id"`
	Sort      int32 `gorm:"column:sort"`
}

func (GroupStorage) TableName() string { return "group_storage" }

// ---------- pages ----------

type Page struct {
	ID          int64      `gorm:"primaryKey;column:id"`
	Type        string     `gorm:"column:type"`
	Name        string     `gorm:"column:name"`
	Icon        string     `gorm:"column:icon"`
	Title       string     `gorm:"column:title"`
	Content     *string    `gorm:"column:content"`
	Keywords    *string    `gorm:"column:keywords"`
	Description *string    `gorm:"column:description"`
	Slug        string     `gorm:"column:slug"`
	URL         string     `gorm:"column:url"`
	ViewCount   int64      `gorm:"column:view_count"`
	Sort        int32      `gorm:"column:sort"`
	IsShow      bool       `gorm:"column:is_show"`
	CreatedAt   *time.Time `gorm:"column:created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at"`
}

func (Page) TableName() string { return "pages" }

// ---------- user_groups ----------

type UserGroup struct {
	ID        int64          `gorm:"primaryKey;column:id"`
	UserID    int64          `gorm:"column:user_id"`
	GroupID   int64          `gorm:"column:group_id"`
	OrderID   *int64         `gorm:"column:order_id"`
	From      string         `gorm:"column:from"`
	ExpiredAt *time.Time     `gorm:"column:expired_at"`
	CreatedAt *time.Time     `gorm:"column:created_at"`
	UpdatedAt *time.Time     `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (UserGroup) TableName() string { return "user_groups" }

// ---------- user_capacities ----------

type UserCapacity struct {
	ID        int64          `gorm:"primaryKey;column:id"`
	UserID    int64          `gorm:"column:user_id"`
	OrderID   *int64         `gorm:"column:order_id"`
	Capacity  float64        `gorm:"column:capacity"`
	From      string         `gorm:"column:from"`
	ExpiredAt *time.Time     `gorm:"column:expired_at"`
	CreatedAt *time.Time     `gorm:"column:created_at"`
	UpdatedAt *time.Time     `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (UserCapacity) TableName() string { return "user_capacities" }

// ---------- feedbacks ----------

type Feedback struct {
	ID        int64      `gorm:"primaryKey;column:id"`
	Type      string     `gorm:"column:type"`
	Title     string     `gorm:"column:title"`
	Name      string     `gorm:"column:name"`
	Email     string     `gorm:"column:email"`
	Content   string     `gorm:"column:content"`
	IPAddress *string    `gorm:"column:ip_address"`
	CreatedAt *time.Time `gorm:"column:created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at"`
}

func (Feedback) TableName() string { return "feedbacks" }

// ---------- photos ----------

type Photo struct {
	ID        int64          `gorm:"primaryKey;column:id"`
	UserID    *int64         `gorm:"column:user_id"`
	GroupID   *int64         `gorm:"column:group_id"`
	StorageID *int64         `gorm:"column:storage_id"`
	Name      string         `gorm:"column:name"`
	Intro     string         `gorm:"column:intro"`
	Filename  string         `gorm:"column:filename"`
	Pathname  string         `gorm:"column:pathname"`
	Mimetype  string         `gorm:"column:mimetype"`
	Extension string         `gorm:"column:extension"`
	MD5       string         `gorm:"column:md5"`
	SHA1      string         `gorm:"column:sha1"`
	Exif      types.JSON     `gorm:"column:exif"`
	Size      float64        `gorm:"column:size"`
	Width     int64          `gorm:"column:width"`
	Height    int64          `gorm:"column:height"`
	IsPublic  bool           `gorm:"column:is_public"`
	Status    string         `gorm:"column:status"`
	IPAddress *string        `gorm:"column:ip_address"`
	ExpiredAt *time.Time     `gorm:"column:expired_at"`
	CreatedAt *time.Time     `gorm:"column:created_at"`
	UpdatedAt *time.Time     `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Photo) TableName() string { return "photos" }

// ---------- albums ----------

type Album struct {
	ID        int64          `gorm:"primaryKey;column:id"`
	UserID    *int64         `gorm:"column:user_id"`
	Name      string         `gorm:"column:name"`
	Intro     string         `gorm:"column:intro"`
	IsPublic  bool           `gorm:"column:is_public"`
	CreatedAt *time.Time     `gorm:"column:created_at"`
	UpdatedAt *time.Time     `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Album) TableName() string { return "albums" }

// ---------- personal_access_tokens ----------

type PersonalAccessToken struct {
	ID            int64      `gorm:"primaryKey;column:id"`
	TokenableType string     `gorm:"column:tokenable_type"`
	TokenableID   int64      `gorm:"column:tokenable_id"`
	Name          string     `gorm:"column:name"`
	Token         string     `gorm:"column:token"`
	Abilities     *string    `gorm:"column:abilities"`
	LastUsedAt    *time.Time `gorm:"column:last_used_at"`
	ExpiresAt     *time.Time `gorm:"column:expires_at"`
	CreatedAt     *time.Time `gorm:"column:created_at"`
	UpdatedAt     *time.Time `gorm:"column:updated_at"`
}

func (PersonalAccessToken) TableName() string { return "personal_access_tokens" }

// ---------- shares / shareables ----------

type Share struct {
	ID        int64      `gorm:"primaryKey;column:id"`
	UserID    int64      `gorm:"column:user_id"`
	Type      string     `gorm:"column:type"`
	Slug      string     `gorm:"column:slug"`
	Content   *string    `gorm:"column:content"`
	Password  string     `gorm:"column:password"`
	ViewCount int64      `gorm:"column:view_count"`
	ExpiredAt *time.Time `gorm:"column:expired_at"`
	CreatedAt *time.Time `gorm:"column:created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at"`
}

func (Share) TableName() string { return "shares" }

type Shareable struct {
	ID            int64  `gorm:"primaryKey;column:id"`
	ShareID       int64  `gorm:"column:share_id"`
	ShareableType string `gorm:"column:shareable_type"`
	ShareableID   int64  `gorm:"column:shareable_id"`
}

func (Shareable) TableName() string { return "shareables" }

// ---------- likes / reports / violations ----------

type Like struct {
	ID           int64      `gorm:"primaryKey;column:id"`
	UserID       int64      `gorm:"column:user_id"`
	LikeableType string     `gorm:"column:likeable_type"`
	LikeableID   int64      `gorm:"column:likeable_id"`
	CreatedAt    *time.Time `gorm:"column:created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at"`
}

func (Like) TableName() string { return "likes" }

type Report struct {
	ID             int64          `gorm:"primaryKey;column:id"`
	ReportUserID   *int64         `gorm:"column:report_user_id"`
	ReportableType string         `gorm:"column:reportable_type"`
	ReportableID   int64          `gorm:"column:reportable_id"`
	Content        *string        `gorm:"column:content"`
	Status         string         `gorm:"column:status"`
	HandledAt      *time.Time     `gorm:"column:handled_at"`
	IPAddress      *string        `gorm:"column:ip_address"`
	CreatedAt      *time.Time     `gorm:"column:created_at"`
	UpdatedAt      *time.Time     `gorm:"column:updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Report) TableName() string { return "reports" }

type Violation struct {
	ID        int64          `gorm:"primaryKey;column:id"`
	UserID    *int64         `gorm:"column:user_id"`
	PhotoID   *int64         `gorm:"column:photo_id"`
	Reason    string         `gorm:"column:reason"`
	Status    string         `gorm:"column:status"`
	HandledAt *time.Time     `gorm:"column:handled_at"`
	CreatedAt *time.Time     `gorm:"column:created_at"`
	UpdatedAt *time.Time     `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Violation) TableName() string { return "violations" }

// ---------- notices ----------

type Notice struct {
	ID        int64          `gorm:"primaryKey;column:id"`
	Title     string         `gorm:"column:title"`
	Content   *string        `gorm:"column:content"`
	ViewCount int64          `gorm:"column:view_count"`
	Sort      int32          `gorm:"column:sort"`
	CreatedAt *time.Time     `gorm:"column:created_at"`
	UpdatedAt *time.Time     `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Notice) TableName() string { return "notices" }

// ---------- plans / coupons / orders ----------

type Plan struct {
	ID        int64          `gorm:"primaryKey;column:id"`
	Type      string         `gorm:"column:type"`
	Name      string         `gorm:"column:name"`
	Intro     *string        `gorm:"column:intro"`
	Features  types.JSON     `gorm:"column:features"`
	Badge     string         `gorm:"column:badge"`
	Sort      int32          `gorm:"column:sort"`
	IsUp      bool           `gorm:"column:is_up"`
	CreatedAt *time.Time     `gorm:"column:created_at"`
	UpdatedAt *time.Time     `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Plan) TableName() string { return "plans" }

type PlanPrice struct {
	ID        int64      `gorm:"primaryKey;column:id"`
	PlanID    int64      `gorm:"column:plan_id"`
	Name      string     `gorm:"column:name"`
	Duration  int32      `gorm:"column:duration"`
	Price     int64      `gorm:"column:price"`
	CreatedAt *time.Time `gorm:"column:created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at"`
}

func (PlanPrice) TableName() string { return "plan_prices" }

type PlanGroup struct {
	ID      int64  `gorm:"primaryKey;column:id"`
	PlanID  int64  `gorm:"column:plan_id"`
	GroupID *int64 `gorm:"column:group_id"`
}

func (PlanGroup) TableName() string { return "plan_groups" }

type PlanCapacity struct {
	ID       int64   `gorm:"primaryKey;column:id"`
	PlanID   int64   `gorm:"column:plan_id"`
	Capacity float64 `gorm:"column:capacity"`
}

func (PlanCapacity) TableName() string { return "plan_capacities" }

type Coupon struct {
	ID         int64          `gorm:"primaryKey;column:id"`
	Type       string         `gorm:"column:type"`
	Name       string         `gorm:"column:name"`
	Code       string         `gorm:"column:code"`
	Value      float64        `gorm:"column:value"`
	UsageLimit int64          `gorm:"column:usage_limit"`
	ExpiredAt  *time.Time     `gorm:"column:expired_at"`
	CreatedAt  *time.Time     `gorm:"column:created_at"`
	UpdatedAt  *time.Time     `gorm:"column:updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Coupon) TableName() string { return "coupons" }

type Order struct {
	ID           int64      `gorm:"primaryKey;column:id"`
	PlanID       *int64     `gorm:"column:plan_id"`
	UserID       *int64     `gorm:"column:user_id"`
	CouponID     *int64     `gorm:"column:coupon_id"`
	TradeNo      string     `gorm:"column:trade_no"`
	OutTradeNo   string     `gorm:"column:out_trade_no"`
	Type         string     `gorm:"column:type"`
	Amount       int64      `gorm:"column:amount"`
	DeductAmount int64      `gorm:"column:deduct_amount"`
	Snapshot     types.JSON `gorm:"column:snapshot"`
	Product      types.JSON `gorm:"column:product"`
	PayMethod    string     `gorm:"column:pay_method"`
	Status       string     `gorm:"column:status"`
	PaidAt       *time.Time `gorm:"column:paid_at"`
	CanceledAt   *time.Time `gorm:"column:canceled_at"`
	CreatedAt    *time.Time `gorm:"column:created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at"`
}

func (Order) TableName() string { return "orders" }

type Ticket struct {
	ID        int64          `gorm:"primaryKey;column:id"`
	UserID    int64          `gorm:"column:user_id"`
	IssueNo   string         `gorm:"column:issue_no"`
	Title     string         `gorm:"column:title"`
	Level     string         `gorm:"column:level"`
	Status    string         `gorm:"column:status"`
	CreatedAt *time.Time     `gorm:"column:created_at"`
	UpdatedAt *time.Time     `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Ticket) TableName() string { return "tickets" }

type TicketReply struct {
	ID        int64          `gorm:"primaryKey;column:id"`
	TicketID  int64          `gorm:"column:ticket_id"`
	UserID    int64          `gorm:"column:user_id"`
	Content   string         `gorm:"column:content"`
	IsNotify  bool           `gorm:"column:is_notify"`
	ReadAt    *time.Time     `gorm:"column:read_at"`
	CreatedAt *time.Time     `gorm:"column:created_at"`
	UpdatedAt *time.Time     `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (TicketReply) TableName() string { return "ticket_replies" }

type Driver struct {
	ID        int64          `gorm:"primaryKey;column:id"`
	Type      string         `gorm:"column:type"`
	Name      string         `gorm:"column:name"`
	Intro     string         `gorm:"column:intro"`
	Options   types.JSON     `gorm:"column:options"`
	CreatedAt *time.Time     `gorm:"column:created_at"`
	UpdatedAt *time.Time     `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Driver) TableName() string { return "drivers" }
