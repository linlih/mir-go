package mgmt

import (
	"minlib/component"
	"minlib/mgmt"
	"minlib/packet"
)

func Auth (topPrefix *component.Identifier, interest *packet.Interest, parameters *mgmt.ControlParameters, accept AuthorizationAccept, reject AuthorizationReject){

}

func ValidatePara(parameters *mgmt.ControlParameters)bool{
	return true
}


var dis Dispacher

