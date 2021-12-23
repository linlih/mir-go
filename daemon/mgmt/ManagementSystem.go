package mgmt

import (
	"mir-go/daemon/lf"
	"mir-go/daemon/table"
)

type ManagementSystem struct {
	csManager       *CsManager
	fibManager      *FibManager
	faceManager     *FaceManager
	identityManager *IdentityManager
}

func (m *ManagementSystem) Init(dispatcher *Dispatcher, logicFaceTable *lf.LogicFaceTable) {
	m.fibManager.Init(dispatcher, logicFaceTable)
	m.faceManager.Init(dispatcher, logicFaceTable)
	m.csManager.Init(dispatcher, logicFaceTable)
	m.identityManager = CreateIdentityManager(dispatcher.keyChain)
	m.identityManager.Init(dispatcher)
}

func (m *ManagementSystem) SetFIB(fib *table.FIB) {
	m.fibManager.fib = fib
}

func (m *ManagementSystem) BindFibCleaner(l *lf.LogicFaceTable) {
	l.OnEvicted = m.fibManager.NextHopCleaner
}

func CreateMgmtSystem() *ManagementSystem {
	return &ManagementSystem{
		csManager:   CreateCsManager(),
		faceManager: CreateFaceManager(),
		fibManager:  CreateFibManager(),
	}
}
