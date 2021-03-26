//
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/13 3:57 下午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package fw

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"minlib/component"
	"minlib/packet"
	"mir-go/daemon/common"
	"mir-go/daemon/lf"
	"mir-go/daemon/plugin"
	"mir-go/daemon/table"
	"time"
)

//
// MIR 转发器实例
//
// @Description:
//
type Forwarder struct {
	table.PIT                                       // 内嵌一个PIT表
	table.FIB                                       // 内嵌一个FIB表
	table.CS                                        // 内嵌一个CS表
	table.StrategyTable                             // 内嵌一个策略选择表
	pluginManager       *plugin.GlobalPluginManager // 插件管理器
}

//
// 初始化转发器
//
// @Description:
// @receiver f
//
func (f *Forwarder) Init(pluginManager *plugin.GlobalPluginManager) {
	// 初始化各个表
	f.PIT.Init()
	f.FIB.Init()
	f.CS.Init()
	f.StrategyTable.Init()
	f.pluginManager = pluginManager
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
	common.LogDebugWithFields(logrus.Fields{
		"faceId":   ingress.LogicFaceId,
		"interest": interest.ToUri(),
	}, "Detect Interest loop")

	// 调用插件锚点
	if f.pluginManager.OnIncomingInterest(ingress, interest) != 0 {
		return
	}
	fmt.Println("interest", interest.TTL)
	// TTL 减一，并且检查 TTL 是否小于0，小于0则判定为循环兴趣包
	if interest.TTL.Ttl() == 0 {
		f.OnInterestLoop(ingress, interest)
		return
	}
	interest.TTL.Minus()

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
			f.CS.Misses+=1
			f.OnContentStoreMiss(ingress, pitEntry, interest)
		} else {
			// 命中缓存
			f.CS.Hits+=1
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
	common.LogDebugWithFields(logrus.Fields{
		"faceId":   ingress.LogicFaceId,
		"interest": interest.ToUri(),
	}, "Detect Interest loop")

	// 调用插件锚点
	if f.pluginManager.OnInterestLoop(ingress, interest) != 0 {
		return
	}

	// 创建一个原因为 duplicate 的Nack
	nack := packet.Nack{
		Interest: interest,
	}
	nack.SetNackReason(component.NackReasonDuplicate)

	// 将Nack通过Face发出
	ingress.SendNack(&nack)
}

//
// 处理兴趣包未命中缓存 （ ContentStore Miss Pipeline ）
//
// @Description:
// 1. 首先根据传入的 Interest 以及对应的传入 LogicFace 在尝试在对应的PIT条目中插入一条 in-record；
//  - 如果对应的PIT条目中已经存在一个相同 LogicFace 的 in-record 记录（比如：下游正在重传同一个兴趣包），那只需要用收到的 Interest 中的
//    Nonce 和 InterestLifetime 来更新对应的 in-record 即可，如果没有指定 InterestLifetime，则默认为4s；
//  - 否则创建一个新的 in-record 记录插入到对应的PIT条目当中。
// 2. 然后将PIT条目的超时计时器设置为当前 PIT 条目中所有 in-record 最大剩余超时时间。
// 3. 然后传递给转发策略作转发决策，在转发策略中按需触发 Outgoing Interest 管道处理逻辑，将 Interest 转发出去。
// @param ingress
// @param pitEntry
// @param interest
//
func (f *Forwarder) OnContentStoreMiss(ingress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest) {
	common.LogDebugWithFields(logrus.Fields{
		"faceId":   ingress.LogicFaceId,
		"interest": interest.ToUri(),
	}, "ContentStore miss")

	// 调用插件锚点
	if f.pluginManager.OnContentStoreMiss(ingress, pitEntry, interest) != 0 {
		return
	}

	currentTime := GetCurrentTime()

	// insert in-record
	inRecord := pitEntry.InsertOrUpdateInRecord(ingress, interest)
	// TODO: 检查一下，这个设置超时时间的操作要不要放到插入 in-record 的内部进行
	inRecord.ExpireTime = currentTime + interest.InterestLifeTime.GetInterestLifeTime()

	// Set PIT Entry ExpiryTimer
	// 设置超时时间为所有 in-record 中最迟的超时时间
	maxTime := uint64(0)
	for _, inRecord := range pitEntry.GetInRecords() {
		if inRecord.ExpireTime > maxTime {
			maxTime = inRecord.ExpireTime
		}
	}
	fmt.Println("yb test1", maxTime)
	duration := maxTime - currentTime
	fmt.Println("duration", duration)
	if duration < 0 {
		duration = 0
	}

	fmt.Println("jkdasdk:", time.Duration(duration)*time.Millisecond)
	f.SetExpiryTime(pitEntry, time.Duration(duration)*time.Millisecond)

	// 查询当前兴趣包所匹配的策略，执行 AfterReceiveInterest 钩子
	if ste := f.StrategyTable.FindEffectiveStrategyEntry(interest.GetName()); ste != nil {
		ste.GetStrategy().AfterReceiveInterest(ingress, interest, pitEntry)
	} else {
		// 输出错误，兴趣包没有找到匹配的可用策略
		common.LogErrorWithFields(logrus.Fields{
			"interest": interest.ToUri(),
		}, "Not found effective StrategyBase")
	}
}

//
// 处理兴趣包命中缓存 （ ContentStore Hit Pipeline ）
//
// @Description:
//  在 incoming Interest 管道中执行 ContentStore 查找并找到匹配项之后触发 ContentStore hit 管道处理逻辑。此管道执行以下步骤：
//   1. 首先将 Interest 对应PIT条目的到期计时器设置为当前时间，这会使得计时器到期，触发 Interest finalize 管道；
//   2. 然后触发 Interest 对应策略的 StrategyBase::afterContentStoreHit 回调。
// @param ingress
// @param pitEntry
// @param interest
// @param data
//
func (f *Forwarder) OnContentStoreHit(ingress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest, data *table.CSEntry) {
	common.LogDebugWithFields(logrus.Fields{
		"faceId":   ingress.LogicFaceId,
		"interest": interest.ToUri(),
	}, "ContentStore hit")

	// 调用插件锚点
	if f.pluginManager.OnContentStoreHit(ingress, pitEntry, interest, data) != 0 {
		return
	}

	// 设置超时时间为当前时间
	f.SetExpiryTime(pitEntry, 0)

	if ste := f.StrategyTable.FindEffectiveStrategyEntry(interest.GetName()); ste != nil {
		ste.GetStrategy().AfterContentStoreHit(ingress, data.GetData(), pitEntry)
	} else {
		// 输出错误，兴趣包没有找到匹配的可用策略
		common.LogErrorWithFields(logrus.Fields{
			"interest": interest.ToUri(),
		}, "Not found effective StrategyBase")
	}
}

//
// 处理将兴趣包通过 LogicFace 发出 （ Outgoing Interest Pipeline ）
//
// @Description:
//  该管道首先在PIT条目中为指定的传出 LogicFace 插入一个 out-record ，或者为同一 LogicFace 更新一个现有的 out-record 。 在这两种情况下，
//  PIT记录都将记住最后一个传出兴趣数据包的 Nonce ，这对于匹配传入的Nacks很有用，还有到期时间戳，它是当前时间加上 InterestLifetime 。最后，
//  Interest 被发送到传出的 LogicFace 。
// @param egress
// @param pitEntry
// @param interest
//
func (f *Forwarder) OnOutgoingInterest(egress *lf.LogicFace, pitEntry *table.PITEntry, interest *packet.Interest) {
	common.LogDebugWithFields(logrus.Fields{
		"faceId":   egress.LogicFaceId,
		"interest": interest.ToUri(),
	}, "Outgoing interest")

	// 调用插件锚点
	if f.pluginManager.OnOutgoingInterest(egress, pitEntry, interest) != 0 {
		return
	}

	// 插入 out-record
	outRecord := pitEntry.InsertOrUpdateOutRecord(egress, interest)
	outRecord.ExpireTime = GetCurrentTime() + interest.InterestLifeTime.GetInterestLifeTime()

	// 转发兴趣包
	egress.SendInterest(interest)
}

//
// 兴趣包最终回收处理，此时兴趣包要么被满足要么被Nack （ Interest Finalize Pipeline ）
//
// @Description:
// @param pitEntry
//
func (f *Forwarder) OnInterestFinalize(pitEntry *table.PITEntry) {
	common.LogDebugWithFields(logrus.Fields{
		"entry": pitEntry.GetIdentifier().ToUri(),
	}, "Interest finalize")

	// 调用插件锚点
	if f.pluginManager.OnInterestFinalize(pitEntry) != 0 {
		return
	}

	// 如果传入的 PITEntry 已经被移除了，就直接返回
	if pitEntry.IsDeleted() {
		return
	}

	// 将对应的PIT条目从PIT表中移除
	if err := f.PIT.EraseByPITEntry(pitEntry); err != nil {
		// 删除 PIT 条目失败，在这边输出提示信息
		common.LogWarnWithFields(logrus.Fields{
			"interest": pitEntry.GetIdentifier().ToUri(),
		}, "Delete PITEntry failed")
	}

	// 标记 PIT 条目已经被删除
	pitEntry.SetDeleted(true)
}

//
// 处理一个数据包到来（ Incoming Data Pipeline ）
//
// @Description:
//  1. 首先，管道使用数据匹配算法（ Data Match algorithm ，第3.4.2节）检查 Data 是否与PIT条目匹配。如果找不到匹配的PIT条目，则将 Data
//     提供给 Data unsolicited 管道；如果找到匹配的PIT条目，则将 Data 插入到 ContentStore 中。
//
//     > 请注意，即使管道将 Data 插入到 ContentStore 中，该数据是否存储以及它在 ContentStore 中的停留时间也取决于 ContentStore 的接纳
//       和替换策略（ admission andreplacement policy）。
//
//  2. 接着管道会将对应PIT条目的到期计时器设置为当前时间，调用对应策略的 StrategyBase::afterReceiveData 回调，将PIT标记为 satisfied ，并清
//     除PIT条目的 out records 。
// @param ingress
// @param data
//
func (f *Forwarder) OnIncomingData(ingress *lf.LogicFace, data *packet.Data) {
	common.LogDebugWithFields(logrus.Fields{
		"faceId": ingress.LogicFaceId,
		"data":   data.ToUri(),
	}, "Incoming data")

	// 调用插件锚点
	if f.pluginManager.OnIncomingData(ingress, data) != 0 {
		return
	}

	// 找到对应的PIT条目
	pitEntry := f.PIT.FindDataMatches(data)
	if pitEntry == nil {
		// 没有找到对应的 PIT 条目，触发 Data unsolicited 管道
		f.OnDataUnsolicited(ingress, data)
		return
	}
	// 找到对应的 PIT 条目
	// 插入到CS缓存当中
	f.CS.Insert(data)

	// 调用对应策略的 StrategyBase::afterReceiveData 回调
	if ste := f.StrategyTable.FindEffectiveStrategyEntry(data.GetName()); ste != nil {
		// 调用策略
		ste.GetStrategy().AfterReceiveData(ingress, data, pitEntry)
		// 标记 PITEntry 为 satisfied
		pitEntry.SetSatisfied(true)
		// 清除对应的出记录
		if err := pitEntry.DeleteOutRecord(ingress); err != nil {
			// 删除出记录失败，这边输出错误
			common.LogWarnWithFields(logrus.Fields{
				"pitEntry": pitEntry.GetIdentifier().ToUri(),
			}, "Delete out-record failed: ", err)
		}
	} else {
		// 输出错误，数据包没有找到匹配的可用策略
		common.LogErrorWithFields(logrus.Fields{
			"data": data.ToUri(),
		}, "Not found effective StrategyBase")
	}
}

//
// 收到一个数据包，但是这个数据包是未被请求的 （ Data Unsolicited Pipeline ）
//
// @Description:
//  在 Incoming data 管道处理过程中发现 Data 是未经请求的时后会触发 Data unsolicited 管道处理逻辑，它的处理过程如下：
//   1. 根据当前配置的针对未经请求的 Data 的处理策略，决定是删除 Data 还是将其添加到 ContentStore 。默认情况下，MIR配置了 drop-all 策略，
//      该策略会丢弃所有未经请求的 Data ，因为它们会对转发器造成安全风险。
//   2. 在某些特殊应用场景下，如果希望MIR将未经请求的 Data 存储到 ContentStore，可以在配置文件中修改对应的策略。
// @param ingress
// @param data
//
func (f *Forwarder) OnDataUnsolicited(ingress *lf.LogicFace, data *packet.Data) {
	common.LogDebugWithFields(logrus.Fields{
		"faceId": ingress.LogicFaceId,
		"data":   data.ToUri(),
	}, "Data unsolicited")

	// 调用插件锚点
	if f.pluginManager.OnDataUnsolicited(ingress, data) != 0 {
		return
	}
	// TODO: 读取配置文件，是否缓存未经请求的 Data
}

//
// 处理将一个数据包发出 （ Outgoing Data Pipeline ）
//
// @Description:
//  在 Incoming Interest 管道（第4.2.1节）处理过程中在 ContentStore 中找到匹配的数据或在 Incoming Data 管道处理过程中发现传入的 Data
//  匹配到 PIT 表项时，调用本管道，它的处理过程如下：
//   1. 通过对应的 LogicFace 将 Data 发出
// @param egress
// @param data
//
func (f *Forwarder) OnOutgoingData(egress *lf.LogicFace, data *packet.Data) {
	common.LogDebugWithFields(logrus.Fields{
		"faceId": egress.LogicFaceId,
		"data":   data.ToUri(),
	}, "Outgoing data")

	// 调用插件锚点
	if f.pluginManager.OnOutgoingData(egress, data) != 0 {
		return
	}

	egress.SendData(data)
}

//
// 处理一个 Nack 到来 （ Incoming Nack Pipeline ）
//
// @Description:
//  1. 首先，从收到的 Nack 中提取到 Interest，然后查询是否有与之匹配的PIT条目，如果没有则丢弃，有则执行下一步；
//  2. 接着，判断匹配到的 PIT 条目中是否有对应 LogicFace 的 out-record ，如果没有则丢弃，有则执行下一步；
//  3. 然后，判断得到的 out-record 是否和 Nack 中的 Interest 的 Nonce 一致，不一致则丢弃，一致则执行下一步；
//  4. 然后标记对应的 out-record 为 Nacked ；
//  5. 如果此时对应的 PIT 条目中所有的 out-record 都已经 Nacked ，则将PIT条目的过期时间设置为当前时间（会触发 Interest finalize 管道）；
//  6. 然后调用对应策略的 StrategyBase::afterReceiveNack 回调，在其中触发 Outgoing Nack 管道。
// @param ingress
// @param nack
//
func (f *Forwarder) OnIncomingNack(ingress *lf.LogicFace, nack *packet.Nack) {
	common.LogDebugWithFields(logrus.Fields{
		"faceId":   ingress.LogicFaceId,
		"interest": nack.Interest.ToUri(),
		"reason":   nack.GetNackReason(),
	}, "Incoming Nack")

	// 调用插件锚点
	if f.pluginManager.OnIncomingNack(ingress, nack) != 0 {
		return
	}

	// 判断 PIT 中是否有对应的条目
	pitEntry, err := f.PIT.Find(nack.Interest)
	if err != nil || pitEntry == nil {
		// 没有找到匹配的 PIT 条目，直接返回丢弃
		common.LogDebugWithFields(logrus.Fields{
			"nack":   nack.Interest.ToUri(),
			"reason": nack.GetNackReason(),
		}, "Have not found match PITEntry for nack")
		return
	}

	outRecord, err := pitEntry.GetOutRecord(ingress)
	if err != nil || outRecord == nil {
		// 如果不存在对应 LogicFace 的 out-record，则丢弃
		common.LogDebugWithFields(logrus.Fields{
			"nack":   nack.Interest.ToUri(),
			"reason": nack.GetNackReason(),
		}, "Have not found match out-record for nack")
		return
	}

	// 记录 NackHeader 到 out-record
	if outRecord.LastNonce.GetNonce() != nack.Interest.GetNonce() {
		// 如果 Nonce 不一致，直接丢弃
		common.LogDebugWithFields(logrus.Fields{
			"nack":   nack.Interest.ToUri(),
			"reason": nack.GetNackReason(),
		}, "Founded matched out-record, but Nonce is diff")
		return
	}
	outRecord.NackHeader = &nack.Interest.NackHeader

	// 如果所有 out-record 都超时或者被 Nack，则触发 PIT 条目过期
	finished := true
	for _, or := range pitEntry.GetOutRecords() {
		if or.ExpireTime > GetCurrentTime() && or.NackHeader != nil {
			finished = false
		}
	}
	if finished {
		f.SetExpiryTime(pitEntry, 0)
	}

	// 触发 StrategyBase::afterReceiveNack
	if ste := f.StrategyTable.FindEffectiveStrategyEntry(nack.Interest.GetName()); ste != nil {
		ste.GetStrategy().AfterReceiveNack(ingress, nack, pitEntry)
	} else {
		// 输出错误，Nack没有找到匹配的可用策略
		common.LogErrorWithFields(logrus.Fields{
			"nack":   nack.Interest.ToUri(),
			"reason": nack.GetNackReason(),
		}, "Not found matched StrategyBase for nack")
	}
}

//
// 处理一个 Nack 发出 （ Outgoing Nack Pipeline ）
//
// @Description:
//  1. 首先，在PIT条目中查询指定的传出 LogicFace （下游）的 in-record 。该记录是必要的，因为协议要求将最后一个从下游接收到的 Interest
//    （包括其Nonce）携带在 Nack 包中，如果未找到记录，请中止此过程，因为如果没有此兴趣，将无法发送 Nack 。
//  2. 然后构造一个 Nack 传递给下游，同时删除对应的 in-record。
// @param egress
// @param pitEntry
// @param header
//
func (f *Forwarder) OnOutgoingNack(egress *lf.LogicFace, pitEntry *table.PITEntry, header *component.NackHeader) {
	common.LogDebugWithFields(logrus.Fields{
		"faceId":   egress.LogicFaceId,
		"pitEntry": pitEntry.GetIdentifier().ToUri(),
		"reason":   header.GetNackReason(),
	}, "Outgoing Nack")

	// 调用插件锚点
	if f.pluginManager.OnOutgoingNack(egress, pitEntry, header) != 0 {
		return
	}

	// 查找对应的 in-record
	inRecord, err := pitEntry.GetInRecord(egress)
	if err != nil || inRecord == nil {
		// 如果不存在对应的 in-record，丢弃包
		return
	}

	// 构造 Nack 发出
	nack := packet.Nack{}
	nack.Interest = inRecord.Interest
	nack.SetNackReason(header.GetNackReason())
	egress.SendNack(&nack)
}

//
// 处理一个 CPacket 到来 （Incoming CPacket Pipeline）
//
// @Description:
//  1. 首先给 CPacket 的 TTL 减一，然后检查 TTL 的值是：
//     - TTL < 0 则认为该包是一个回环的 CPacket ，直接丢弃；
//     - TTL >= 0 则执行下一步。
//   > 因为 CPacket 是一种推式语义的网络包，不能向 Interest 那样通过 PIT 聚合来检测回环，所以这边和 IP 一样使用 TTL 来避免网络包无限回环。
//
//  2. 接着调用对应策略的 StrategyBase::afterReceiveCPacket 回调，在其中触发 Outgoing CPacket 管道。
// @param ingress
// @param cPacket
//
func (f *Forwarder) OnIncomingCPacket(ingress *lf.LogicFace, cPacket *packet.CPacket) {
	common.LogDebugWithFields(logrus.Fields{
		"faceId":  ingress.LogicFaceId,
		"cPacket": cPacket.ToUri(),
	}, "Incoming CPacket")

	// 调用插件锚点
	if f.pluginManager.OnIncomingCPacket(ingress, cPacket) != 0 {
		return
	}

	// TTL 减一，并且检查 TTL 是否小于0，小于0则判定为循环包
	if cPacket.TTL.Ttl() == 0 {
		common.LogDebugWithFields(logrus.Fields{
			"faceId":  ingress.LogicFaceId,
			"cPacket": cPacket.ToUri(),
		}, "CPacket TTL < 0")
		return
	}
	cPacket.TTL.Minus()

	// 调用 StrategyBase::afterReceiveCPacket
	if ste := f.StrategyTable.FindEffectiveStrategyEntry(cPacket.DstIdentifier()); ste != nil {
		ste.GetStrategy().AfterReceiveCPacket(ingress, cPacket)
	} else {
		// 输出错误，CPacket没有找到匹配的可用策略
		common.LogErrorWithFields(logrus.Fields{
			"faceId":  ingress.LogicFaceId,
			"cPacket": cPacket.ToUri(),
		}, "Not found matched StrategyBase for CPacket")
	}
}

//
// 处理一个 CPacket 发出 （Outgoing CPacket Pipeline）
//
// @Description:
// @param egress
// @param cPacket
//
func (f *Forwarder) OnOutgoingCPacket(egress *lf.LogicFace, cPacket *packet.CPacket) {
	common.LogDebugWithFields(logrus.Fields{
		"faceId":  egress.LogicFaceId,
		"cPacket": cPacket.ToUri(),
	}, "Outgoing CPacket")

	// 调用插件锚点
	if f.pluginManager.OnOutgoingCPacket(egress, cPacket) != 0 {
		return
	}

	egress.SendCPacket(cPacket)
}

//
// 设置 PIT 条目的超时时间，并在超时时触发 OnInterestFinalize 管道
//
// @Description:
// @receiver f
// @param pitEntry
// @param duration			单位 ms
//
func (f *Forwarder) SetExpiryTime(pitEntry *table.PITEntry, duration time.Duration) {
	// TODO: 这边要check一下，是不是调用 SetExpiryTime 的时候之前的定时任务还没有触发，如果已经触发过了，是不是会有问题

	// 首先取消之前的定时任务
	pitEntry.CancelTimer()
	// 接着设置新的定时任务
	pitEntry.SetExpiryTimer(duration*time.Millisecond, func(entry *table.PITEntry) {
		f.OnInterestFinalize(entry)
	})
}
