package main

// zap 字段工具
import "go.uber.org/zap"

func zString(k, v string) zap.Field    { return zap.String(k, v) }
func zBool(k string, v bool) zap.Field { return zap.Bool(k, v) }
func zInt(k string, v int) zap.Field   { return zap.Int(k, v) }
func zapErr(err error) zap.Field       { return zap.Error(err) }
