package utils

import (
	"math/rand"
	"minlib/component"
	"minlib/packet"
	"sync"
	"testing"
)

var w sync.WaitGroup

//func TestCreateBlockQueue(t *testing.T) {
//	que := CreateBlockQueue(100)
//	w.Add(1)
//
//	count:=0
//	wr:=0
//	identifier, _ := component.CreateIdentifierByString("/min")
//	interest := new(packet.Interest)
//	interest.SetTtl(3)
//	go func() {
//		defer func(){
//			w.Done()
//		}()
//		for{
//			get:=que.Read()
//			getinterest:=get.(*packet.Interest)
//			fmt.Println("get interest",getinterest.Payload)
//			count++
//			fmt.Println("count",count)
//			if count==1000{
//				break
//			}
//		}
//
//	}()
//	interest.SetName(identifier)
//	for i:=0;i<1000;i++{
//		fmt.Println("yb ")
//		token := make([]byte, 7000)
//		rand.Read(token)
//		interest.Payload.SetValue(token)
//		que.Write(interest)
//		time.Sleep(2*time.Millisecond)
//		wr++
//		fmt.Println("wr",wr)
//	}
//
//	w.Wait()
//
//}

func TestCreateBlockQueue(t *testing.T) {
	que := CreateBlockQueue(10)
	for i := 0; i < 20; i++ {
		que.Write(i)
	}
}

func BenchmarkCreateBlockQueue(b *testing.B) {
	que := CreateBlockQueue(100)
	w.Add(1)

	count := 0
	wr := 0
	identifier, _ := component.CreateIdentifierByString("/min")
	interest := new(packet.Interest)
	interest.SetTtl(3)
	b.ResetTimer()
	go func() {
		defer func() {
			w.Done()
		}()
		for {
			//get:=que.Read()
			que.Read()
			//getinterest:=get.(*packet.Interest)
			//fmt.Println("get interest",getinterest.Payload)
			count++
			//fmt.Println("count",count)
			if count == 1000000 {
				break
			}
		}

	}()
	interest.SetName(identifier)
	//defer func(){
	//	w.Done()
	//}()
	for i := 0; i < 1000; i++ {

		//fmt.Println("yb ")
		token := make([]byte, 7000)
		rand.Read(token)
		interest.Payload.SetValue(token)
		go func() {
			for j := 0; j < 1000; j++ {
				que.Write(interest)
				wr++
				//fmt.Println("wr", wr)
			}
		}()

		//time.Sleep(2*time.Millisecond)
		//wr++
		//fmt.Println("wr",wr)
	}

	w.Wait()
}
