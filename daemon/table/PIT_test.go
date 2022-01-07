// Copyright [2022] [MIN-Group -- Peking University Shenzhen Graduate School Multi-Identifier Network Development Group]
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Package table
// @Author: yzy
// @Description:
// @Version: 1.0.0
// @Date: 2021/12/21 4:32 PM
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//

package table

import (
	"fmt"
	"minlib/component"
	"minlib/packet"
	"mir-go/daemon/lf"
	"testing"
	//"strconv"
)

func TestInsert(t *testing.T) {
	pit := CreatePIT()
	identifier, err := component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	interest := &packet.Interest{}

	// 测试 异常值插入
	var PrefixList []string
	interest.SetName(&component.Identifier{})
	for _, v := range pit.Insert(interest).Identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	fmt.Println(PrefixList)
	fmt.Println(pit.lpm.table)

	//测试 正常插入
	interest.SetName(identifier)
	for _, v := range pit.Insert(interest).Identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	fmt.Println(PrefixList)
	fmt.Println(pit.lpm.table)

}

func TestSize(t *testing.T) {
	pit := CreatePIT()
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	interest := &packet.Interest{}
	interest.SetName(identifier)
	pit.Insert(interest)
	fmt.Println(pit.Size())

	identifier, err = component.CreateIdentifierByString("/min/pku")
	if err != nil {
		fmt.Println(err)
	}
	interest.SetName(identifier)
	pit.Insert(interest)
	fmt.Println(pit.Size())

	identifier, err = component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	interest.SetName(identifier)
	pit.Insert(interest)
	fmt.Println(pit.Size())

	identifier, err = component.CreateIdentifierByString("/min/pku/cn")
	if err != nil {
		fmt.Println(err)
	}
	interest.SetName(identifier)
	pit.Insert(interest)
	fmt.Println(pit.Size())

	identifier, err = component.CreateIdentifierByString("/min/gdcni15")
	if err != nil {
		fmt.Println(err)
	}
	interest.SetName(identifier)
	pit.Insert(interest)
	fmt.Println(pit.Size())

	identifier, err = component.CreateIdentifierByString("/min/gdcni15/mir")
	if err != nil {
		fmt.Println(err)
	}
	interest.SetName(identifier)
	pit.Insert(interest)
	fmt.Println(pit.Size())

	identifier, err = component.CreateIdentifierByString("/min/gdcni15/filegator")
	if err != nil {
		fmt.Println(err)
	}
	interest.SetName(identifier)
	pitEntry := pit.Insert(interest)
	fmt.Println(pit.Size())
	// 删除一个再测
	pit.EraseByPITEntry(pitEntry)
	fmt.Println(pit.Size())
}

func TestFind(t *testing.T) {
	pit := CreatePIT()
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	interest := &packet.Interest{}
	interest.SetName(identifier)
	pit.Insert(interest)
	// 找到数据
	fmt.Println(pit.Find(interest))

	// 没有找到数据
	interest.SetName(&component.Identifier{})
	fmt.Println(pit.Find(interest))

}

func TestFindDataMatches(t *testing.T) {
	pit := CreatePIT()
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	interest := &packet.Interest{}
	interest.SetName(identifier)
	pit.Insert(interest)
	fmt.Println(pit.Size())
	data := &packet.Data{}
	data.SetName(identifier)

	var PrefixList []string
	for _, v := range pit.FindDataMatches(data).Identifier.GetComponents() {
		PrefixList = append(PrefixList, v.ToString())
	}
	fmt.Println(PrefixList)

	fmt.Println(pit.Size())
}

func TestEraseByPITEntry(t *testing.T) {
	pit := CreatePIT()
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	interest := &packet.Interest{}
	interest.SetName(identifier)
	pitEntry := pit.Insert(interest)
	// 删除存在数据
	fmt.Println(pit.Size())
	fmt.Println(pit.EraseByPITEntry(pitEntry))
	fmt.Println(pit.Size())
	// 删除不存在数据
	pit.Insert(interest)
	fmt.Println(pit.Size())
	identifier, err = component.CreateIdentifierByString("/min/pku")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(pit.EraseByPITEntry(&PITEntry{Identifier: identifier}))
	fmt.Println(pit.Size())
}

func TestEraseByLogicFace(t *testing.T) {
	pit := CreatePIT()
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	interest := &packet.Interest{}
	interest.SetName(identifier)
	pitEntry := pit.Insert(interest)
	pitEntry.InRecordList[0] = &InRecord{LogicFace: &lf.LogicFace{LogicFaceId: 0}}

	identifier, err = component.CreateIdentifierByString("/min/pku")
	if err != nil {
		fmt.Println(err)
	}
	interest.SetName(identifier)
	pitEntry = pit.Insert(interest)
	pitEntry.InRecordList[0] = &InRecord{LogicFace: &lf.LogicFace{LogicFaceId: 0}}

	// 删除存在
	fmt.Println(pit.EraseByLogicFace(&lf.LogicFace{LogicFaceId: 0}))
	// 删除不存在
	fmt.Println(pit.EraseByLogicFace(&lf.LogicFace{LogicFaceId: 0}))
}

func BenchmarkInsert(b *testing.B) {
	pit := CreatePIT()
	identifier, err := component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	interest := &packet.Interest{}
	interest.SetName(identifier)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pit.Insert(interest)
	}
}

// 速度基本与表项个数呈线性关系
func BenchmarkSize(b *testing.B) {
	pit := CreatePIT()
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	interest := &packet.Interest{}
	interest.SetName(identifier)
	pit.Insert(interest)

	identifier, err = component.CreateIdentifierByString("/min/pku")
	if err != nil {
		fmt.Println(err)
	}
	interest.SetName(identifier)
	pit.Insert(interest)

	identifier, err = component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	interest.SetName(identifier)
	pit.Insert(interest)

	identifier, err = component.CreateIdentifierByString("/min/pku/cn")
	if err != nil {
		fmt.Println(err)
	}
	interest.SetName(identifier)
	pit.Insert(interest)

	identifier, err = component.CreateIdentifierByString("/min/gdcni15")
	if err != nil {
		fmt.Println(err)
	}
	interest.SetName(identifier)
	pit.Insert(interest)

	identifier, err = component.CreateIdentifierByString("/min/gdcni15/mir")
	if err != nil {
		fmt.Println(err)
	}
	interest.SetName(identifier)
	pit.Insert(interest)

	identifier, err = component.CreateIdentifierByString("/min/gdcni15/filegator")
	if err != nil {
		fmt.Println(err)
	}
	interest.SetName(identifier)
	pit.Insert(interest)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pit.Size()
	}
}

func BenchmarkFind(b *testing.B) {
	pit := CreatePIT()
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	interest := &packet.Interest{}
	interest.SetName(identifier)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pit.Find(interest)
	}
}

func BenchmarkFindDataMatches(b *testing.B) {
	pit := CreatePIT()
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	interest := &packet.Interest{}
	interest.SetName(identifier)
	pit.Insert(interest)

	data := &packet.Data{}
	data.SetName(identifier)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pit.FindDataMatches(data)
	}
}

func BenchmarkEraseByPITEntry(b *testing.B) {
	pit := CreatePIT()
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	interest := &packet.Interest{}
	interest.SetName(identifier)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		pitEntry := pit.Insert(interest)
		b.StartTimer()
		pit.EraseByPITEntry(pitEntry)
	}

}

func BenchmarkEraseByLogicFace(b *testing.B) {
	pit := CreatePIT()
	identifier, err := component.CreateIdentifierByString("/min")
	if err != nil {
		fmt.Println(err)
	}
	interest := &packet.Interest{}
	interest.SetName(identifier)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		pitEntry := pit.Insert(interest)
		pitEntry.InRecordList[0] = &InRecord{LogicFace: &lf.LogicFace{LogicFaceId: 0}}
		b.StartTimer()
		pit.EraseByLogicFace(&lf.LogicFace{LogicFaceId: 0})
	}

}

/*
//插入、查询性能测试
func BenchmarkInsert(b *testing.B) {
	pit := CreatePIT()
	identifierString := "/test"
	for i := 1; i <= 100; i++ {
		for j := 1; j <= 100; j++ {
			identifierString = identifierString + strconv.Itoa(j)
			identifier, err := component.CreateIdentifierByString(identifierString)
			if err != nil {
				fmt.Println(err)
			}
			interest := &packet.Interest{}
			interest.SetName(identifier)
			pit.Insert(interest)
		}
		identifierString = identifierString + "/test"
		identifier, err := component.CreateIdentifierByString(identifierString)
		if err != nil {
			fmt.Println(err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	identifier, err := component.CreateIdentifierByString("/test1/test/test2")
	if err != nil {
		fmt.Println(err)
	}
	interest := &packet.Interest{}
	interest.SetName(identifier)
	for i := 0; i < b.N; i++ {
		pit.Insert(interest)
	}
}
*/
