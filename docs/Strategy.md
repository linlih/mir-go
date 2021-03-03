# MIR Strategy

本节对MIR的策略模块作详细的说明和设计。

在MIR的处理包转发的过程中，转发策略（ *forwarding strategy* ）对转发的行为作智能的决策，它决定了是否、何时以及将网络包转发到何处。具体地，每个转发策略都由以下部分组成：

- **触发器（Triggers）** ：触发器由一系列的触发函数构成，每个触发函数均为执行策略程序的入口；
  - `AfterReceiveInterest`
  - `AfterContentStoreHit`
  - `AfterReceiveData`
  - `AfterReceiveNack`
  - `AfterReceiveCPacket`
- **操作（Actions） ** ：每个操作（ *Action* ）实际上就是策略程序实际作出的转发决策。
  - `sendInterest`
  - `sendData`
  - `sendNack`
  - `sendCPacket`

MIR中可以定义很多的策略，但是对于某个具体的网络包的转发必须由单一的转发策略决定，为此我们根据命名空间来进行策略的选择。网络管理员可以为某个前缀配置特定的策略，默认至少会为 `/` 前缀配置一个策略，保证所有的包至少是可以匹配到策略的。实际使用时，转发管道会去策略选择表进行最长前缀匹配，找到匹配的策略来进行转发决策。

## 1. Triggers

触发器（ *Triggers* ）是策略程序的入口，由转发管道调用并触发。

### 1.1 After Receive Interest Trigger

```go
//
// 当收到一个兴趣包时，会触发本触发器
//
// @Description:
//	Interest 需要满足以下条件：
//		- Interest 不是回环的
//		- Interest 没有命中缓存
//		- Interest 位于当前策略的命名空间下
//
// @param ingress		Interest到来的入口LogicFace
// @param interest		收到的兴趣包
// @param pitEntry		兴趣包对应的PIT条目
//
AfterReceiveInterest(ingress *lf.LogicFace, interest *packet.Interest, pitEntry *table.PITEntry)
```

当MIR收到一个 `Interest` 时，会传递给 **Incoming Interest** 管道处理，如果这个 `Interest` 满足下述的几个条件，那么 **Incoming Interest** 管道将会触发 **After Receive Interest** 触发器：

- 存在 `TTL` ，并且 `TTL` 大于等于1；

- `Interest` 不是回环的；
- `Interest` 没有命中缓存；
- `Interest` 位于当前策略的命名空间下。

当本触发器被触发后，策略程序需要决定将 `Interest` 转发往何处（即从哪个或哪些 *LogicFace* 将 `Interest` 转发出去）。大多数策略此时的行为都是通过查询FIB表决定如何转发 `Interest` ，这个可以通过调用 `Strategy.lookupFib` 来实现。

- 如果策略决定转发该 `Interest` ，则应该至少调用一次 `Strategy.sendInterest` 操作将其转发出去；
- 如果策略决定不转发该 `Interest` ，则应该调用 `Strategy.setExpiryTimer` 操作并将对应PIT条目的超时时间设置为当前时间，使得对应的PIT条目记录可以正确的清除。

### 1.2 After ContentStore Hit

```go
//
// 当兴趣包命中缓存时，会触发本触发器
//
// @Description:
//
// @param ingress		Interest到来的入口LogicFace
// @param data			缓存中得到的可以满足 Interest 的 Data
// @param entry			兴趣包对应的PIT条目
//
AfterContentStoreHit(ingress *lf.LogicFace, data *packet.Data, entry *table.PITEntry)
```

当MIR收到一个 `Interest` 时，会传递给 **Incoming Interest** 管道处理，如果在管道处理过程中在 *ContentStore* 中查询到匹配的内容，且内有有效，则会触发本触发器。

> *ContentStore* 中查询到匹配的内容有效：
>
> - 一种情况是，所缓存的 Data 仍然是新鲜的；
> - 另一种情况是，所缓存的 Data 虽然不是新鲜的，但是 Interest 可以接受不新鲜的数据

此触发器默认使用 `Strategy.sendData` 操作将匹配的 `Data` 发送到兴趣包到来方向的下游路由器。

### 1.3 After Receive Data

```go
//
// 当收到一个 Data 时，会触发本触发器
//
// @Description:
//	Data 应当满足下列条件：
//		- Data 被验证过可以匹配对应的PIT条目
//		- Data 位于当前策略的命名空间下
// @param ingress		Data 到来的入口 LogicFace
// @param data			收到的 Data
// @param pitEntry		Data 对应匹配的PIT条目
//
AfterReceiveData(ingress *lf.LogicFace, data *packet.Data, pitEntry *table.PITEntry)
```

当传入的 `Data` 匹配到 PIT 条目时，会触发本触发器，此时策略将对 `Data` 的转发具有完全的控制权。默认情况下，本触发器将把 `Data` 转发到所有有效的下游 *LogicFace*。调用此触发器是需要满足以下先决条件：

- 存在 `TTL` ，并且 `TTL` 大于等于1

- `Data` 经过查表和验证，可以匹配到 PIT 条目；
- `Data` 位于当前策略的命名空间下。

此触发器内部应当完成以下功能：

- 策略应当通过 `Strategy.sendData` 或者 `Strategy.sendDataToAll` 将 `Data` 发送给下游的节点；
- 策略可以对 `Data` 进行适当的更改，只要修改之后 `Data` 仍然能够匹配对应的 PIT 条目即可，例如添加或者删除拥塞标记；
- 策略应当至少调用一次 调用`Strategy.setExpiryTimer`：
  - 默认情况下， `Strategy.setExpiryTimer` 将PIT条目的超时时间设置为当前时间，以启动 PIT 条目的清理流程；
  - 策略也可以选择调用 `Strategy.setExpiryTimer` 延长 PIT 条目的存活期，从而延迟 `Data` 的转发，只要保证在 PIT 条目被移除之前转发 `Data` 即可。
- 策略可以在此触发器内收集有关上游的度量信息（比如可以计算RTT）；
- 策略可以通过延长收到 `Data` 的PIT条目的生存期，从而等待其它上游节点返回 `Data` （可以从多个上游节点收集 `Data` ，并决策将哪个 `Data` 转发到下游），需要注意的是，**对于每一个下有节点，要保证只有一个 Data 转发到下游路由器**。

### 1.4 After Receive Nack

```go
//
// 当收到一个 Nack 时，会触发本触发器
//
// @Description:
//
// @param ingress		Nack 到来的入口 LogicFace
// @param nack			收到的 Nack
// @param pitEntry		Nack 对应匹配的PIT条目
//
AfterReceiveNack(ingress *lf.LogicFace, nack *packet.Nack, pitEntry *table.PITEntry)
```

当MIR收到一个 `Nack` ，会传递给 **Incoming Nack** 管道处理，如果 `Nack` 满足下述的几个条件，那么 **Incoming Nack** 管道将会触发 **After Receive Nack** 触发器：

- 存在 `TTL` ，并且 `TTL` 大于等于1；

- `Nack` 响应一个已经转发的 `Interest` ，即使用 `Nack` 中包含的 `Interest` 可以在 PIT 表中检索到匹配的 PIT 条目；

  > 如果收到一个 `Nack` 却没有检索到匹配的 PIT 条目，可能是原有的 PIT 条目已经过期或者被来自其它上游的 `Data` 满足，此时应当直接丢弃它。

- `Nack` 是对转发给该上游的最后一个 `Interest` 的响应，即在对应的 PIT 条目中存在一个 *out-record* ，并且 `Nack` 中包含的 `Nonce` 和该 *out-record* 中的相同；

- `Nack` 位于当前策略的命名空间下；

  > 注意：`Nack` 对应的 `Interest` 不一定是由同一个策略转发的。如果在转发 `Interest` 后更改了有效策略，然后收到了对应的 `Nack` ，则会触发新的有效策略，而不是先前转发 `Interest` 的策略。

- `NackHeader` 已经被记录在对应 *out-record* 的 *Nacked* 字段。

当 **After Receive Nack** 触发器被触发后，策略程序通常可以执行下述的某一种操作：

- 通过调用 *send Interest* 操作将其转发到相同或不同的上游来重试兴趣（ *Retry the Interest* ）。大多数策略都需要一个FIB条目来找出潜在的上游，这可以通过调用 `Strategy.lookupFib` 访问器函数获得；
- 通过调用 *send Nack* 操作将 `Nack` 反回到下游，放弃对该 `Interest` 的重传尝试；
- 不对这个 `Nack` 做任何处理。如果 `Nack` 对应的 `Interest` 转发给了多个上游，且某些（但不是全部）上游回复了 `Nack` ，则该策略可能要等待来自更多上游的 `Data` 或 `Nack` 。

### 1.5 After Receive CPacket

```go
//
// 当收到一个 CPacket 时，会触发本触发器
//
// @Description:
// @param ingress		CPacket 到来的入口 LogicFace
// @param cPacket		收到的 CPacket
//
AfterReceiveCPacket(ingress *lf.LogicFace, cPacket *packet.CPacket)
```

当MIR收到一个 `CPacket`，会传递给 **Incoming CPacket** 管道处理，如果 `CPacket` 满足下述的几个条件，那么 **Incoming CPacket** 管道将会触发 **After Receive CPacket** 触发器：

- 存在 `TTL` ，并且 `TTL` 大于等于1

当 **After Receive CPacket** 触发器被触发后，策略程序通常的行为为查询FIB表，找到可用的路由将 `CPacket` 转发出去

## 2. Actions

所谓操作（ *Action* ） 是转发策略 （ *forwarding strategy* ）对网络包的转发作出的决策，由上一节提到的触发器调用。

### 2.1 Send Interest

```go
// 将兴趣包从指定的逻辑接口转发出去
//
// @Description:
// @param egress		转发 Interest 的出口 LogicFace
// @param interest		要转发的 Interest
// @param entry			Interest 对应匹配的 PIT 条目
//
sendInterest(egress *lf.LogicFace, interest *packet.Interest, entry *table.PITEntry)
```

转发策略（ *forwarding strategy* ）可以调用 **sendInterest** 操作转发一个 `Interest` ，调用之前应该保证传入的 `Interest` 是对应的PIT条目中某个 *in-record* 中所存储的。只要不影响 `Interest` 和对应 PIT 条目的匹配关系，可以创建 `Interest` 的副本并对其进行修改。

最后本操作将会启动 **Outgoing Interest** 管道处理流程。

### 2.2 Send Data

```go
//
// 将 Data 从指定的逻辑接口转发出去
//
// @Description:
// @param egress		转发 Data 的出口 LogicFace
// @param data			要转发的 Data
// @param pitEntry		Data 对应匹配的 PIT 条目
//
sendData(egress *lf.LogicFace, data *packet.Data, pitEntry *table.PITEntry)
```

转发策略（ *forwarding strategy* ）可以调用 **sendData** 操作转发一个 `Data`，该操作会执行以下步骤：

- 首先，删除对应 PIT 条目中对应的 *in-record*；
- 接着启动 **Outgoing Data** 管道处理流程。

在多数情况下，转发策略通常希望将 `Data` 发送到每个有效下游，所以我们还定义一个辅助函数，用来将 `Data` 发往所有符合条件的下游：

```go
//
// 将 Data 发送给对应 PIT 条目记录的所有符合条件的下游节点
//
// @Description:
// @param ingress		Data 到来的入口 LogicFace => 主要是用来避免往收到 Data 包的 LogicFace 转发 Data
// @param data			要转发的 Data
// @param pitEntry		Data 对应匹配的 PIT 条目
//
sendDataToAll(ingress *lf.LogicFace, data *packet.Data, pitEntry *table.PITEntry)
```

**sendDataToAll** 函数应该包含以下执行步骤：

- 首先，从 PIT 条目的 *in-records* 中过滤出有效的可用于转发的连接下游的 *LogicFace*，满足下列条件的 *LogicFace* 称为有效的 *LogicFace*  ：
  - 首先，该 *LogicFace* 必然是包含在要转发的 `Data` 对应匹配的 PIT 条目的 *in-records* 当中的；
  - 然后，对应的 *in-record* 应该是没有过期的（ *in-record* 的超时时间大于当前时间）；
  - 最后，该 *LogicFace* 不能是收到 `Data` 的 *LogicFace*。
- 然后往每一个有效的 *LogicFace* 通过调用 **sendData** 操作，将 `Data` 发往该下游。

### 2.3 Send Nack

```go
//
// 往指定的逻辑接口发送一个 Nack
//
// @Description:
// @param egress		转发 Nack 的出口 LogicFace
// @param nackHeader	要转发出的Nack的元信息
// @param pitEntry		Nack 对应匹配的 PIT 条目
//
sendNack(egress *lf.LogicFace, nackHeader *component.NackHeader, pitEntry *table.PITEntry)
```

转发策略（ *forwarding strategy* ）可以调用 **sendNack** 操作尝试往一个指定的下游转发一个 `Nack`，本操作会启动 **Outgoing Nack** 管道处理流程。

> **Outgoing Nack** 管道会执行以下步骤：
>
> - 首先尝试根据 *egress* 去 *pitEntry* 中找到匹配的 *in-record* ，如果没有，则操作无效直接返回；
> - 接着从 *in-record* 中提取出对应的 `Interest` ，和 *nackHeader* 组合成一个 `Nack` ，然后通过 *egress* 转发出去。

在多数情况下，转发策略通常希望将 `Nack` 发送到每个有效的下游，所以我们定义一个辅助函数，用来将 `Nack` 发往所有符合条件的下游：

```go
//
// 将 Nack 发送给对应 PIT 条目记录的所有符合条件的下游节点
//
// @Description:
// @param ingress		收到 Nack 的入口 LogicFace
// @param nackHeader	要转发出的Nack的元信息
// @param pitEntry		Nack 对应匹配的 PIT 条目
//
sendNackToAll(ingress *lf.LogicFace, nackHeader *component.NackHeader, pitEntry *table.PITEntry)
```

**sendNackToAll** 函数应该包含以下执行步骤：

- 首先，从 PIT 条目的 *in-records* 中过滤出有效的可用于转发的连接下游的 *LogicFace*，满足下列条件的 *LogicFace* 称为有效的 *LogicFace*  ：
  - 首先，该 *LogicFace* 必然是包含在要转发的 `Data` 对应匹配的 PIT 条目的 *in-records* 当中的；
  - 然后，该 *LogicFace* 不能是收到 `Data` 的 *LogicFace*。
- 然后往每一个有效的 *LogicFace* 通过调用 **sendNack** 操作，将 `Nack` 发往该下游。

### 2.4 Send CPacket

```go
//
// 往指定的逻辑接口发送一个 CPacket
//
// @Description:
// @param egress		转发 CPacket 的出口 LogicFace
// @param cPacket		要转发出的 CPacket
//
sendCPacket(egress *lf.LogicFace, cPacket *packet.CPacket)
```

转发策略（ *forwarding strategy* ）可以调用 **sendCPacket** 操作转发一个 `CPacket` ，本操作将会启动 **Outgoing CPacket** 管道处理流程。

## 3. 其它辅助函数

### 3.1 lookupFibForInterest

```go
//
// 在 FIB 表中查询可用于转发 Interest 的 FIB 条目
//
// @Description:
// @param interest
//
lookupFibForInterest(interest *packet.Interest)
```

### 3.2 lookupFibForCPacket

```go
//
// 在 FIB 表中查询可用于转发 CPacket 的 FIB 条目
//
// @Description:
// @param cPacket
//
lookupFibForCPacket(cPacket *packet.CPacket)
```

