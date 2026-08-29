package db_test

import (
	"os"
	"time"
)

// testDir 自管临时目录：Windows 下 sqlite 句柄释放有延迟，TempDir 的强制清理会误报。
func testDir(t interface{ Cleanup(func()) }) string {
	dir, err := os.MkdirTemp("", "ywty-test-")
	if err != nil {
		panic(err)
	}
	t.Cleanup(func() {
		for i := 0; i < 10; i++ {
			if os.RemoveAll(dir) == nil {
				return
			}
			time.Sleep(200 * time.Millisecond)
		}
	})
	return dir
}
