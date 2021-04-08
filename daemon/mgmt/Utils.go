package mgmt

import "minlib/mgmt"

func MakeControlResponse(code int, msg, data string) *mgmt.ControlResponse {
	return &mgmt.ControlResponse{Code: code, Msg: msg, Data: data}
}
