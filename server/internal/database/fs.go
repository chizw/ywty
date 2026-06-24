package database

import "os"

// mkdirAll 简单封装（避免引入额外的内部 pkg 引用）
func mkdirAll(path string) error {
	if path == "" || path == "." {
		return nil
	}
	return os.MkdirAll(path, 0755)
}
