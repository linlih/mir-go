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

// Package common
// @Author: Guohua Wei
// @Description:
// @Version: 1.0.0
// @Date: 2021/3/31 上午11:15
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//

package common

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
)

//
// @Author: Wei Guohua
// @Description:
// @Version: 1.0.0
// @Date: 2021/6/9 下午3:45
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//

type DefaultRouteConfig struct {
	Link []Link
}

type Link struct {
	RemoteUri   string
	LocalUri    string
	Persistence int
	Routes      Routes
}

type Routes struct {
	Route []Route
}

type Route struct {
	Identifier  string
	Cost        int
	Persistence int
}

func ParseDefaultConfig(configPath string) (*DefaultRouteConfig, error) {
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var defaultConfig DefaultRouteConfig
	err = xml.Unmarshal(content, &defaultConfig)
	if err != nil {
		return nil, err
	}
	fmt.Println(defaultConfig, "==================")
	return &defaultConfig, nil
}
