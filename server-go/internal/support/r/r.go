// Package r 统一 JSON 响应包。
//
// 契约：{"status": "success"|"error", "message": "...", "data": ..., "time": unix秒}
// message 为空串或 data 为 nil 时省略键。
// 注意：业务错误的 HTTP 状态码默认是 200，与原版行为一致。
package r

import (
	"encoding/json"
	"net/http"
	"time"
)

type Envelope struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Time    int64  `json:"time"`
}

func write(w http.ResponseWriter, statusCode int, status, message string, data any) {
	body := Envelope{Status: status, Message: message, Data: data, Time: time.Now().Unix()}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(body)
}

// Success 成功响应，HTTP 200，默认 message 为 "successful"。
func Success(w http.ResponseWriter, data any) {
	write(w, http.StatusOK, "success", "successful", data)
}

// SuccessWithMessage 成功响应并自定义 message。
func SuccessWithMessage(w http.ResponseWriter, message string, data any) {
	write(w, http.StatusOK, "success", message, data)
}

// Created 成功响应，HTTP 201。
func Created(w http.ResponseWriter, data any) {
	write(w, http.StatusCreated, "success", "successful", data)
}

// Error 业务错误：HTTP 200 + status=error（与原版默认行为一致）。
func Error(w http.ResponseWriter, message string) {
	write(w, http.StatusOK, "error", message, nil)
}

// ErrorWithCode 指定 HTTP 状态码的错误响应。
func ErrorWithCode(w http.ResponseWriter, statusCode int, message string) {
	write(w, statusCode, "error", message, nil)
}

// ErrorData 带附加数据的错误响应（如 422 校验错误的 errors 字段）。
func ErrorData(w http.ResponseWriter, statusCode int, message string, data any) {
	write(w, statusCode, "error", message, data)
}
