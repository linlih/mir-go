// Package utils
// @Author: Jianming Que
// @Description:
// @Version: 1.0.0
// @Date: 2021/4/26 11:27 上午
// @Copyright: MIN-Group；国家重大科技基础设施——未来网络北大实验室；深圳市信息论与未来网络重点实验室
//
package utils

import (
	"fmt"
	"io/ioutil"
	"minlib/common"
	"os"
	"strings"
)

func GetRelPath(path string) string {
	if strings.HasPrefix(path, "~") {
		homePath, err := common.Home()
		if err != nil {
			common.LogFatal("Get current user home path failed!")
		}
		return homePath + path[1:]
	} else {
		return path
	}
}

func ReadFromFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	if content, err := ioutil.ReadAll(file); err != nil {
		return "", err
	} else {
		return string(content), nil
	}
}

// WriteFile 写入文件,文件不存在则创建,如在则追加内容
func WriteFile(path string, str string) {
	_, b := IsFile(path)
	var f *os.File
	var err error
	if b {
		//打开文件，
		f, _ = os.OpenFile(path, os.O_APPEND, 0666)
	} else {
		//新建文件
		f, err = os.Create(path)
	}

	//使用完毕，需要关闭文件
	defer func() {
		err = f.Close()
		if err != nil {
			fmt.Println("err = ", err)
		}
	}()

	if err != nil {
		fmt.Println("err = ", err)
		return
	}
	_, err = f.WriteString(str)
	if err != nil {
		fmt.Println("err = ", err)
	}
}

// IsExists 判断路径是否存在
func IsExists(path string) (os.FileInfo, bool) {
	f, err := os.Stat(path)
	return f, err == nil || os.IsExist(err)
}

// IsDir 判断所给路径是否为文件夹
func IsDir(path string) (os.FileInfo, bool) {
	f, flag := IsExists(path)
	return f, flag && f.IsDir()
}

// IsFile 判断所给路径是否为文件
func IsFile(path string) (os.FileInfo, bool) {
	f, flag := IsExists(path)
	return f, flag && !f.IsDir()
}
