// Copyright [2022] [MIN-Group -- Peking University Shenzhen Graduate School Multi-Identifier Network Development Group]
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Package mgmt
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/12/21 4:32 PM
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//

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
