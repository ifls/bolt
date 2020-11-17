package bolt

import (
	"fmt"
	"os"
	"sort"
	"unsafe"
)

const pageHeaderSize = int(unsafe.Offsetof(((*page)(nil)).ptr)) // 16

const minKeysPerPage = 2

const branchPageElementSize = int(unsafe.Sizeof(branchPageElement{})) // 16B
const leafPageElementSize = int(unsafe.Sizeof(leafPageElement{}))     // 16B

const (
	branchPageFlag   = 0x01 // 分支页
	leafPageFlag     = 0x02 // 叶子页面
	metaPageFlag     = 0x04 // 存放db的元数据
	freelistPageFlag = 0x10 // 存放db的空闲page
)

const (
	bucketLeafFlag = 0x01 // 此标记的Bucket 指向子bucket
)

type pgid uint64

// 页面 4096B
type page struct {
	id       pgid
	flags    uint16  // 区分不同类型的 页面
	count    uint16  // 此页面中的数据个数
	overflow uint32  // 单个页面不够, 会分配多个
	ptr      uintptr // 存放数据的起始地址, 大小size是固定的, 所以范围也是固定的, node的数据存放在此位置
}

// typ returns a human readable page type string used for debugging.
func (p *page) typ() string {
	if (p.flags & branchPageFlag) != 0 {
		return "branch"
	} else if (p.flags & leafPageFlag) != 0 {
		return "leaf"
	} else if (p.flags & metaPageFlag) != 0 {
		return "meta"
	} else if (p.flags & freelistPageFlag) != 0 {
		return "freelist"
	}
	return fmt.Sprintf("unknown<%02x>", p.flags)
}

// meta returns a pointer to the metadata section of the page.
func (p *page) meta() *meta {
	return (*meta)(unsafe.Pointer(&p.ptr))
}

// leafPageElement retrieves the leaf node by index
func (p *page) leafPageElement(index uint16) *leafPageElement {
	n := &((*[0x7FFFFFF]leafPageElement)(unsafe.Pointer(&p.ptr)))[index]
	return n
}

// leafPageElements retrieves a list of leaf nodes.
func (p *page) leafPageElements() []leafPageElement {
	if p.count == 0 {
		return nil
	}
	return ((*[0x7FFFFFF]leafPageElement)(unsafe.Pointer(&p.ptr)))[:]
}

// branchPageElement retrieves the branch node by index
func (p *page) branchPageElement(index uint16) *branchPageElement {
	return &((*[0x7FFFFFF]branchPageElement)(unsafe.Pointer(&p.ptr)))[index]
}

// branchPageElements retrieves a list of branch nodes.
func (p *page) branchPageElements() []branchPageElement {
	if p.count == 0 {
		return nil
	}
	return ((*[0x7FFFFFF]branchPageElement)(unsafe.Pointer(&p.ptr)))[:]
}

// dump writes n bytes of the page to STDERR as hex output.
func (p *page) hexdump(n int) {
	buf := (*[maxAllocSize]byte)(unsafe.Pointer(p))[:n]
	fmt.Fprintf(os.Stderr, "%x\n", buf)
}

type pages []*page

func (s pages) Len() int           { return len(s) }
func (s pages) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s pages) Less(i, j int) bool { return s[i].id < s[j].id } // 根据id排序

// branchPageElement represents a node on a branch page.
type branchPageElement struct {
	pos   uint32
	ksize uint32
	pgid  pgid // branch node 的 value 是子节点的 page id，存放在 branchPageElement 里
}

// key returns a byte slice of the node key.
func (n *branchPageElement) key() []byte {
	buf := (*[maxAllocSize]byte)(unsafe.Pointer(n))
	// pos 是key相对element的偏移
	return (*[maxAllocSize]byte)(unsafe.Pointer(&buf[n.pos]))[:n.ksize]
}

// leafPageElement represents a node on a leaf page.
type leafPageElement struct {
	flags uint32 // 区分 subbucket 和 普通的value
	pos   uint32 // key 距离 leafPageElement 的位移
	ksize uint32
	vsize uint32
}

// key returns a byte slice of the node key.
func (n *leafPageElement) key() []byte {
	buf := (*[maxAllocSize]byte)(unsafe.Pointer(n))
	// pos 是key相对element的偏移
	return (*[maxAllocSize]byte)(unsafe.Pointer(&buf[n.pos]))[:n.ksize:n.ksize]
}

// value returns a byte slice of the node value.
func (n *leafPageElement) value() []byte {
	buf := (*[maxAllocSize]byte)(unsafe.Pointer(n))
	// pos 是key相对element的偏移
	return (*[maxAllocSize]byte)(unsafe.Pointer(&buf[n.pos+n.ksize]))[:n.vsize:n.vsize]
}

// PageInfo represents human readable information about a page.
type PageInfo struct {
	ID            int
	Type          string
	Count         int
	OverflowCount int
}

type pgids []pgid

func (s pgids) Len() int           { return len(s) }
func (s pgids) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s pgids) Less(i, j int) bool { return s[i] < s[j] }

// merge returns the sorted union of a and b.
func (a pgids) merge(b pgids) pgids {
	// Return the opposite slice if one is nil.
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	merged := make(pgids, len(a)+len(b))
	mergepgids(merged, a, b)
	return merged
}

// mergepgids copies the sorted union of a and b into dst.
// If dst is too small, it panics.
// 有序合并数组
func mergepgids(dst, a, b pgids) {
	if len(dst) < len(a)+len(b) {
		panic(fmt.Errorf("mergepgids bad len %d < %d + %d", len(dst), len(a), len(b)))
	}
	// Copy in the opposite slice if one is nil.
	if len(a) == 0 {
		copy(dst, b)
		return
	}
	if len(b) == 0 {
		copy(dst, a)
		return
	}

	// Merged will hold all elements from both lists.
	merged := dst[:0]

	// Assign lead to the slice with a lower starting value, follow to the higher value.
	lead, follow := a, b
	if b[0] < a[0] {
		lead, follow = b, a
	}

	// Continue while there are elements in the lead.
	for len(lead) > 0 {
		// Merge largest prefix of lead that is ahead of follow[0].
		n := sort.Search(len(lead), func(i int) bool { return lead[i] > follow[0] })
		merged = append(merged, lead[:n]...)
		if n >= len(lead) {
			break
		}

		// Swap lead and follow.
		lead, follow = follow, lead[n:]
	}

	// Append what's left in follow.
	_ = append(merged, follow...)
}
