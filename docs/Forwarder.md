# MIR Forwarder

本节讲述MIR路由器的转发流程。MIR具有智能的转发平面，该转发平面由 **转发管道** （ *Forwarding Pipelines* ） 和 **转发策略** （ *Forwarding Strategy* ）组成。

- **转发管道** （ *Forwarding Pipelines* ）：每个转发管道都是由一系列对网络包的处理逻辑组成的，如下图中每个蓝色框都表示一个转发管道。
- **转发策略** （ *Forwarding Strategy* ）：转发策略主要用于在网络包的处理流程中，对网络包的转发行为进行决策，如下图中每个白色框代表的就是一个决策点。

当某个事件被触发并且满足对应的条件时，会启动一个转发管道处理流程，或者从一个转发管道传递到另一个转发管道。

![Forwarder转发管道图](https://gitee.com/quejianming/pic-bed/raw/master/uPic/Forwarder%E8%BD%AC%E5%8F%91%E7%AE%A1%E9%81%93%E5%9B%BE.svg)

## 1. 转发管道

管道（ *pipelines* ）处理网络层数据包（`Interest`、`Data`、`GPPkt` 或 `Nack`），并且每个数据包都从一个管道传递到另一个管道（在满足某些条件的情况下，也会传递给策略决策点），直到所有处理流程完成为止。管道内的处理使用CS、PIT、FIB和策略选择表，但是管道对后两个表仅具有只读访问权限，因为这些表由相应的管理器管理，并且不直接受数据面流量的影响。

`LogicFaceTable`跟踪MIR中所有的活动的逻辑接口（ *LogicFace* ） 。它是网络层数据包进入转发管道进行处理的入口点，管道还可以通过 *LogicFace* 发送数据包。

MIR中对`Interest`、`Data`、`GPPkt` 和 `Nack`数据包的处理是完全不同的。我们将转发管道分为 **内容兴趣包处理路径** （ *Interest processing path* ）、 **内容数据包处理路径** （ *Data processing path* ）、**Nack处理路径** （ *Nack processing path* ）和 **通用推式包处理路径** （ *GPPkt processing path* ），这将在以下各节中进行介绍。

## 2. 兴趣包处理路径

MIR中Interest包的处理流程包含以下管道：

- **Incoming Interest** ：路由器收到一个新的兴趣包时，会触发本管道
- **Interest loop** ：路由器检测到一个兴趣包为循环的兴趣包时，会触发本管道
- **ContentStore miss** ：路由器收到一个兴趣包，并且没有命中缓存时，会触发本管道
- **ContentStore hit** ：路由器收到一个兴趣包，并且命中缓存时，会触发本管道
- **Outgoing Interest** ：当路由器做完决策，并需要将兴趣包转发出去时，会触发本管道
- **Interest finalize** ：当兴趣包被满足或者超时，会触发本管道清理PIT条目

### 2.1 Incoming Interest Pipeline

![incoming-interest-pipeline](https://gitee.com/quejianming/pic-bed/raw/master/uPic/2021/02/20/incoming-interest-pipeline-1613803561.svg)

如上图所示的是 **Incoming Interest** 管道的处理流程图，包括以下步骤：

1. 首先给 `Interest` 的 `TTL` 减一，然后检查 `TTL` 的值是：

   - `TTL` < 0 则认为该兴趣包是一个回环的 `Interest` ，直接将其传递给 **Interest loop** 管道进一步处理；
   - `TTL` >= 0 则执行下一步。

   > 问题：下面第3步通过PIT条目进行回环的 `Interest` 检测，那为什么这边还需要一个 `TTL` 的机制来检测回环呢？
   >
   > - 因为通过PIT条目对比 `Nonce` 的方式来检测循环存在一个问题，当一个兴趣包被转发出去，假设其对应的PIT条目的过期时间为 $x$ 秒，那如果这个兴趣包在经过 $y$ 秒之后回环到当前路由器（$y > x$），则此时该兴趣包对应的PIT条目已经被移除，路由器无法通过PIT聚合的方式来检测回环的兴趣包；
   > - 所以，为了解决上述问题，我们通过类比IP中简单的设置 `TTL` 的方式来检测上面描述的特殊情况下，不能检测回环 `Interest` 的问题。
   > - NDN通过引入 *Dead Nonce List* 来进行检测，我们觉得会引入复杂性，且是概率性可靠的，直接用 `TTL` 的方式可以实现简单，且减少查询 *Dead Nonce List* 的开销。

2. 根据传入的 `Interest` 创建一个PIT条目（如果存在同名的PIT条目，则直接使用，不存在则创建）。

3. 然后查询PIT条目中是否有和传入的 `Interest` 的 `Nonce` 相同，并且是从不同的 *LogicFace* 收到的入记录（ *in-records* ），如果找到匹配的入记录，则认为传入的 `Interest` 是回环的循环 `Interest` ，直接将其传递给 **Interest loop** 管道进一步处理；否则执行下一步。

   > - 如果从同一个逻辑接口收到同名且 `Nonce` 相同的 `Interest`，则可能是同一个消费者发送的 `Interest`，该 `Interst` 被判定为合法的重传包；
   > - 如果从不同的逻辑接口收到同名且 `Nonce` 相同的 `Interest`，则可能是循环的 `Interest` 或者是同一个 `Interest` 沿着多个不同的路径到达，此时，将传入的 `Interest` 判定为循环的 `Interest` ，触发  **Interest loop** 管道。

4. 然后通过查询 PIT 条目中的记录，判断当前 `Interst` 是否是未决的（ *pending* ），如果**传入的 `Interest` 对应的PIT条目包含其它记录**，则认为该 `Interest` 是未决的。

5. 如果 `Interest` 是未决的，则直接传递给 **ContentStore miss** 管道处理；如果 `Interest` 不是未决的，则查询CS，如果存在缓存，则传递给 **ContentStore hit** 管道进行进一步处理，否则传递给 **Content miss** 管道进行进一步的处理。

### 2.3 Interest Loop Pipeline

在 **Incoming Interest** 管道处理过程中，如果检测到 `Interest` 循环就会触发 **Interest loop** 管道，本管道会向收到 `Interest` 的 `LogicFace` 发送一个原因为 "重复" （ *duplicate* ） 的 `Nack`。

### 2.4 ContentStore Hit Pipeline

在 **incoming Interest** 管道中执行 `ContentStore` 查找并找到匹配项之后触发 **ContentStore hit** 管道处理逻辑。此管道执行以下步骤：

![ContentStore-hit-pipeline](https://gitee.com/quejianming/pic-bed/raw/master/uPic/ContentStore-hit-pipeline.svg)

1. 首先将 `Interest` 对应PIT条目的到期计时器设置为当前时间，这会使得计时器到期，触发 **Interest finalize** 管道；
2. 然后触发 `Interest` 对应策略的 `Strategy::afterContentStoreHit` 回调。

### 2.5 ContentStore Miss Pipeline

在 **incoming Interest** 管道中执行 `ContentStore` 查找，但是没有找到匹配项会触发 **ContentStore miss** 管道处理逻辑。此管道执行以下步骤：

![ContentStore-miss-pipeline](https://gitee.com/quejianming/pic-bed/raw/master/uPic/ContentStore-miss-pipeline.svg)

1. 首先根据传入的 `Interest` 以及对应的传入 *LogicFace* 在尝试在对应的PIT条目中插入一条 *in-record*；
   - 如果对应的PIT条目中已经存在一个相同 *LogicFace* 的 *in-record* 记录（比如：下游正在重传同一个兴趣包），那只需要用收到的 `Interest` 中的 `Nonce` 和 `InterestLifetime` 来更新对应的 *in-record* 即可，如果没有指定 `InterestLifetime`，则默认为4s；
   - 否则创建一个新的 *in-record* 记录插入到对应的PIT条目当中。
2. 然后将PIT条目的超时计时器设置为当前 PIT 条目中所有 *in-record* 最大剩余超时时间。
3. 然后传递给转发策略作转发决策，在转发策略中按需触发 **Outgoing Interest** 管道处理逻辑，将 `Interest` 转发出去。

### 2.6 Outgoing Interest Pipeline

该管道首先在PIT条目中为指定的传出 *LogicFace* 插入一个 *out-record* ，或者为同一 *LogicFace* 更新一个现有的 *out-record* 。 在这两种情况下，PIT记录都将记住最后一个传出兴趣数据包的 *Nonce* ，这对于匹配传入的Nacks很有用，还有到期时间戳，它是当前时间加上 *InterestLifetime* 。最后， `Interest` 被发送到传出的 *LogicFace* 。

### 2.7 Interest Finalize Pipeline

**Interest finalize** 管道通常是由超时计时器到期时触发的，包含以下步骤：

4. 最后将对应的PIT条目从PIT表中移除。

## 3. 数据包处理路径

MIR中 `Data` 的处理流程包含以下管道：

- **Incoming Data**：处理传入 `Data`
- **Data unsolicited**：处理传入的未经请求的 `Data`
- **Outgoing Data**：准备并发送出 `Data`

### 3.1 Incoming Data Pipeline

![IncomingData-pipeline](https://gitee.com/quejianming/pic-bed/raw/master/uPic/2021/02/20/IncomingData-pipeline-3806033-1613825224.svg)

如上图所示的是 **Incoming Data** 管道的处理流程图，包括以下步骤：

1. 首先，管道使用数据匹配算法（ *Data Match algorithm* ，第3.4.2节）检查 `Data` 是否与PIT条目匹配。如果找不到匹配的PIT条目，则将 `Data` 提供给 **Data unsolicited** 管道；如果找到匹配的PIT条目，则将 `Data` 插入到 `ContentStore` 中。

   > 请注意，即使管道将 `Data` 插入到 `ContentStore` 中，该数据是否存储以及它在 `ContentStore` 中的停留时间也取决于 `ContentStore` 的接纳和替换策略（ *admission andreplacement policy*）。

2. 接着管道会将对应PIT条目的到期计时器设置为当前时间，调用对应策略的 `Strategy::afterReceiveData` 回调，将PIT标记为 *satisfied* ，并清除PIT条目的 *out records* 。

### 3.2 Data Unsolicited Pipeline

在 **Incoming data** 管道处理过程中发现 `Data` 是未经请求的时后会触发 **Data unsolicited** 管道处理逻辑，它的处理过程如下：

1. 根据当前配置的针对未经请求的 `Data` 的处理策略，决定是删除 `Data` 还是将其添加到 `ContentStore` 。默认情况下，MIR配置了 *drop-all* 策略，该策略会丢弃所有未经请求的 `Data` ，因为它们会对转发器造成安全风险。
2. 在某些特殊应用场景下，如果希望MIR将未经请求的 `Data` 存储到 `ContentStore`，可以在配置文件中修改对应的策略。

### 3.3 Outgoing Data Pipeline

在 **Incoming Interest** 管道（第4.2.1节）处理过程中在 `ContentStore` 中找到匹配的数据或在 **Incoming Data** 管道处理过程中发现传入的 `Data` 匹配到 PIT 表项时，调用本管道，它的处理过程如下：

1. 通过对应的 *LogicFace* 将 `Data` 发出。

## 4. Nack 处理路径

MIR中 `Nack` 的处理流程包含以下管道：

- **Incoming Nack** ：处理传入的 `Nack`
- **Outgoing Nack** ：准备和传出 `Nack`

### 4.1 Incoming Nack Pipeline

![incoming-nack-pipeline](https://gitee.com/quejianming/pic-bed/raw/master/uPic/2021/02/20/incoming-nack-pipeline-1613825225.svg)

如上图所示的是 **Incoming Nack** 管道的处理流程图，包括以下步骤：

1. 首先，从收到的 `Nack` 中提取到 `Interest`，然后查询是否有与之匹配的PIT条目，如果没有则丢弃，有则执行下一步；
2. 接着，判断匹配到的 PIT 条目中是否有对应 *LogicFace* 的 *out-record* ，如果没有则丢弃，有则执行下一步；
3. 然后，判断得到的 *out-record* 是否和 `Nack` 中的 `Interest` 的 `Nonce` 一致，不一致则丢弃，一致则执行下一步；
4. 然后标记对应的 *out-record* 为 *Nacked* ；
5. 如果此时对应的 PIT 条目中所有的 *out-record* 都已经 *Nacked* ，则将PIT条目的过期时间设置为当前时间（会触发 **Interest finalize** 管道）；
6. 然后调用对应策略的 `Strategy::afterReceiveNack` 回调，在其中触发 **Outgoing Nack** 管道。

### 4.2 Outgoing Nack Pipeline

本管道的处理流程包含以下步骤：

1. 首先，在PIT条目中查询指定的传出 *LogicFace* （下游）的 *in-record* 。该记录是必要的，因为协议要求将最后一个从下游接收到的 `Interest` （包括其Nonce）携带在 `Nack` 包中，如果未找到记录，请中止此过程，因为如果没有此兴趣，将无法发送 *Nack* 。
2. 然后构造一个 `Nack` 传递给下游，同时删除对应的 *in-record*

## 5. 通用推式包处理路径

MIR中 `GPPkt` 的处理流程包含以下管道：

- **Incoming GPPkt** ：处理传入的 `GPPkt`
- **Outgoing GPPkt** ：准备和传出 `GPPkt`

### 5.1 Incoming GPPkt Pipeline

![incoming-cpacket-pipeline](https://gitee.com/quejianming/pic-bed/raw/master/uPic/2021/02/20/incoming-cpacket-pipeline-1613828880.svg)

如上图所示的是 **Incoming GPPkt** 管道的处理流程图，包括以下步骤：

1. 首先给 `GPPkt` 的 `TTL` 减一，然后检查 `TTL` 的值是：

   - `TTL` < 0 则认为该兴趣包是一个回环的 `Interest` ，直接将其传递给 **Interest loop** 管道进一步处理；
   - `TTL` >= 0 则执行下一步。

   > 因为 GPPkt 是一种推式语义的网络包，不能向 `Interest` 那样通过 PIT 聚合来检测回环，所以这边和 IP 一样使用 TTL 来避免网络包无限回环。

2. 接着调用对应策略的 `Strategy::afterReceiveGPPkt` 回调，在其中触发 **Outgoing GPPkt **管道。

### 5.2 Outgoing GPPkt Pipeline

本管道的处理流程包含以下步骤：

1. 通过对应的 *LogicFace* 将 `GPPkt` 发出。





```go
type StrategyChoiceEntry struct {
  StrategyInstanceName string 			// 策略实例的名字
  prefix *Identifier								// 标识前缀
  strategy *Strategy								// 策略实例
}

// Get and Set functions

type StrategyChoiceTable interface {
  /**
  * 为指定的名字查找一个可以使用的策略实例
  */
  FindEffectiveStrategy(prefix *Identifier) (*Strategy, error)
}
```

