# MIR Management

 MIR管理通信协议允许用户、工具和控制平面程序检索，监视和更改MIR的状态。

## 1. MIR管理协议

### 1.1 基本机制

MIR通信管理协议使用管理通信包（Interest-Data exchange）来进行通信，管理通信包的顶层TLV的TYPE值为1。

- **Status Dataset** （状态数据集）

  > https://redmine.named-data.net/projects/nfd/wiki/StatusDataset

  状态数据集定义了一个实体集合的状态信息如何作为管理数据包（Management Data）进行编码、分片和发布。这个机制搭配兴趣包数据包交换（Interest-Data exchange）的通信方式可以很方便的对MIR中的各种状态信息进行检索。

- **Notification Stream**（通知流）=> 初版设计暂不实现

  MIR管理通信协议要求实现一个基于事件订阅的通知机制，一旦MIR的相关状态发生改变，可以通过该机制通知对应的管理模块。这个机制可以很有效的对MIR的状态进行监视。

- **Control Command**（控制命令）

  控制命令定义了一系列可以更改MIR状态的命令的请求和回复格式，*以及如何对这些命令进行签名和认证*。这个机制对于更改MIR的状态很有用。

![management](https://gitee.com/quejianming/pic-bed/raw/master/uPic/2021/03/10/management-1615364676.svg)

如上图所示的是 MIR 中管理模块的交互示意图。管理模块通过 `Dispatcher` 与路由器进行数据交互，具体的交互流程如下：

1. 首先，路由器每收到一个管理兴趣包 `CommandInterest`，就会传递给 `Dispatcher` 处理；
2. `Dispatcher` 会对收到的 `CommandInterest` 进行解析、授权验证、参数校验等一系列操作，如果过程中发现授权不通过或者参数校验失败，则直接返回错误信息；
3. 所有校验都通过之后，会调用对应模块的对应方法去处理 `CommandInterest`；
4. 每个管理模块处理完之后将生成的响应数据传递给 `Dispatcher`，通过 `Dispatcher` 发送到路由器当中。

### 1.2 模块

- **LogicFace Management**（逻辑接口管理模块）
  - `add` => 一个控制命令，用于添加一个逻辑接口
  - `del` => 一个控制命令，用于删除一个逻辑接口
  - `list` => 一个数据集（dataset）用于发布发布 LogicFace 状态和计数器；
- **FIB Management**（转发表管理模块）
  - 
  - 插入、更新和删除FIB条目的控制命令；
  - 一个数据集（dataset）用于发布FIB表的条目信息；
- **CS Management**（缓存管理模块）

### 1.3 管理请求包的基本格式

```
CommandInterest = 1 TLV-LENGTH
             { InterestIdentifier }        => 标识区
             { Signature }                 => 签名区
```

- **用于本地管理**：

  ```
  /min-mir/mgmt/localhost/<模块名称>/<命令>/<参数>/[版本号]/[分片号]
  ```

  - 模块名称：每个管理模块会有一个唯一的模块名称，例如：`fib-mgmt`

  - 命令：表示要在对应模块执行的动作，例如：`add`、`list`

  - 参数：由一个 `ControlParameters` TLV 编码而成，其中包含一系列子 TLV，每个子TLV表示一个请求的参数字段（Field），例如：

    ```
    ControlParameters = CONTROL-PARAMETERS-TYPE TLV-LENGTH
                        [LocalFaceUri]
                        [RemoteFaceUri]
    ```

  - 版本号：数据集在每次版本更迭时，会将发布的数据状态的版本号加1，可以凭借版本号获取最新的状态信息

  - 分片号：数据集通常无法用一个Data包装下，需要进行分片发布，分片号指定要拉取的数据分片

- **用于单跳接入**（用于使用TCP LogicFace等接入MIR的场景）

  ```
  /min-mir/mgmt/localhop/<模块名称>/<命令>/<参数>/[版本号]/[分片号]
  ```
  
- **用于远程管理**

  ```
  <路由器前缀>/min-mir/mgmt/<模块名称>/<命令>/<参数>/[版本号]/[分片号]
  ```

### 1.4 管理回复包的基本格式

```
CommandData = 1 TLV-LENGTH
             { DataIdentifier }            => 标识区
             { Signature }                 => 签名区
             {                             => 只读区
                 <Payload>
             }
```

回复数据放在管理回复包的 `Payload` 部分，格式为json格式，如果数据较多，需要分片进行传输

## 2. Dispatcher

`Dispatcher` 是管理模块与 MIR 交互的统一出入口，负责对所有的管理请求进行有效性校验，并把管理模块返回的响应数据发送给 MIR，主要具备以下几点功能：

- 每个管理模块都可以在 `Dispatcher` 中注册可以处理的命令以及数据集；
- `Dispatcher` 负责与MIR进行具体的通信，收到命令兴趣包之后进行解析，解析完分发给对应的管理模块进行处理，管理模块处理完成后，再将处理的结果通过 `Dispatcher` 发送给MIR；
- 内置一个小型的内存缓存（与MIR中的CS相互独立，互不相关），主要用于缓存数据集分段后的 `Data`；

```go
//
// @Author: Jianming Que
// @Description: 
// @Version: 1.0.0
// @Date: 2021/3/10 10:07 上午  
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package mgmt

import (
	"minlib/component"
	"minlib/mgmt"
	"minlib/packet"
)

//
// 授权通过回调
//
// @Description:
//
type AuthorizationAccept func()

//
// 授权拒绝回调
//
// @Description:
//
type AuthorizationReject func()

//
// 一个回调函数，用于对收到的控制命令进行授权验证
//
// @Description:
//
// @param topPrefix	顶级管理前缀，例如 "/min-mir/mgmt/localhost，可以通过本参数实现以下控制需求
//					1. 比如可以控制只有指定的顶级管理前缀可以授权通过，其它都不行，例如：/min-mir/mgmt/localhost 合法，其它前缀均不合法；
//					2. 也可以实现对不同的顶级前缀实现不同级别的授权，比如 "/min-mir/mgmt/localhost" 认为是本地管理员，拥有较高权限，默认
//					   可以控制和获取路由器状态；"/<路由器前缀>/min-mir/mgmt/remote" 认定为远程管理员发过来的命令，拥有较局限的权限，只能
//					   做基本的获取状态操作，如果要修改路由器状态，需要进一步的授权。
//
// @param interest		收到的命令兴趣包
// @param parameters	命令兴趣包中携带的参数
// @param accept		授权通过则调用此回调
// @param reject		授权拒绝则调用此回调
// @return unc
//
type Authorization func(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *mgmt.ControlParameters,
	accept AuthorizationAccept,
	reject AuthorizationReject)

//
// 一个回调函数，用于对收到的控制命令参数进行验证
//
// @Description:
// @param parameters	要验证的参数
// @return bool
//
type ValidateParameters func(parameters *mgmt.ControlParameters) bool

//
// 一个回调函数，用于对收到的已授权的命令进行处理（每个管理模块会通过传入本回调函数自己实现对特定命令的处理逻辑）
//
// @Description:
// @param topPrefix		顶级管理前缀，例如："/min-mir/mgmt/localhost"
// @param interest		收到的命令兴趣包
// @param parameters	已通过参数验证的命令参数
// @return *mgmt.ControlResponse	返回一个 ControlResponse 返回给调用方
//
type ControlCommandHandler func(topPrefix *component.Identifier, interest *packet.Interest,
	parameters *mgmt.ControlParameters) *mgmt.ControlResponse

//
// 一个回调函数，处理收到的请求数据集的命令
//
// @Description:
// @param topPrefix
// @param interest
// @param context
//
type StatusDatasetHandler func(topPrefix *component.Identifier, interest *packet.Interest,
	context *StatusDatasetContext)

//
// 兴趣包处理回调
//
// @Description:
// @param topPrefix
// @param interest
//
type InterestHandler func(topPrefix *component.Identifier, interest *packet.Interest)

//
// 定义一个管理命令解析和分发程序
//
// @Description:
//	1. 每个管理模块实现时，需要调用本类进行命令注册；
//	2. 本类负责与MIR进行具体的通信，收到命令兴趣包之后进行解析，解析完分发给对应的管理模块进行处理，管理模块处理完成后，再将处理的结果通过
//	   本类发送给MIR；
//
type IDispatcher interface {
	//
	// 添加一个顶层前缀
	//
	// @Description:
	// @param topPrefix
	//
	AddTopPrefix(topPrefix *component.Identifier)

	//
    // 移除一个顶层前缀
    //
    // @Description:
    // @param topPrefix
    //
	RemoveTopPrefix(topPrefix *component.Identifier)

	//
    // 注册一个控制命令
    //
    // @Description:
    // @param relPrefix				模块名 + 命令名，例如："/fib-mgmt/add"
    // @param authorization			授权验证回调
    // @param validateParameters	参数验证回调
    // @param handler				命令处理回调
    //
	AddControlCommand(relPrefix *component.Identifier, authorization Authorization, validateParameters ValidateParameters,
		handler ControlCommandHandler)

	//
    // 注册一个数据集
    //
    // @Description:
    //	1. 对于收到的请求数据集的 Interest，首先判断是否携带版本号和分段号，如果携带则直接不响应（因为不带版本号和分段号的请求认为是请求的起始，
    //		会触发数据集收集、分段和发布，携带版本号和分段号的请求认为是对某个具体的数据集的快照的请求，只能期望在缓存中命中，如果没有命中缓存，则应该
    //		重新发起一个不带版本号和序列号的请求，以触发数据集的收集、分段和发布）；
    //	2. 接着就使用 authorization 对请求进行授权鉴定，授权没有通过，则返回错误信息，通过则执行下一步；
    //	3. 然后触发 handler 回调，让对应的管理模块对数据集进行收集、分段和发布。
    // @param relPrefix				模块名 + 命令名，例如："/fib-mgmt/add"
    // @param authorization			授权验证回调
    // @param handler				数据集处理回调
    //
	AddStatusDataset(relPrefix *component.Identifier, authorization Authorization, handler StatusDatasetHandler)

	//
    // 发送一个控制命令应答
    //
    // @Description:
    // @param response	要响应的应答数据
    // @param interest	收到的命令请求兴趣包
    //
	sendControlResponse(response *mgmt.ControlResponse, interest *packet.Interest)

	//
    // 发送一个数据包到 MIR
    //
    // @Description:
    // @param data
    //
	sendData(data *packet.Data)

	//
    // 查询 Dispatcher 内置缓存，尝试获取数据集分段数据
    //
    // @Description:
    // @param topPrefix
    // @param interest
    // @param missStorage
    //
	queryStorage(topPrefix *component.Identifier, interest *packet.Interest, missStorage InterestHandler)
}
```

## 3. 请求参数TLV

```go
// 管理通信协议
TlvManagementControlParameters = 220 // 控制命令参数
TlvManagementLogicFaceId       = 221 // 逻辑接口Id
TlvManagementUri               = 222 // 逻辑接口地址，例如: tcp://192.168.1.1:13899，存在localUri时，也可表示remoteUri

// 管理通信协议 -> LogicFaceManager
TlvManagementCost                 = 223 // 链路开销
TlvManagementLogicFacePersistency = 224 // 逻辑接口持久性
TlvManagementUriScheme			  = 225 // 逻辑接口地址采用的模式，例如：tcp
TlvManagementMtu				  = 226 // 最大传输单元
```

## 3. LogicFace Management

> 模块名称：`minlf-mgmt`

### 3.1 控制命令

- **`add`**

  > add 命令用于添加一个逻辑接口

  - 命令行工具命令

    ```bash
    mirc lf add remote <LFURI> [persistency <PERSISTENCY>] [local <LFURI>] [mtu <MTU>]
    ```

  - 请求参数

    在命令兴趣包的参数 `ControlParameters` 部分，需要填充以下参数：

    - < `Uri` > : 远端地址
    - [ `LocalUri` ] : 本地地址
    - < `Persistency` > : 接口持久性
    - [ `Mtu` ] : 最大传输单元

  - 回复数据格式：

    ```json
    // 操作成功
    {
      "code": 200,
      "errMsg": "",
    }
    
    // 操作失败
    {
      "code": 400,
      "errMsg": "Missing parameter Uri"
    }
    ```

- **`del`**

  > del 命令用于删除一个逻辑接口

  - 命令行工具命令

    ```bash
    mirc lf del <LFID|LFURI>
    ```

  - 请求参数

    在命令兴趣包的参数 `ControlParameters` 部分，需要填充以下参数：

  - 

- **`show`**

  > show 命令用于展示指定ID的逻辑接口的信息

  ```
  mirc lf show <LFID>
  ```

### 3.2 数据集

- **`list`**

  > list 命令用于显示逻辑接口的详细信息

  ```
  mirc lf list [remote <LFURI>] [local <LFURI>] [scheme <SCHEME>]
  ```

  - 请求参数：

    - [`remoteUri`]
    - [`localUri`]
    - [`scheme`]

  - 回复数据格式：

    ```json
    {
      "code": 200,
      "errMsg": "",
      "data": [
        {
          "lfId": 5,
          "remoteUri": "tcp://192.168.1.2:13899",
          "localUri": "tcp://192.168.1.3:19533",
          "mtu": 7000,
          <Face 的详细信息，等Face设计完毕>
        }
      ]
    }
    ```

### 2.2 OPTIONS

- **LFID**

  LfId 表示逻辑接口在MIR中的唯一数字标识

- **LFURI**

  LfUri 表示本地或者远端的逻辑接口的地址，示例如下：

  - `tcp://192.168.1.2:13899`
  - `udp://192.168.1.3:13899`
  - `ether://[08:00:27:01:01:01]`
  - `dev://eth0`
  - `unix:///var/run/mir.sock`

- **SCHEME**

  Scheme 表示本地或者远端的接口地址所使用的Uri方案，示例如下：

  - udp
  - tcp
  - unix
  - dev

- **PERSISTENCY**

  Persistency 参数指定了逻辑接口的持久性，取值可以为 `persistent` （持久的）和 `permanent`（永久的）

  - `persistent`：拥有 `persistent` 持久性的逻辑接口，在通信过程中发生套接字错误时，会自动关闭并销毁该逻辑接口；
  - `permanent`：拥有 `permanent` 持久性的逻辑接口，在通信过程中如果发生套接字错误是，逻辑接口不会直接销毁，会一直保留，并尝试重新建立套接字连接。

- **MTU**

  Mtu 参数指定了逻辑接口的最大传输单元的大小。
  
- **COST**

  链路开销

## 3. FIB Management

> 模块名称：`fib-mgmt`

### 3.1 控制命令

- **`list`**

  > list 命令用于展示 FIB 表的信息

  `mirc fib list `

- **`add`**

- **`del`**