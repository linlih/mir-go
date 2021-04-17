/**
 * @Author: yzy
 * @Description:
 * @Version: 1.0.0
 * @Date: 2021/3/30 下午6:54
 *@Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
 */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/urfave/cli/v2"
	"minlib/common"
	"minlib/component"
	"minlib/encoding"
	"minlib/logicface"
	mgmtlib "minlib/mgmt"
	"minlib/packet"
	"mir-go/daemon/mgmt"
)

var remote string

const moduleName = "face-mgmt"

const (
	actionList = "list"
	actionAdd  = "add"
	actionDel  = "del"
)

var faceCommands = cli.Command{
	Name:        "lf",
	Usage:       "logic Face Management",
	Subcommands: []*cli.Command{&GetFaceInfoCommand, &CreateNewFaceCommand, &DestroyFaceCommand},
}

var GetFaceInfoCommand = cli.Command{
	Name:   "list",
	Usage:  "Show all face info",
	Action: GetAllFaceInfo,
}

var CreateNewFaceCommand = cli.Command{
	Name:   "add",
	Usage:  "Create new face",
	Action: CreateNewFace,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "remote",
			Value:    "",
			Usage:    "remote address for connect",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "local",
			Value: "",
			Usage: "local address for accept",
		},
		//&cli.StringFlag{
		//	Name:  "scheme",
		//	Value: "",
		//	Usage: "create connection type",
		//}
	},
}

var DestroyFaceCommand = cli.Command{
	Name:   "del",
	Usage:  "Delete face",
	Action: DeleteFace,
	Flags: []cli.Flag{
		&cli.StringFlag{
			//远端地址
			Name:        "id",
			Value:       "",
			Destination: &remote,
		},
	},
}

func GetAllFaceInfo(c *cli.Context) error {
	// 接入路由器
	face := logicface.LogicFaceICN{}
	err := face.InitWithUnixSocket(unixPath)
	if err != nil {
		common.LogError("connect MIR fail!the err is:", err)
		return err
	}

	// 首先拉取到元数据
	face.ExpressInterest(newCommandInterest(moduleName, actionList),
		func(interest *packet.Interest, data *packet.Data) {
			// OnData
			common.LogInfo("OnData")
			var responseHeader *mgmt.ResponseHeader
			_ = json.Unmarshal(data.GetValue(), &responseHeader)

			bytesBuilder := bytes.Buffer{}
			for i := 1; i <= responseHeader.FragNums; i++ {
				identifierFrag := data.GetName()
				identifierFrag.Append(component.CreateIdentifierComponentByNonNegativeInteger(uint64(i)))
				interest.SetName(identifierFrag)
				interest.SetTtl(5)
				interest.InterestLifeTime.SetInterestLifeTime(4000)

				face.SendInterest(interest)
				data, _ := face.ReceiveData()
				bytesBuilder.Write(data.Payload.GetValue())
			}
			var faceInfoList []mgmt.FaceInfo
			err = json.Unmarshal(bytesBuilder.Bytes(), &faceInfoList)
			if err != nil {
				common.LogError("parse data fail!the err is:", err)
				return
			}
			for _, v := range faceInfoList {
				fmt.Printf("%+v\n", v)
			}
		}, func(interest *packet.Interest) {
			// OnTimeout
			common.LogInfo("OnTimeout")
		}, func(interest *packet.Interest, nack *packet.Nack) {
			// OnNack
			common.LogInfo("OnNack")
		})
	face.ProcessEvent()
	return nil
}

func CreateNewFace(c *cli.Context) error {
	face := logicface.LogicFaceICN{}
	// 建立unix连接
	err := face.InitWithUnixSocket(unixPath)
	if err != nil {
		common.LogError("connect MIR fail!the err is:", err)
		return err
	}
	commandInterest := newCommandInterest(moduleName, actionAdd)
	commandInterest.GetName()

	parameters := &component.ControlParameters{}
	parameters.SetUriScheme(c.Uint64("scheme"))
	parameters.SetLocalUri(c.String("local"))
	parameters.SetUri(c.String("remote"))

	if err := commandInterest.AppendCommandParameters(parameters); err != nil {
		common.LogFatal("Append parameters failed!")
	}

	face.ExpressInterest(commandInterest,
		func(interest *packet.Interest, data *packet.Data) {
			// OnData
			var response mgmtlib.ControlResponse
			if err := json.Unmarshal(data.GetValue(), &response); err != nil {
				common.LogError("parse data fail!the err is:", err)
				return
			}
			fmt.Printf("%+v\n", response)
		}, func(interest *packet.Interest) {
			// OnTimeout
			common.LogError("onTimeout")
		}, func(interest *packet.Interest, nack *packet.Nack) {
			// OnNack
			common.LogError("onNack")
		})
	face.ProcessEvent()
	return nil
}

func DeleteFace(c *cli.Context) error {
	face := logicface.LogicFace{}
	// 建立unix连接
	err := face.InitWithUnixSocket("/tmp/mirsock")
	if err != nil {
		common.LogError("connect MIR fail!the err is:", err)
		return err
	}
	interest := &packet.Interest{}
	identifierHead, _ := component.CreateIdentifierByString("/min-mir/mgmt/localhost/face-mgmt/destroy")
	interest.SetName(identifierHead)
	interest.SetTtl(5)
	interest.InterestLifeTime.SetInterestLifeTime(4000)
	parameters := &component.ControlParameters{}
	parameters.SetLogicFaceId(c.Uint64("id"))
	var encoder = &encoding.Encoder{}
	encoder.EncoderReset(encoding.MaxPacketSize, 0)
	parameters.WireEncode(encoder)
	buf, _ := encoder.GetBuffer()
	identifierHead.Append(component.CreateIdentifierComponentByByteArray(buf))
	face.SendInterest(interest)
	data, err := face.ReceiveData()
	var response mgmtlib.ControlResponse
	if err := json.Unmarshal(data.GetValue(), &response); err != nil {
		common.LogError("parse data fail!the err is:", err)
		return err
	}
	fmt.Printf("%+v\n", response)
	return nil
}
