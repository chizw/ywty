package main

// zap 字段工具，避免在多个 main 文件重复引入 zap 包
import "go.uber.org/zap"

func zString(k, v string) zap.Field  { return zap.String(k, v) }
func zInt(k string, v int) zap.Field { return zap.Int(k, v) }
func zapErr(err error) zap.Field     { return zap.Error(err) }
