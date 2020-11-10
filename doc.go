/*
Package bolt implements a low-level key/value store in pure Go.
It supports fully serializable可串行化 transactions, ACID semantics,
and lock-free无锁 MVCC with multiple readers and a single writer. 一写多读
Bolt can be used for projects that want a simple data store without the need to add large dependencies such as
Postgres or MySQL. 简单可靠的存储引擎, 被etcd 和 consul使用

Bolt is a single-level, zero-copy, B+tree data store.
This means that Bolt is optimized for fast read access
and does not require recovery in the event of a system crash.
Transactions which have not finished committing will simply be rolled back in the event of a crash. 奔溃自行回滚??

The design of Bolt is based on Howard Chu's LMDB(Lightning Memory-Mapped Database) database project.

Bolt currently works on Windows, Mac OS X, and Linux.


Basics

There are only a few types in Bolt: DB, Bucket, Tx, and Cursor. 4个对象
The DB is a collection of buckets 桶集合 and is represented by a single file单文件 on disk.
A bucket is a collection of unique keys that are associated with values. 唯一 kv 集合

Transactions provide either read-only or read-write access to the database.
Read-only transactions can retrieve检索 key/value pairs and can use Cursors to iterate over the dataset sequentially.
Read-write transactions can create and delete buckets and can insert and remove keys.
Only one read-write transaction is allowed at a time. 读写事务,同时只能有一个


Caveats 警告

The database uses a read-only, memory-mapped data file只读的映射文件 to ensure that
applications cannot corrupt不能破坏 the database,
however, this means that keys and values returned from Bolt cannot be changed. 取得的数据无法修改
Writing to a read-only byte slice will cause Go to panic. 导致恐慌

Keys and values retrieved from the database are only valid for the life of the transaction. 数据只在事务期内有效, 否则是无效数据,或者panic
When used outside the transaction,
these byte slices can point to different data or can point to invalid memory which will cause a panic.

// 参考源码分析 https://youjiali1995.github.io/storage/boltdb/
// https://youjiali1995.github.io/categories/#storage
*/
package bolt
