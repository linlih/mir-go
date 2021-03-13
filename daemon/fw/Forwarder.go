package fw

import (
	"minlib/component"
	"minlib/packet"
	"mir-go/daemon/lf"
	"mir-go/daemon/table"
)

//
// MIR 转发器实例
//
// @Description:
//
type Forwarder struct {
	table.PIT // 内嵌一个PIT表
	table.FIB // 内嵌一个FIB表
	table.CS  // 内嵌一个CS表
}

//
// 处理一个兴趣包到来 （ Incoming Interest Pipeline）
//
// @Description:
// 1. 首先给 Interest 的 TTL 减一，然后检查 TTL 的值是：
//	- TTL < 0 则认为该兴趣包是一个回环的 Interest ，直接将其传递给 Interest loop 管道进一步处理；
//	- TTL >= 0 则执行下一步。
//	问题：下面第3步通过PIT条目进行回环的 Interest 检测，那为什么这边还需要一个 TTL 的机制来检测回环呢？
//	 - 因为通过PIT条目对比 Nonce 的方式来检测循环存在一个问题，当一个兴趣包被转发出去，假设其对应的PIT条目的过期时间为 $x$ 秒，那如果这个兴趣
//	   包在经过 $y$ 秒之后回环到当前路由器（$y > x$），则此时该兴趣包对应的PIT条目已经被移除，路由器无法通过PIT聚合的方式来检测回环的兴趣包；
//   - 所以，为了解决上述问题，我们通过类比IP中简单的设置 TTL 的方式来检测上面描述的特殊情况下，不能检测回环 Interest 的问题。
//   - NDN通过引入 Dead Nonce List 来进行检测，我们觉得会引入复杂性，且是概率性可靠的，直接用 TTL 的方式可以实现简单，且减少查询
//     Dead Nonce List 的开销。
//
// 2. 根据传入的 Interest 创建一个PIT条目（如果存在同名的PIT条目，则直接使用，不存在则创建）。
//
// 3. 然后查询PIT条目中是否有和传入的 Interest 的 Nonce 相同，并且是从不同的 LogicFace 收到的入记录（ in-records ），如果找到匹配的入记
//    录，则认为传入的 Interest 是回环的循环 Interest ，直接将其传递给 Interest loop 管道进一步处理；否则执行下一步。
//  - 如果从同一个逻辑接口收到同名且 Nonce 相同的 Interest，则可能是同一个消费者发送的 Interest，该 Interst 被判定为合法的重传包；
//  - 如果从不同的逻辑接口收到同名且 Nonce 相同的 Interest，则可能是循环的 Interest 或者是同一个 Interest 沿着多个不同的路径到达，此时，将
//    传入的 Interest 判定为循环的 Interest ，触发 Interest loop 管道。
//
// 4. 然后通过查询 PIT 条目中的记录，判断当前 Interst 是否是未决的 （ pending ），如果传入的 Interest 对应的PIT条目包含其它记录，则认为
//    该 Interest 是未决的。
//
// 5. 如果 Interest 是未决的，则直接传递给 ContentStore miss 管道处理；如果 Interest 不是未决的，则查询CS，如果存在缓存，则传递给
//    ContentStore hit 管道进行进一步处理，否则传递给 Content miss 管道进行进一步的处理。
// @param ingress	入口Face
// @param interest	收到的内容兴趣包
//
func (f *Forwarder) OnIncomingInterest(ingress *lf.LogicFace, interest *packet.Interest) {
	// TTL 减一，并且检查 TTL 是否小于0，小于0则判定为循环兴趣包
	if interest.TTL.Minus() < 0 {
		f.OnInterestLoop(ingress, interest)
		return
	}

	// PIT insert
	// 此时如果PIT条目已存在，则返回之前创建的PIT条目；
	// 如果PIT条目不存在，会创建一个空条目（注意，此时只是创建PIT条目，并没有插入in-record）
	pitEntry := f.PIT.Insert(interest)

	// Detect duplicate Nonce in PIT entry
	// 存在从不同 LogicFace 收到的重复 Nonce，则认定为兴趣包重复，触发循环兴趣包处理流程
	findDuplicateNonceResult := FindDuplicateNonce(pitEntry, &interest.Nonce, ingress)
	if findDuplicateNonceResult&DuplicateNonceInOther == DuplicateNonceInOther {
		f.OnInterestLoop(ingress, interest)
		return
	}

	// is Pending ?
	if pitEntry.HasInRecords() {
		// 如果 PIT 条目中存在 in-record 则说明这是一个悬而未决（pending）的兴趣包，路由器中肯定没有对应的缓存
		// 所以直接执行内容缓存未命中逻辑
		// TODO: 其实有些兴趣包设置了 MustBeFresh = true，所以被转发了，这样就是 PIT 条目中存在 in-record，但是CS中存在不新鲜的缓存数据
		// TODO: 此时如果收到一个兴趣包，它的 MustBeFresh = false，是否要考虑执行 CS 查找
		f.OnContentStoreMiss(ingress, pitEntry, interest)
	} else {
		// CS Lookup
		if csEntry := f.CS.Find(interest); csEntry == nil {
			// 没有命中缓存
			f.OnContentStoreMiss(ingress, pitEntry, interest)
		} else {
			// 命中缓存
			f.OnContentStoreHit(ingress, pitEntry, interest, csEntry)
		}
	}
}

//
// 处理一个回环的兴趣包 （ Interest Loop Pipeline ）
//
// @Description:
//  在 Incoming Interest 管道处理过程中，如果检测到 Interest 循环就会触发 Interest loop 管道，本管道会向收到 Interest 的 LogicFace
//  发送一个原因为 "重复" （ duplicate ） 的 Nack。
// @param ingress
// @param interest
//
func (f *Forwarder) OnInterestLoop(ingress *lf.LogicFace, interest *packet.Interest) {
	// 创建一个原因为 duplicate 的Nack
	nack := packet.Nack{
		Interest: interest,
	}
	nack.SetNackReason(component.NackReasonDuplicate)

	// TODO: 将Nack通过Face发出
	// ingress.putNack(nack)
}

//
// 处理兴趣包未命中缓存 （ ContentStore Miss Pipeline ）
//
// @Description:
// @param ingress
// @param pitEntry
// @param interest
//
func (f *Forwarder) OnContentStoreMiss(ingress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest) {

	panic("implement me")
}

func (f *Forwarder) OnContentStoreHit(ingress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest, data *table.CSEntry) {
	panic("implement me")
}

func (f *Forwarder) OnOutgoingInterest(egress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest) {
	panic("implement me")
}

func (f *Forwarder) OnInterestFinalize(pitEntry *table.PITEntry) {
	panic("implement me")
}

func (f *Forwarder) OnIncomingData(ingress *lf.LogicFace, data *packet.Data) {
	panic("implement me")
}

func (f *Forwarder) OnDataUnsolicited(ingress *lf.LogicFace, data *packet.Data) {
	panic("implement me")
}

func (f *Forwarder) OnOutgoingData(egress *lf.LogicFace, data *packet.Data) {
	panic("implement me")
}

func (f *Forwarder) OnIncomingNack(ingress *lf.LogicFace, nack *packet.Nack) {
	panic("implement me")
}

func (f *Forwarder) OnOutgoingNack(egress *lf.LogicFace, pitEntry *table.PITEntry, header *component.NackHeader) {
	panic("implement me")
}

func (f *Forwarder) OnIncomingCPacket(ingress *lf.LogicFace, cPacket *packet.CPacket) {
	panic("implement me")
}

func (f *Forwarder) OnOutgoingCPacket(egress *lf.LogicFace, cPacket *packet.CPacket) {
	panic("implement me")
}
