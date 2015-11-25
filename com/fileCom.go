package com

import "os"

func IsExists(path string) bool {
	_, err := os.Stat(path)
	//返回一个布尔值说明该错误是否表示一个文件或目录不存在。ErrNotExist和一些系统调用错误会使它返回真。
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
