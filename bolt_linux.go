package bbolt

import (
	"syscall"
)

// fdatasync flushes written data to a file descriptor.
// 只对文件的数据存盘, 除非文件的size变了, 否则不存文件的元数据(访问和修改时间)
// 只在linux 系统上有??
func fdatasync(db *DB) error {
	return syscall.Fdatasync(int(db.file.Fd()))
}
