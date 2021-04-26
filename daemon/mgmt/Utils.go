package mgmt

import (
	"minlib/mgmt"
	"minlib/minsecurity/crypto/sm3"
	"minlib/minsecurity/crypto/sm4"
)

func MakeControlResponse(code int, msg, data string) *mgmt.ControlResponse {
	response := &mgmt.ControlResponse{Code: code, Msg: msg, Data: data}
	response.SetString(data)
	return response
}

// EncryptStr 使用一个字符串秘钥对一个字符串进行加密
//
// @Description:
// @param key
// @param passwd
// @return string
//
func EncryptStr(key string, str string) ([]byte, error) {
	hashFunc := sm3.New()
	hashFunc.Write([]byte(key))
	passHash := hashFunc.Sum(nil)
	if len(passHash) == 32 {
		for i := 0; i < 16; i++ {
			passHash[i] += passHash[i+16]
		}
	}
	encMsg, err := sm4.Sm4Ecb(passHash[:16], []byte(str), sm4.ENC)
	if err != nil {
		return nil, err
	}
	return encMsg, nil
}

// DecryptStr 使用一个字符串秘钥解密一串密文为字符串
//
// @Description:
// @param key
// @param encMsg
// @return string
//
func DecryptStr(key string, encMsg []byte) (string, error) {
	hashFunc := sm3.New()
	hashFunc.Write([]byte(key))
	passHash := hashFunc.Sum(nil)
	if len(passHash) == 32 {
		for i := 0; i < 16; i++ {
			passHash[i] += passHash[i+16]
		}
	}
	dec, err := sm4.Sm4Ecb(passHash[:16], encMsg, sm4.DEC)
	if err != nil {
		return "", err
	}
	return string(dec), nil
}
