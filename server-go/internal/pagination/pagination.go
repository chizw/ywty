// Package pagination 输出与 Laravel LengthAwarePaginator + Resource collection
// ->response()->getData() 相同的 JSON 结构：{data, links, meta}。
package pagination

import (
	"net/http"
	"strconv"
)

type Links []map[string]any

type Meta struct {
	CurrentPage int    `json:"current_page"`
	From        *int   `json:"from"`
	LastPage    int    `json:"last_page"`
	Path        string `json:"path"`
	PerPage     int    `json:"per_page"`
	To          *int   `json:"to"`
	Total       int64  `json:"total"`
}

type Page struct {
	Data  any   `json:"data"`
	Links Links `json:"links"`
	Meta  Meta  `json:"meta"`
}

// Params 解析 q / per_page / page。
type Params struct {
	Q       string
	PerPage int
	Page    int
	Path    string
}

// FromRequest 读取分页参数，默认 per_page=20（与 PHP paginate 默认一致）。
func FromRequest(req *http.Request) Params {
	p := Params{PerPage: 20, Page: 1, Path: req.URL.Path, Q: req.URL.Query().Get("q")}
	if v := req.URL.Query().Get("per_page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 && n <= 9999 {
			p.PerPage = n
		}
	}
	if v := req.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 {
			p.Page = n
		}
	}
	return p
}

// Offset 计算偏移。
func (p Params) Offset() int { return (p.Page - 1) * p.PerPage }

// New 组装分页响应。
func New(data any, total int64, p Params) *Page {
	lastPage := 1
	if p.PerPage > 0 {
		lastPage = int((total + int64(p.PerPage) - 1) / int64(p.PerPage))
	}
	var from, to *int
	if total > 0 {
		f := p.Offset() + 1
		t := p.Offset() + p.PerPage
		if int64(t) > total {
			t = int(total)
		}
		from, to = &f, &t
	}
	link := func(url *string, label string, active bool, page any) map[string]any {
		return map[string]any{"url": url, "label": label, "active": active}
	}
	sptr := func(s string) *string { return &s }
	links := Links{link(nil, "&laquo; Previous", false, nil)}
	for i := 1; i <= lastPage; i++ {
		u := p.Path + "?page=" + strconv.Itoa(i)
		links = append(links, link(sptr(u), strconv.Itoa(i), i == p.Page, i))
	}
	links = append(links, link(nil, "Next &raquo;", false, nil))
	return &Page{
		Data:  data,
		Links: links,
		Meta: Meta{
			CurrentPage: p.Page, From: from, LastPage: lastPage,
			Path: p.Path, PerPage: p.PerPage, To: to, Total: total,
		},
	}
}
