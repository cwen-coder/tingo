package tingo

import (
	//"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"tingo/com"
)

type Tingo struct {
	indexUrl         string
	agreement        string
	host             string
	agreementAndHost string
	targetPath       string
	exceptionUrl     map[string]bool
	goroutineCount   int            // 记录goroutine的数量
	wg               sync.WaitGroup //WaitGroup用于等待一组线程的结束。父线程调用Add方法来设定应等待的线程的数量。每个被等待的线程在结束时应调用Done方法。同时，主线程里可以调用Wait方法阻塞至所有线程结束。
	lock             sync.Mutex
	ch               chan bool
}

func NewTingo() *Tingo {
	tin := &Tingo{
		goroutineNum: 0,
		lock:         &sync.Mutex{},
	}
	return tin
}

func (this *Tingo) Fetch(url, targetPath string) {
	url = strings.TrimSpace(url) //去两边的空格
	//格式化url
	this.parseUrl(url)
	//保存路径
	this.handleTargetPath(targetPath)
	//去掉"http://"或是"https://"
	url = com.Substring(url, len(this.agreement))

	this.handleUrl(url, false)
	//Wait方法阻塞直到WaitGroup计数器减为0
	this.wg.Wait()

	// 处理异常
	this.handleExceptionUrl()

}

func (this *Tingo) handleUrl(url string, needException bool) {
	this.wg.Add(1)

	this.ch <- true //使用资源

	this.lock.Lock() //上锁
	this.goroutineCount++
	log.Println("当前共运行", this.goroutineNum, "goroutine")
	this.lock.Unlock() //解锁

	go func() {
		defer func() {
			this.wg.Done()
		}()

		children := this.handle(url, needException)

		this.lock.Lock()
		this.goroutineNum--
		log.Println("当前共运行", this.goroutineNum, " goroutine")
		this.lock.Unlock()

		<-this.ch // 释放资源

		for _, childUrl := range children {
			this.handleUrl(childUrl, false)
		}
	}()

}

//格式化url
func (this *Tingo) parseUrl(url string) {
	if strings.HasPrefix(url, "http://") {
		this.agreement = "http://"
	} else {
		this.agreement = "https://"
	}

	url = strings.Replace(url, this.agreement, "", 1)
	index := strings.Index(url, "/")
	if index == -1 {
		this.host = url
	} else {
		this.host = com.Substr(url, 0, index)
	}
	this.agreementAndHost = this.agreement + this.host
}

//保存路径
func (this *Tingo) handleTargetPath(path string) {
	path = strings.TrimRight(path, "/")
	path = strings.Trim(path, "\\")
	if path != "" {
		this.targetPath = path
	}

	if this.targetPath != "" {
		os.MkdirAll(this.targetPath, 0777)
	} else {
		panic("输入的存储位置有问题")
	}
}

func (this *Tingo) trimUrl(url string) string {
	if url != "" {
		url = strings.TrimSpace(url)
		url = strings.Trim(url, "\"")
		url = strings.Trim(url, "'")
		url = strings.Trim(url, "/")
		url = strings.Trim(url, "\\")
	}

	return url
}

//异常处理
func handleExceptionUrl() {
	if len(this.exceptionUrl) > 0 {
		log.Println("正在处理异常Url....")
		for url, _ := range this.exceptionUrl {
			this.do(url, false)
		}
	}
}
