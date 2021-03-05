/**
 * @Author: wzx
 * @Description:
 * @Version: 1.0.0
 * @Date: 2021/3/4 上午12:48
 *@Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
 */

package table

import (
	"minlib/component"
)

type NextHop struct {
	LogicFaceId uint64
	Cost        uint64
}

type FIBEntry struct {
	Identifier  *component.Identifier
	NextHopList map[uint64]NextHop
}

func(f *FIBEntry) GetIdentifier()component.Identifier{return *f.Identifier}

func(f *FIBEntry) GetNextHops()[]NextHop{
	res := make([]NextHop, 0)

	for _, nextHop := range f.NextHopList{
		res = append(res, nextHop)
	}

	return res
}

func(f *FIBEntry) HasNextHops()bool{return  f.NextHopList == nil || len(f.NextHopList) == 0}

func(f *FIBEntry) HasNextHop(logicFaceId uint64) bool{

	for _, nextHop := range f.NextHopList{
		if nextHop.LogicFaceId == logicFaceId{
			return true
		}
	}

	return false
}

func(f *FIBEntry) AddOrUpdateNextHop(logicFaceId uint64, cost uint64)bool{
	res := f.NextHopList == nil || len(f.NextHopList) ==0
	if f.NextHopList == nil{
		f.NextHopList = make(map[uint64]NextHop)
	}
	delete(f.NextHopList, logicFaceId)
	f.NextHopList[logicFaceId] = NextHop{LogicFaceId: logicFaceId, Cost: cost}
	return res
}

func(f *FIBEntry) RemoveNextHop(logicFaceId uint64){
	if f.NextHopList == nil{
		f.NextHopList = make(map[uint64]NextHop)
	}
	delete(f.NextHopList, logicFaceId)
}