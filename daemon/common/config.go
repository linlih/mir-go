//
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/4/2 10:20 上午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package common

import "gopkg.in/ini.v1"

//
// 表示 MIR 配置文件的配置，与 mirconf.ini 中的配置一一对应
//
// @Description:
//
type MIRConfig struct {
	GeneralConfig   `ini:"General"`
	LogConfig       `ini:"Log"`
	TableConfig     `ini:"Table"`
	LogicFaceConfig `ini:"LogicFace"`
	SecurityConfig  `ini:"Security"`
}

type GeneralConfig struct {
	////////////////////////////////////////////////////////////////////////////////////////////////
	//// General
	////////////////////////////////////////////////////////////////////////////////////////////////
	DefaultId      string `ini:"DefaultId"`      // 默认网络身份
	IdentifierType int    `ini:"IdentifierType"` // 当前路由器支持的标识类型，102 => CPacket | 103 => 内容兴趣标识（Interest）| 103 => 内容兴趣标识（Interest）
}

type LogConfig struct {
	////////////////////////////////////////////////////////////////////////////////////////////////
	//// Log
	////////////////////////////////////////////////////////////////////////////////////////////////
	LogLevel     string `ini:"LogLevel"`     // 日志等级
	ReportCaller bool   `ini:"ReportCaller"` // 日志输出时是否添加文件名和函数名
	LogFormat    string `ini:"LogFormat"`    // 输出日志的格式 "json" | "text"
	LogFilePath  string `ini:"LogFilePath"`  // 日志输出文件路径，为空则输出至控制台
}

type TableConfig struct {
	////////////////////////////////////////////////////////////////////////////////////////////////
	//// Table
	////////////////////////////////////////////////////////////////////////////////////////////////
	CSSize            int    `ini:"CSSize"`            // CS缓存大小，包为单位
	CSReplaceStrategy string `ini:"CSReplaceStrategy"` // 缓存替换策略
}

type LogicFaceConfig struct {
	////////////////////////////////////////////////////////////////////////////////////////////////
	//// LogicFace
	////////////////////////////////////////////////////////////////////////////////////////////////
	SupportTCP  bool   `ini:"SupportTCP"`  // 是否开启TCP
	TCPPort     int    `ini:"TCPPort"`     // TCP 端口号
	SupportUDP  bool   `ini:"SupportUDP"`  // 是否开启UDP
	UDPPort     int    `ini:"UDPPort"`     // UDP 端口号
	SupportUnix bool   `ini:"SupportUnix"` // 是否开启Unix
	UnixPath    string `ini:"UnixPath"`    // Unix 套接字路径设置
}

type SecurityConfig struct {
	////////////////////////////////////////////////////////////////////////////////////////////////
	//// Security
	////////////////////////////////////////////////////////////////////////////////////////////////
	VerifyPacket          bool `ini:"VerifyPacket"`          // 是否开启包签名验证
	Log2BlockChain        bool `ini:"Log2BlockChain"`        // 是否发送日志到区块链
	MiddleRouterSignature bool `ini:"MiddleRouterSignature"` //是否开启中间路由器签名
	MaxRouterSignatureNum int  `ini:"MaxRouterSignatureNum"` // 最大中间路由器签名数量
}

//
// 解析配置文件
//
// @Description:
// @receiver m
// @param configPath
// @return error
//
func ParseConfig(configPath string) (*MIRConfig, error) {
	cfg, err := ini.Load(configPath)
	if err != nil {
		return nil, err
	}
	mirConfig := new(MIRConfig)
	if err = cfg.MapTo(&mirConfig); err != nil {
		return nil, err
	}
	return mirConfig, nil
}
