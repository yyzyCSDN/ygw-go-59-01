package model

import "errors"

var (
	// ErrRangeExhausted 表示当前序列区间已用尽且没有可用新区间。
	ErrRangeExhausted = errors.New("sequence range exhausted")
	// ErrLeaseInvalid 表示租约不存在、过期或已转移，不能签发。
	ErrLeaseInvalid = errors.New("lease is not valid for issuing")
	// ErrNodeMissing 表示请求节点未注册或已离线。
	ErrNodeMissing = errors.New("node is not registered or inactive")
	// ErrPersistence 表示 WAL 持久化写入失败。
	ErrPersistence = errors.New("persistence write failed")
	// ErrRequestCancelled 表示签发请求已被调用方取消。
	ErrRequestCancelled = errors.New("issue request cancelled")
	// ErrDuplicateRange 表示重放时遇到重复区间。
	ErrDuplicateRange = errors.New("duplicate range record")
)
