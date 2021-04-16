// Package fw
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/13 3:57 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package fw

import (
	"minlib/component"
	"mir-go/daemon/common"
	"mir-go/daemon/lf"
	"mir-go/daemon/table"
)

const (
	DuplicateNonceNone     = 0      // 不存在重复
	DuplicateNonceInSame   = 1 << 0 // in-record 中存在一个来自相同 LogicFace 的重复
	DuplicateNonceInOther  = 1 << 1 // in-record 中存在一个来自不同 LogicFace 的重复
	DuplicateNonceOutSame  = 1 << 2 // out-record 中存在一个来自相同 LogicFace 的重复
	DuplicateNonceOutOther = 1 << 3 // out-record 中存在一个来自不同 LogicFace 的重复
)

// FindDuplicateNonce
// 查询 PIT 条目中是否存在某个 in-record / out-record 与刚收到的兴趣包中的 Nonce 重复
//
// @Description:
//
// @param entry
// @param nonce
// @param ingress
//
func FindDuplicateNonce(pitEntry *table.PITEntry, nonce *component.Nonce, ingress *lf.LogicFace) int {
	result := DuplicateNonceNone
	// 查看 in-record
	for _, inRecord := range pitEntry.GetInRecords() {
		if inRecord.LastNonce.GetNonce() == nonce.GetNonce() {
			if inRecord.LogicFace.LogicFaceId == ingress.LogicFaceId {
				result |= DuplicateNonceInSame
			} else {
				result |= DuplicateNonceInOther
			}
		}
	}

	// 查看 out-record
	for _, outRecord := range pitEntry.GetOutRecords() {
		if outRecord.LastNonce.GetNonce() == nonce.GetNonce() {
			if outRecord.LogicFace.LogicFaceId == ingress.LogicFaceId {
				result |= DuplicateNonceOutSame
			} else {
				result |= DuplicateNonceOutOther
			}
		}
	}
	return result
}

// HasPendingOutRecords
// 判断 PIT 条目中是否存在仍在 pending 的 out-record
//
// @Description:
// @param entry
// @return bool
//
func HasPendingOutRecords(entry *table.PITEntry) bool {
	if entry == nil {
		return false
	}
	now := common.GetCurrentTime()
	for _, outRecord := range entry.GetOutRecords() {
		if outRecord.ExpireTime > now && outRecord.NackHeader == nil {
			return true
		}
	}
	return false
}
