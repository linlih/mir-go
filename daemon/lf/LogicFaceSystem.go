//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/14 下午10:04
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

import "time"

//
// @Description:  全局logicFace表指针
//
var GLogicFaceTable *LogicFaceTable

//
// @Description:  用于保存IP：PORT信息和logicFace的映射关系
//			key 的格式是收到UDP包的 "<源IP地址>:<源端口号>"
//			value 的格式是logicFace对象指针
//		使用这个映射表的原因与gEtherAddrFaceMap类似
//		（1） 一个UDP端口13899可能会对应多个不同的logicFace。
//		（2）
//		（3） 在创建logicFace1时，与ether类型的logicFace不同的是，我们不会创建一个新的handle，而是一直使用logicFace0的handle。
//			TODO 这样做可能会有问题，现在还没考虑到，到时候改成新建一个handle也比较简单，现在先这么做
//
var gUdpAddrFaceMap *map[string]*LogicFace

//
// @Description:  用于保存mac地址和LogicFace的映射关系表。
//			key 的格式是收到以太网帧的 "<目的MAC地址>-<源MAC地址>"
//			value 是logicFace对象指针
//		使用这个映射表的原因在于：
//		（1） 一个物理网卡可能会对应多个logicFace，在MIR启动的时候，我们会启动一个以 "01:00:5e:00:17:aa" 为目的MAC地址的LogicFace，
//			我们先将这个logicFace称为logicFace0,这个logicFace0用于
//			接收从该网卡收到的以太网帧，同时也可以使用这个LogicFace0向该物理网卡对应的以太网发送以太网帧。由于使用该logicFace0发送的以太网帧的
//			目的MAC地址是"01:00:5e:00:17:aa"，是一个组播地址，所以这个logicFace0发出的以太网帧会被物理网卡所在的以太网中的所有其他网卡接收。
//		（2） 有时候，我们会需要创建一个一对一的以太网类型的logicFace，这时候我们会新建一个logicFace对应一个物理网卡，为了方便说明，
//			我们将这个新建的logicFace称为logicFace1。logicFace1的目的MAC地址假设是"fc:aa:14:cf:a6:97"，这是一个确切的对应特定物理网卡的地址
//			由于我们已经启动了logicFace0来接收物理网卡收到的所有MIN网络包，包括本应发往logicFace1的包，所以我们在创建logicFace1时不再启动收包协程，
//			这个时候，如果logicFace0收到了本应发往logicFace1的网络包，logicFace0会需要通过查找gEtherAddrFaceMap这个映射表，知道要调用logicFace1
//			的收到函数来处理网络分组。
//		（3） 在创建logicFace1时，我们会为logicFace1的etherTransport新创建一个pcap的handle用于发送网络包。
//
var gEtherAddrFaceMap *map[string]*LogicFace

//
// @Description: 全局logicFace系统
//
var gLogicFaceSystem *LogicFaceSystem

//
// @Description: 启动所有类型的Face监听
//
type LogicFaceSystem struct {
	ethernetListener EthernetListener
	tcpListener      TcpListener
	udpListener      UdpListener
	unixListener     UnixStreamListener
	logicFaceTable   *LogicFaceTable
	udpAddrFaceMap   map[string]*LogicFace
	etherAddrFaceMap map[string]*LogicFace
}

//
// @Description: 初始化LogicFaceSystem对象
// @receiver l
// @param table
//
func (l *LogicFaceSystem) Init(table *LogicFaceTable) {
	l.logicFaceTable = table
	l.ethernetListener.Init()
	l.tcpListener.Init()
	l.udpListener.Init()
	l.unixListener.Init()
	l.udpAddrFaceMap = make(map[string]*LogicFace)
	l.etherAddrFaceMap = make(map[string]*LogicFace)
	GLogicFaceTable = table
	gUdpAddrFaceMap = &l.udpAddrFaceMap
	gEtherAddrFaceMap = &l.etherAddrFaceMap

	gLogicFaceSystem = l
}

//
// @Description: 启动所有类型的Face监听,启用logicFace的清理协程
//		清理协程的工作机制是：每隔300秒扫描一篇logicFaceTable中的Face，如果logicFace在状态等于false，或者logicFace的超时时间已经过期，
//		则清理logicFace。
// @receiver l
//
func (l *LogicFaceSystem) Start() {
	l.ethernetListener.Start()
	l.tcpListener.Start()
	l.udpListener.Start()
	l.unixListener.Start()
	go l.faceCleaner()
}

func (l *LogicFaceSystem) destroyFace(logicFaceId uint64, logicFace *LogicFace) {
	if logicFace.logicFaceType == LogicFaceTypeUDP {
		delete(*gUdpAddrFaceMap, logicFace.transport.GetRemoteAddr())
	} else if logicFace.logicFaceType == LogicFaceTypeEther {
		delete(*gEtherAddrFaceMap, logicFace.transport.GetLocalAddr()+"-"+logicFace.transport.GetRemoteAddr())
	}
	l.logicFaceTable.RemoveByLogicFaceId(logicFaceId)
}

//
// @Description: 	篇历faceTable，清除过期或失效的logicFace
// @receiver l
//
func (l *LogicFaceSystem) doFaceClean() {
	curTime := getTimestampMS()
	for k, v := range GLogicFaceTable.mLogicFaceTable {
		if v.state == false {
			l.destroyFace(k, v)
		} else if v.expireTime < curTime { // logicFace已经超时
			v.Shutdown()        // 调用shutdown关闭logicFace
			l.destroyFace(k, v) // 将logicFace从全局logicFaceTable中删除
		}
	}
}

//
// @Description:  由协程调用，每300秒执行一个清表操作
// @receiver l
//
func (l *LogicFaceSystem) faceCleaner() {
	for true {
		l.doFaceClean()
		time.Sleep(time.Second * 300)
	}
}

//
// @Description: 	获取当前unix时间， 单位是 ms
// @return int64
//
func getTimestampMS() int64 {
	curTime := time.Now().UnixNano() / 1000000
	return curTime
}
