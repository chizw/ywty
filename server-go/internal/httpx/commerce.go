package httpx

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/chizw/ywty/server-go/internal/authx"
	"github.com/chizw/ywty/server-go/internal/model"
	"github.com/chizw/ywty/server-go/internal/orderx"
	"github.com/chizw/ywty/server-go/internal/pagination"
	"github.com/chizw/ywty/server-go/internal/payment"
	"github.com/chizw/ywty/server-go/internal/support/r"
	"github.com/go-chi/chi/v5"
)

// ---------- GET /api/v2/plans（公共） ----------

func (d *deps) handlePlansIndex(w http.ResponseWriter, req *http.Request) {
	p := pagination.FromRequest(req)
	var total int64
	d.gdb.Raw("SELECT count(*) FROM `plans` WHERE `is_up` = 1 AND `deleted_at` IS NULL").Scan(&total)
	var rows []model.Plan
	d.gdb.Raw("SELECT * FROM `plans` WHERE `is_up` = 1 AND `deleted_at` IS NULL ORDER BY `sort` ASC, `id` ASC LIMIT ? OFFSET ?",
		p.PerPage, p.Offset()).Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, map[string]any{
			"id": rows[i].ID, "type": rows[i].Type, "name": rows[i].Name,
			"intro": rows[i].Intro, "features": jsonOrNull(rows[i].Features), "badge": rows[i].Badge,
		})
	}
	r.Success(w, pagination.New(out, total, p))
}

// GET /api/v2/plans/{id}
func (d *deps) handlePlanShow(w http.ResponseWriter, req *http.Request) {
	id := pathInt(req, "id")
	var plan model.Plan
	if err := d.gdb.Where("is_up = 1 AND deleted_at IS NULL").First(&plan, id).Error; err != nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	var prices []model.PlanPrice
	d.gdb.Where("plan_id = ?", plan.ID).Order("price ASC").Find(&prices)
	priceList := make([]map[string]any, 0, len(prices))
	for _, pr := range prices {
		priceList = append(priceList, map[string]any{"id": pr.ID, "name": pr.Name, "duration": pr.Duration, "price": pr.Price})
	}
	r.Success(w, map[string]any{
		"id": plan.ID, "type": plan.Type, "name": plan.Name, "intro": plan.Intro,
		"features": jsonOrNull(plan.Features), "badge": plan.Badge, "prices": priceList,
	})
}

// ---------- 订单（auth） ----------

// GET /api/v2/user/orders
func (d *deps) handleUserOrders(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	p := pagination.FromRequest(req)

	where := "o.`user_id` = ?"
	args := []any{u.ID}
	order := "`created_at` DESC"
	if q := p.Q; q != "" {
		var keywords []string
		for _, part := range strings.Fields(q) {
			switch part {
			case "sort:amount:ascend":
				order = "`amount` ASC"
			case "sort:amount:descend":
				order = "`amount` DESC"
			case "sort:status:ascend":
				order = "`status` ASC"
			case "sort:status:descend":
				order = "`status` DESC"
			case "sort:created_at:ascend":
				order = "`created_at` ASC"
			case "sort:created_at:descend":
				order = "`created_at` DESC"
			default:
				keywords = append(keywords, part)
			}
		}
		if len(keywords) > 0 {
			where += " AND o.`trade_no` LIKE ?"
			args = append(args, "%"+strings.Join(keywords, " ")+"%")
		}
	}

	var total int64
	d.gdb.Raw("SELECT count(*) FROM `orders` o WHERE "+where, args...).Scan(&total)
	var rows []model.Order
	d.gdb.Raw("SELECT o.* FROM `orders` o WHERE "+where+" ORDER BY "+order+" LIMIT ? OFFSET ?",
		append(append([]any{}, args...), p.PerPage, p.Offset())...).Scan(&rows)
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		out = append(out, orderx.OrderResource(d.gdb, &rows[i]))
	}
	r.Success(w, pagination.New(out, total, p))
}

// POST /api/v2/orders/preview
func (d *deps) handleOrderPreview(w http.ResponseWriter, req *http.Request) {
	var body struct {
		PriceID    int64   `json:"price_id"`
		CouponCode *string `json:"coupon_code"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	pv, err := orderx.Preview(d.gdb, body.PriceID, body.CouponCode)
	if err != nil {
		r.Error(w, err.Error())
		return
	}
	r.Success(w, map[string]any{"amount": pv.Amount, "deduct_amount": pv.DeductAmount})
}

// POST /api/v2/user/orders（创建）
func (d *deps) handleOrderStore(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	var body struct {
		PriceID    int64   `json:"price_id"`
		CouponCode *string `json:"coupon_code"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	order, isPaid, err := orderx.Create(d.gdb, d.cfg, u.ID, body.PriceID, body.CouponCode)
	if err != nil {
		r.Error(w, err.Error())
		return
	}
	if !isPaid {
		// 延时 1 小时自动取消
		_ = d.queue.DispatchAt("cancel_order", map[string]any{"order_id": order.ID}, time.Now().Add(time.Hour).Unix())
	}
	r.Created(w, map[string]any{"trade_no": order.TradeNo, "is_paid": isPaid})
}

// GET /api/v2/user/orders/{trade_no}
func (d *deps) handleOrderShow(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	order, err := orderx.GetByTradeNo(d.gdb, u.ID, chi.URLParam(req, "trade_no"))
	if err != nil {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	r.Success(w, orderx.OrderResource(d.gdb, order))
}

// PUT /api/v2/orders/{trade_no}/cancel
func (d *deps) handleOrderCancel(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	order, err := orderx.GetByTradeNo(d.gdb, u.ID, chi.URLParam(req, "trade_no"))
	if err != nil || order.Status != "unpaid" {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	if err := orderx.Cancel(d.gdb, order); err != nil {
		r.Error(w, "取消失败")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /api/v2/user/orders/{trade_no}
func (d *deps) handleOrderDestroy(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	order, err := orderx.GetByTradeNo(d.gdb, u.ID, chi.URLParam(req, "trade_no"))
	if err != nil || order.Status != "cancelled" {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	d.gdb.Delete(order)
	w.WriteHeader(http.StatusNoContent)
}

// POST /api/v2/orders/{trade_no}/pay
func (d *deps) handleOrderPay(w http.ResponseWriter, req *http.Request) {
	u := authx.From(req).User
	var body struct {
		Platform  string `json:"platform"`
		Channel   string `json:"channel"`
		Method    string `json:"method"`
		ReturnURL string `json:"return_url"`
		CancelURL string `json:"cancel_url"`
	}
	if err := readBody(req, &body); err != nil {
		r.Error(w, "请求体解析失败")
		return
	}
	if !validPaymentEnum(body.Platform, body.Channel, body.Method) {
		r.ErrorData(w, http.StatusUnprocessableEntity, "The given data was invalid.",
			map[string]any{"errors": map[string][]string{"platform": {"支付平台 不存在。"}}})
		return
	}
	order, err := orderx.GetByTradeNo(d.gdb, u.ID, chi.URLParam(req, "trade_no"))
	if err != nil || order.Status != "unpaid" {
		r.ErrorWithCode(w, http.StatusNotFound, "Not Found")
		return
	}
	ctx := authx.From(req)
	if ctx.Group == nil {
		r.Error(w, "系统未初始化角色组")
		return
	}

	// 组的支付驱动中按 provider 匹配（options.provider）
	var drivers []struct {
		ID      int64
		Options *string
	}
	d.gdb.Raw(
		"SELECT dr.`id`, dr.`options` FROM `drivers` dr "+
			"INNER JOIN `group_driver` gd ON gd.driver_id = dr.id AND gd.type = 'payment' "+
			"WHERE gd.group_id = ? AND dr.`type` = 'payment' AND dr.`deleted_at` IS NULL "+
			"ORDER BY gd.`sort` ASC", ctx.Group.ID,
	).Scan(&drivers)

	var driverID int64
	var driverOptions map[string]any
	for _, dr := range drivers {
		if dr.Options == nil {
			continue
		}
		opts := map[string]any{}
		if jsonUnmarshalInto(*dr.Options, &opts) != nil {
			continue
		}
		if p, _ := opts["provider"].(string); p == body.Platform {
			driverID = dr.ID
			driverOptions = opts
			break
		}
	}
	if driverID == 0 {
		r.Error(w, "未配置支付驱动，请联系管理员")
		return
	}

	// 银联/paypal/网银重复提交需重新生成支付订单号
	if body.Channel == "unipay" || body.Channel == "paypal" || body.Channel == "bank" {
		newNo := orderx.GenerateOutTradeNo(d.gdb)
		d.gdb.Model(&model.Order{}).Where("id = ?", order.ID).Update("out_trade_no", newNo)
		order.OutTradeNo = newNo
	}

	// 组装配置
	config := make(map[string]any, len(driverOptions)+3)
	for k, v := range driverOptions {
		config[k] = v
	}
	if body.ReturnURL != "" {
		config["return_url"] = body.ReturnURL
	}
	if body.CancelURL != "" {
		config["cancel_url"] = body.CancelURL
	}
	config["notify_url"] = strings.TrimRight(d.cfg.AppURL, "/") +
		"/api/v2/payment/callback/" + itoa(int(driverID)) + "/" + order.OutTradeNo

	driver, err := payment.New(body.Platform, config)
	if err != nil {
		r.Error(w, err.Error())
		return
	}
	result, err := driver.CreateOrder(payment.CreateOrderDTO{
		OutTradeNo: order.OutTradeNo, Subject: "购买套餐",
		Amount: order.Amount, ClientIP: authx.ClientIP(req),
	}, body.Channel, body.Method)
	if err != nil {
		r.Error(w, err.Error())
		return
	}
	_ = d.gdb.Model(&model.Order{}).Where("id = ?", order.ID).Update("pay_method", body.Platform)
	r.Created(w, result)
}

// ANY /api/v2/payment/callback/{id}/{out_trade_no}（NotifyController）
func (d *deps) handlePaymentNotify(w http.ResponseWriter, req *http.Request) {
	driverID := pathInt(req, "id")
	outTradeNo := chi.URLParam(req, "out_trade_no")

	var order model.Order
	if err := d.gdb.Where("out_trade_no = ?", outTradeNo).First(&order).Error; err != nil {
		r.ErrorWithCode(w, http.StatusNotFound, "订单不存在")
		return
	}
	if order.Status != "unpaid" {
		r.ErrorWithCode(w, http.StatusInternalServerError, "订单已支付")
		return
	}
	var drv struct {
		Options *string
	}
	d.gdb.Raw("SELECT `options` FROM `drivers` WHERE `id` = ? LIMIT 1", driverID).Scan(&drv)
	if drv.Options == nil {
		r.ErrorWithCode(w, http.StatusInternalServerError, "支付驱动不存在")
		return
	}
	opts := map[string]any{}
	if jsonUnmarshalInto(*drv.Options, &opts) != nil {
		r.ErrorWithCode(w, http.StatusInternalServerError, "支付驱动配置错误")
		return
	}
	provider, _ := opts["provider"].(string)
	driver, err := payment.New(provider, opts)
	if err != nil {
		r.ErrorWithCode(w, http.StatusInternalServerError, err.Error())
		return
	}
	ok, verr := driver.VerifyNotify(req)
	if !ok {
		if verr != nil {
			r.ErrorWithCode(w, http.StatusInternalServerError, verr.Error())
			return
		}
		r.ErrorWithCode(w, http.StatusInternalServerError, "回调验证失败")
		return
	}
	if err := orderx.Complete(d.gdb, d.cfg, &order, provider); err != nil {
		slog.Warn("支付回调完成订单失败", "trade_no", order.TradeNo, "err", err)
		r.ErrorWithCode(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(driver.VerifyReturnBody()))
}

// validPaymentEnum 校验 platform/channel/method 枚举。
func validPaymentEnum(platform, channel, method string) bool {
	platforms := map[string]bool{"alipay": true, "wechat": true, "unipay": true, "paypal": true, "epay": true}
	channels := map[string]bool{"alipay": true, "wechat": true, "unipay": true, "paypal": true,
		"wxpay": true, "usdt": true, "qqpay": true, "bank": true, "jdpay": true, "unified": true}
	methods := map[string]bool{"web": true, "h5": true, "app": true, "mini": true, "pos": true, "scan": true, "mp": true}
	return platforms[platform] && channels[channel] && methods[method]
}
