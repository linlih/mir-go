package main

import "fmt"

type A struct {
	v int
}

func (a *A) Add(b *A) int {
	return a.v + b.v
}

type Callback func(b *A) int

func Test(c Callback) {
	b := A{
		2,
	}
	fmt.Println(c(&b))
}

func main() {
	a := A{
		1,
	}
	Test(a.Add)
}
