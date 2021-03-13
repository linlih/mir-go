package table

import (
	"fmt"
	"minlib/component"
	"minlib/packet"
	"testing"
)

func TestCSSize(t *testing.T) {
	cs := CreateCS()
	data := &packet.Data{}
	identifier, err := component.CreateIdentifierByString("/min/pku/edu")
	if err != nil {
		fmt.Println(err)
	}
	data.SetName(identifier)
	cs.Insert(data)
	fmt.Println(cs.Size())

	identifier, err = component.CreateIdentifierByString("/min/pku")
	if err != nil {
		fmt.Println(err)
	}
	data.SetName(identifier)
	cs.Insert(data)
	fmt.Println(cs.Size())

	//异常空值 直接报错
	//data=&packet.Data{}
	//cs.Insert(data)
	//fmt.Println(cs.Size())
}
