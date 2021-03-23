//
// @Author: weiguohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/14 下午10:04
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package lf

var gLogicFaceTable *LogicFaceTable
var gUdpAddrFaceMap *map[string]*LogicFace
var gEtherAddrFaceMap *map[string]*LogicFace

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
	gLogicFaceTable = table
	gUdpAddrFaceMap = &l.udpAddrFaceMap
	gEtherAddrFaceMap = &l.etherAddrFaceMap

	gLogicFaceSystem = l
}

//
// @Description: 启动所有类型的Face监听
// @receiver l
//
func (l *LogicFaceSystem) Start() {
	l.ethernetListener.Start()
	l.tcpListener.Start()
	l.udpListener.Start()
	l.unixListener.Start()
}
