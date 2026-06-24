package main

import "go.uber.org/zap"

func zString(k, v string) zap.Field { return zap.String(k, v) }
func zapErr(err error) zap.Field    { return zap.Error(err) }
