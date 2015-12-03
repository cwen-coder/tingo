package tingoOne

import (
	"os"
	"testing"
)

func Test_Fetch(t *testing.T) {
	tmp := NewTingo()
	tmp.Fetch("http://imooc.com", "/home/yin_cwen/tingoTest")
	if tmp.agreementAndHost != "http://imooc.com" {
		t.Error("Fetch　函数 url(parseUrl) 处理失败")
	} else {
		t.Log("Fetch 函数 url(parseUrl) 处理成功")
	}

	if _, err := os.Stat("/home/yin_cwen/tingoTest"); err != nil {
		t.Error("Fetch　函数 targetPath(handleTargetPath) 处理失败")
	} else {
		t.Log("Fetch　函数 targetPath(handleTargetPath) 处理成功")
	}

}
