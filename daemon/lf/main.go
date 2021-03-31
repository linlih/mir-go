package lf

func main() {
	var lfTb LogicFaceTable
	lfTb.Init()

	var lfsys LogicFaceSystem
	lfsys.Init(&lfTb)
	lfsys.Start()


}
