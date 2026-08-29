// Package appfiles 提供 LICENSE.md / CHANGELOG.md 等随包文档的读取。
// 按工作目录向上查找，Docker 镜像中它们被 COPY 到工作目录。
package appfiles

import (
	"os"
	"path/filepath"
)

// FindFile 从工作目录及其最多 3 级父目录查找文件（兼容从 server-go/ 子目录启动的场景）。
func FindFile(name string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for i := 0; i < 4; i++ {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", os.ErrNotExist
}

// ReadFile 读取随包文档内容，找不到时返回空串。
func ReadFile(name string) string {
	p, err := FindFile(name)
	if err != nil {
		return ""
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return string(b)
}
