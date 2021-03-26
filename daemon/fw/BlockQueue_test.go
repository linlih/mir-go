package fw

import (
	"math/rand"
	"minlib/component"
	"minlib/packet"
	"testing"
)

func TestCreateBlockQueue(t *testing.T) {
	que := CreateBlockQueue(10)
	token := make([]byte, 7000)
	rand.Read(token)
	identifier, _ := component.CreateIdentifierByString("/min")
	interest := new(packet.Interest)
	interest.SetTtl(3)
	interest.Payload.SetValue(token)
	interest.SetName(identifier)
	go que.write(interest)
	que.read()
}
