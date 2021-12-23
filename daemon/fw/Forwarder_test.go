package fw

import (
	"fmt"
	"minlib/component"
	"minlib/packet"
	"minlib/utils"
	"mir-go/daemon/lf"
	"mir-go/daemon/plugin"
	"testing"
)

func TestForwarder_Init(t *testing.T) {
	forwarder := new(Forwarder)
	newPlugin := new(plugin.GlobalPluginManager)
	queue := utils.NewBlockQueue(20)
	forwarder.Init(nil, newPlugin, queue)
	fmt.Println("forwarder", forwarder.FIB.GetDepth(), forwarder.PIT.Size())
	face := new(lf.LogicFace)
	face.LogicFaceId = 234
	interest := new(packet.Interest)
	interest.TTL.SetTtl(3)
	newName, _ := component.CreateIdentifierByString("/min/pkusz")
	interest.SetName(newName)
	interest.Nonce.SetNonce(13451310354534135)
	interest.InterestLifeTime.SetInterestLifeTime(4000)
	data := new(packet.Data)
	data.FreshnessPeriod.SetFreshnessPeriod(5)
	data.SetName(newName)
	forwarder.ICS.Insert(data)
	forwarder.OnIncomingInterest(face, interest)
	//pitEntry,piterr:=forwarder.PIT.Find(interest)
	//if piterr!=nil{
	//	fmt.Println("piterr",piterr)
	//}
	//fmt.Println("pit entry",pitEntry.Identifier.ToUri(),pitEntry.InRecordList,pitEntry.OutRecordList)
	fmt.Println("PIT", forwarder.PIT.Size())
	//time.Sleep(time.Duration(4)*time.Second)
	//fmt.Println("PIT",forwarder.PIT.Size())
	csEntry, _ := forwarder.ICS.Find(interest)
	fmt.Println("cs entry", csEntry.Interest.ToUri(), csEntry.Interest.InterestLifeTime, csEntry.Interest.TTL, csEntry.Interest.Nonce)

}

func TestForwarder_OnOutgoingInterest(t *testing.T) {
	//strategy:=new(table.StrategyTable)
	//strategy.SetDefaultStrategy("/")
	newName1, _ := component.CreateIdentifierByString("/min")
	forwarder := new(Forwarder)
	newPlugin := new(plugin.GlobalPluginManager)
	queue := utils.NewBlockQueue(20)
	forwarder.Init(nil, newPlugin, queue)
	forwarder.SetDefaultStrategy("/")
	forwarder.StrategyTable.Init()

	brs := BestRouteStrategy{StrategyBase{forwarder: forwarder}}
	forwarder.StrategyTable.Insert(newName1, "best", &brs)

	fmt.Println("forwarder", forwarder.FIB.GetDepth(), forwarder.PIT.Size())
	face := new(lf.LogicFace)
	face.LogicFaceId = 234
	interest := new(packet.Interest)
	interest.TTL.SetTtl(3)
	newName, _ := component.CreateIdentifierByString("/min/pkusz")
	interest.SetName(newName)
	interest.Nonce.SetNonce(13451310354534135)
	interest.InterestLifeTime.SetInterestLifeTime(2000)
	forwarder.FIB.AddOrUpdate(newName1, face, 233)
	forwarder.OnIncomingInterest(face, interest)
	pitEntry, piterr := forwarder.PIT.Find(interest)
	if pitEntry != nil {
		fmt.Println("pitEntry empty")
	}
	if piterr == nil {
		fmt.Println("piterr", piterr)
	}
	//fmt.Println("pit entry", pitEntry)
	fmt.Println("PIT", forwarder.PIT.Size())
	fmt.Println("FIB", forwarder.FIB.Size())
}

func TestForwarder_OnInterestLoop(t *testing.T) {
	newName1, _ := component.CreateIdentifierByString("/min")
	forwarder := new(Forwarder)
	newPlugin := new(plugin.GlobalPluginManager)
	queue := utils.NewBlockQueue(20)
	forwarder.Init(nil, newPlugin, queue)
	forwarder.SetDefaultStrategy("/")
	forwarder.StrategyTable.Init()

	brs := BestRouteStrategy{StrategyBase{forwarder: forwarder}}
	forwarder.StrategyTable.Insert(newName1, "best", &brs)

	fmt.Println("forwarder", forwarder.FIB.GetDepth(), forwarder.PIT.Size())
	face := new(lf.LogicFace)
	face.LogicFaceId = 234
	interest := new(packet.Interest)
	interest.TTL.SetTtl(1)
	newName, _ := component.CreateIdentifierByString("/min/pkusz")
	interest.SetName(newName)
	interest.Nonce.SetNonce(13451310354534135)
	interest.InterestLifeTime.SetInterestLifeTime(2000)
	forwarder.FIB.AddOrUpdate(newName1, face, 233)
	forwarder.OnIncomingInterest(face, interest)
	pitEntry, piterr := forwarder.PIT.Find(interest)
	if pitEntry != nil {
		fmt.Println("pitEntry empty")
	}
	if piterr == nil {
		fmt.Println("piterr", piterr)
	}
	//fmt.Println("pit entry", pitEntry)
	fmt.Println("PIT", forwarder.PIT.Size())
	fmt.Println("FIB", forwarder.FIB.Size())
}

func TestForwarder_OnContentStoreHit(t *testing.T) {
	newName1, _ := component.CreateIdentifierByString("/min")

	forwarder := new(Forwarder)
	newPlugin := new(plugin.GlobalPluginManager)
	queue := utils.NewBlockQueue(20)
	forwarder.Init(nil, newPlugin, queue)
	fmt.Println("forwarder", forwarder.FIB.GetDepth(), forwarder.PIT.Size())
	brs := BestRouteStrategy{StrategyBase{forwarder: forwarder}}
	forwarder.StrategyTable.Insert(newName1, "best", &brs)

	face := new(lf.LogicFace)
	face.LogicFaceId = 234
	interest := new(packet.Interest)
	interest.TTL.SetTtl(3)
	newName, _ := component.CreateIdentifierByString("/min/pkusz")
	interest.SetName(newName)
	interest.Nonce.SetNonce(13451310354534135)
	interest.InterestLifeTime.SetInterestLifeTime(4000)
	data := new(packet.Data)
	data.FreshnessPeriod.SetFreshnessPeriod(5)
	data.SetName(newName)
	forwarder.ICS.Insert(data)
	forwarder.OnIncomingInterest(face, interest)
	//pitEntry,piterr:=forwarder.PIT.Find(interest)
	//if piterr!=nil{
	//	fmt.Println("piterr",piterr)
	//}
	//fmt.Println("pit entry",pitEntry.Identifier.ToUri(),pitEntry.InRecordList,pitEntry.OutRecordList)
	fmt.Println("PIT", forwarder.PIT.Size())
	//time.Sleep(time.Duration(4)*time.Second)
	//fmt.Println("PIT",forwarder.PIT.Size())
	csEntry, _ := forwarder.ICS.Find(interest)
	fmt.Println("cs entry", csEntry.Interest.ToUri(), csEntry.Interest.InterestLifeTime, csEntry.Interest.TTL, csEntry.Interest.Nonce)

}
