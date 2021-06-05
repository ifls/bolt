// +build !windows,!plan9,!solaris,!aix

package bbolt

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

// 系统调用 flock https://man7.org/linux/man-pages/man2/flock.2.html
// flock acquires an advisory lock on a file descriptor.
// 只是告知其他flock调用 这里有个锁， 但无法强制禁止其他进程读写文件
func flock(db *DB, exclusive bool, timeout time.Duration) error {
	var t time.Time
	if timeout != 0 {
		t = time.Now()
	}
	fd := db.file.Fd()
	flag := syscall.LOCK_NB // non-block
	if exclusive {
		flag |= syscall.LOCK_EX // 排他
	} else {
		flag |= syscall.LOCK_SH // 只读的话, 加共享锁就可以
	}
	for {
		// Attempt to obtain an exclusive lock.
		err := syscall.Flock(int(fd), flag)
		if err == nil {
			return nil
		} else if err != syscall.EWOULDBLOCK { // 如果无法完成操作，不会阻塞，而是返回此错误，这里使用循环超时，模拟阻塞
			return err
		}

		// If we timed out then return an error. 不够时间下一次重试，就返回超时错误
		if timeout != 0 && time.Since(t) > timeout-flockRetryTimeout {
			return ErrTimeout
		}

		// Wait for a bit and try again.
		time.Sleep(flockRetryTimeout)
	}
}

// funlock releases an advisory lock on a file descriptor.
func funlock(db *DB) error {
	return syscall.Flock(int(db.file.Fd()), syscall.LOCK_UN) // 解锁标记
}

// mmap memory maps a DB's data file.
func mmap(db *DB, sz int) error {
	// Map the data file to memory.
	// 保护模式，可读
	// MAP_SHARED 共享映射， 修改对其他进程可见
	b, err := syscall.Mmap(int(db.file.Fd()), 0, sz, syscall.PROT_READ, syscall.MAP_SHARED|db.MmapFlags)
	if err != nil {
		return err
	}

	// Advise the kernel that the mmap is accessed randomly.
	// 主要是用来改善性能，建议内核不需要预读之后的几页，因为不是连续访问，不太可能访问的到
	err = madvise(b, syscall.MADV_RANDOM)
	if err != nil && err != syscall.ENOSYS {
		// Ignore not implemented error in kernel because it still works.
		return fmt.Errorf("madvise: %s", err)
	}

	// Save the original byte slice and convert to a byte array pointer.
	db.dataref = b
	db.data = (*[maxMapSize]byte)(unsafe.Pointer(&b[0]))
	db.datasz = sz
	return nil
}

// munmap unmaps a DB's data file from memory.
func munmap(db *DB) error {
	// Ignore the unmap if we have no mapped data.
	if db.dataref == nil {
		return nil
	}

	// Unmap using the original byte slice.
	err := syscall.Munmap(db.dataref)
	db.dataref = nil
	db.data = nil
	db.datasz = 0
	return err
}

// NOTE: This function is copied from stdlib because it is not available on darwin.
func madvise(b []byte, advice int) (err error) {
	_, _, e1 := syscall.Syscall(syscall.SYS_MADVISE, uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)), uintptr(advice))
	if e1 != 0 {
		err = e1
	}
	return
}
