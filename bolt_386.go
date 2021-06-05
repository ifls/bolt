package bbolt

// maxMapSize represents the largest mmap size supported by Bolt. 最大内存映射大小
const maxMapSize = 0x7FFFFFFF // 2G 31bit

// maxAllocSize is the size used when creating array pointers.
const maxAllocSize = 0xFFFFFFF // 29bit  512M
