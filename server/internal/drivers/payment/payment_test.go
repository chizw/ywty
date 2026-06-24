package payment

import (
	"net/url"
	"testing"
)

func TestEPaySignAndVerify(t *testing.T) {
	d := NewEPayDriver("1001", "secret_key_xyz", "https://pay.example.com", "https://notify", "https://return")

	params := map[string]string{
		"pid":          "1001",
		"type":         "alipay",
		"out_trade_no": "T20240101001",
		"notify_url":   "https://notify",
		"return_url":   "https://return",
		"name":         "test product",
		"money":        "9.99",
	}
	sig := d.sign(params)
	if sig == "" {
		t.Fatal("sign returned empty")
	}
	if !d.verify(url.Values{
		"pid":          []string{"1001"},
		"type":         []string{"alipay"},
		"out_trade_no": []string{"T20240101001"},
		"notify_url":   []string{"https://notify"},
		"return_url":   []string{"https://return"},
		"name":         []string{"test product"},
		"money":        []string{"9.99"},
		"sign":         []string{sig},
	}, sig) {
		t.Fatal("verify should pass for correct sign")
	}
	if d.verify(url.Values{"sign": []string{sig}}, sig) {
		// 空参数验证应失败
		t.Fatal("verify should fail for empty params")
	}
}

func TestMockDriverCreatePay(t *testing.T) {
	d := NewMockDriver()
	res, err := d.CreatePay(nil, &Order{ID: "O1", Subject: "test", Amount: 9.99})
	if err != nil {
		t.Fatalf("CreatePay: %v", err)
	}
	if res.OrderID != "O1" {
		t.Fatalf("OrderID mismatch: %s", res.OrderID)
	}
	if res.PayURL == "" {
		t.Fatal("PayURL empty")
	}
}

func TestMockDriverVerifyNotify(t *testing.T) {
	d := NewMockDriver()
	p, err := d.VerifyNotify(nil, []byte("anything"), nil)
	if err != nil {
		t.Fatalf("VerifyNotify: %v", err)
	}
	if p.Status != "paid" {
		t.Fatalf("status expected paid, got %s", p.Status)
	}
}

func TestMockDriverRefund(t *testing.T) {
	d := NewMockDriver()
	r, err := d.Refund(nil, &RefundRequest{OrderID: "O1", RefundNo: "R1", Amount: 9.99})
	if err != nil {
		t.Fatalf("Refund: %v", err)
	}
	if !r.Accepted {
		t.Fatal("refund not accepted")
	}
}

func TestFactoryRegisterAndGet(t *testing.T) {
	f := NewFactory()
	md := NewMockDriver()
	f.Register(md)

	got, err := f.Get("mock")
	if err != nil {
		t.Fatalf("Get mock: %v", err)
	}
	if got.Name() != "mock" {
		t.Fatalf("name mismatch: %s", got.Name())
	}

	if _, err := f.Get("unknown"); err == nil {
		t.Fatal("expected error for unknown driver")
	}
}

func TestFactoryList(t *testing.T) {
	f := NewFactory()
	f.Register(NewMockDriver())
	list := f.List()
	if len(list) != 1 || list[0] != "mock" {
		t.Fatalf("list mismatch: %v", list)
	}
}
