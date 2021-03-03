package main

import (
	"fmt"
	"mir/daemon/fw"
)

func main() {
	inte := 2
	bo := true
	fw.OnIncomingInterest()
	fmt.Println(inte, bo)
}
