package mgmt

import "minlib/mgmt"

func MakeControlResponse(code int, msg, data string) *mgmt.ControlResponse {
	response := &mgmt.ControlResponse{Code: code, Msg: msg, Data: data}
	response.SetString(data)
	return response
}
