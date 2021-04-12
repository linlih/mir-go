package mgmt

import (
	"minlib/component"
	"mir-go/daemon/lf"
	"mir-go/daemon/table"
)

type ManagementSystem struct {
	csManager   *CsManager
	fibManager  *FibManager
	faceManager *FaceManager
}

func (m *ManagementSystem) Init(dispatcher *Dispatcher, logicFaceTable *lf.LogicFaceTable) {
	m.fibManager.Init(dispatcher, logicFaceTable)
	m.faceManager.Init(dispatcher, logicFaceTable)
	m.csManager.Init(dispatcher, logicFaceTable)
}

func (m *ManagementSystem) SetFIB(fib *table.FIB) {
	m.fibManager.fib = fib
}

func (m *ManagementSystem) AddInnerFace(identifier *component.Identifier, logicFace *lf.LogicFace, cost uint64) {
	m.fibManager.fib.AddOrUpdate(identifier, logicFace, cost)
}

func CreateMgmtSystem() *ManagementSystem {
	return &ManagementSystem{
		csManager:   CreateCsManager(),
		faceManager: CreateFaceManager(),
		fibManager:  CreateFibManager(),
	}
}
