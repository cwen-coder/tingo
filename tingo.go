package tingo

import (
	//"fmt"
	"http"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"tingo/com"
)

type Tingo struct {
	indexUrl           string
	agreement          string
	host               string
	agreementAndHost   string
	targetPath         string
	defaultFilename    string
	exceptionUrl       map[string]bool
	hadHandleUrl       map[string]bool
	noChildrenFileExts []string
	goroutineCount     int            // 记录goroutine的数量
	wg                 sync.WaitGroup //WaitGroup用于等待一组线程的结束。父线程调用Add方法来设定应等待的线程的数量。每个被等待的线程在结束时应调用Done方法。同时，主线程里可以调用Wait方法阻塞至所有线程结束。
	lock               sync.Mutex
	ch                 chan bool
}

func NewTingo() *Tingo {
	tin := &Tingo{
		goroutineNum:       0,
		defaultFilename:    "index.html",
		lock:               &sync.Mutex{},
		noChildrenFileExts: []string{".js", ".ico", ".png", ".jpg", ".gif"},
	}
	tin.ch = make(chan bool, 1000) // 仅limit个goroutine
	tin.hadHandleUrl = make(map[string]bool, 1000)
	tin.exceptionUrl = make(map[string]bool, 1000)
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

func handle(url string, needException bool) (children []string) {
	children = nil
	url = this.trimUrl(url)
	if this.isNotNeedUrl(url, needException) {
		return
	}

	// 文件是否已存在
	// url = a.com/a/?id=12&id=1221, 那么genUrl=a.com/a/index.html?id=121
	genUrl := this.genUrl(url)
	if this.isExists(genUrl) {
		return
	}

	// 得到内容
	fullUrl := this.agreement + url
	if needException {
		log.Println("正在处理 `异常` " + fullUrl)
	} else {
		log.Println("正在处理 " + fullUrl)
	}

	content, err := this.getContent(fullUrl)
	if !needException && (err != nil || content == "") { // !needException防止处理异常时无限循环
		this.exceptionUrl[url] = true
		return
	}

	this.hadHandleUrl[url] = true

	ext := strings.ToLower(filename.Ext(this.trimQueryParams(url)))
	if ext == ".css" {
		children = this.doCSS(url, content)
		return
	}

	// 如果是js, image文件就不往下执行了
	if util.InArray(this.noChildrenFileExts, ext) {
		// 保存该文件
		if !this.writeFile(url, content) {
			return
		}
		return
	}

}

// 处理css
func (this *Tingo) doCSS(url, content string) (children []string) {
	children = nil
	// 保存该文件
	if !this.writeFile(url, content) {
		return
	}

	regular := "(?i)url\\((.+?)\\)"
	reg := regexp.MustCompile(regular)
	re := reg.FindAllStringSubmatch(content, -1)

	log.Println(url + " 含有: ")
	log.Println(re)
	baseDir := filepath.Dir(url)

	for _, each := range re {
		cUrl := this.trimUrl(each[1])
		// 这里, goDo会申请资源, 导致doCSS一直不能释放资源
		children = append(children, this.cleanUrl(baseDir+"/"+cUrl))
	}

	return
}

//获取内容
func (this *Tingo) getContent(url string) (content string) {
	var resp *http.Response
	resp, err = http.Get(url)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	} else {
		log.Println("ERROR " + url + " 返回为空 ")
	}

	if resp == nil || resp.Body == nil || err != nil || resp.StatusCode != http.StatusOK {
		log.Println("ERROR " + url)
		log.Println(err)
		return
	}

	var buf []byte
	buf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	content = string(buf)
	return
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

// 不需要处理的url
// needException false 表示不要处理, 那么就要判断是否在其中
func (this *Tingo) isNotNeedUrl(url string, needException bool) bool {
	if _, ok := this.hadHandleUrl[url]; ok {
		return true
	}

	if _, ok := this.exceptionUrl[url]; !needException && ok {
		return true
	}

	// http:\\/|https:\\/|
	regular := "#|javascript:|mailto:|&quot; class=|@.*?\\..+"
	reg := regexp.MustCompile(regular)
	if reg.MatchString(url) {
		return true
	}

	if (strings.HasPrefix(url, "http:/") || strings.HasPrefix(url, "https:/")) &&
		!strings.HasPrefix(url, this.scheme+this.host) {
		return true
	}

	return false
}

//htt
func (this *Tingo) genUrl(url string) string {
	queryParam, fragment := "", "" // 包含?,#
	paramIndex := strings.Index(url, "?")
	if paramIndex != -1 {
		queryParam = com.Substring(url, paramIndex) //"?"后边的参数
		url = com.Substr(url, 0, paramIndex)
	} else {
		paramIndex = strings.Index(url, "#")
		if paramIndex != -1 {
			fragment = com.Substring(url, paramIndex) //"#"后边的参数
			url = com.Substr(url, 0, paramIndex)
		}
	}
	// 如果url == host
	if url == this.host || url == this.schemeAndHost {
		return url + "/" + this.defaultFilename + queryParam + fragment
	}

	genFilename, needApend := this.genFilename(url)
	if genFilename != "" {
		if needApend {
			url += "/" + genFilename + queryParam + fragment
		} else {
			// 是a.php => a.html
			urlArr := strings.Split(url, "/")
			urlArr = urlArr[:len(urlArr)-1]
			url = strings.Join(urlArr, "/") + "/" + genFilename
		}
	}

	return url
}

func (this *Tingo) genFilename(url string) (string bool) {
	urlArr := strings.Split(url, "/")
	if urlArr != nil {
		last := urlArr[len(urlArr)-1]
		ext := strings.ToLower(filepath.Ext(url)) //获取后缀
		if ext == "" {
			return this.defaultFilename, true // 需要append到url后面
		} else if com.InArray([]string{".php", ".jsp", ".asp", ".aspx"}, ext) {
			filename := filepath.Base(last)                             // a.php
			filename = util.Substr(filename, 0, len(filename)-len(ext)) // a
			return filename + ".html", false
		}
	}
	return "", true
}

// 判断是否已存在
// url = a/b/c/d.html
func (this *Tingo) isExists(url string) bool {
	return util.IsExists(this.targetPath + "/" + url)
}

func (this *Tingo) trimQueryParams(url string) string {
	pos := strings.Index(url, "?")
	if pos != -1 {
		url = com.Substr(url, 0, pos)
	}

	pos = strings.Index(url, "#")
	if pos != -1 {
		url = com.Substr(url, 0, pos)
	}

	return url
}

func (this *Tingo) cleanUrl(url string) string {
	url = filepath.Clean(url)
	return strings.Replace(url, "\\", "/", -1)
}

func (this *Tingo) writeFile(url, content string) bool {
	// $path = a.html?a=a11
	url = this.trimQueryParams(url)

	fullPath := this.targetPath + "/" + url
	dir := filepath.Dir(fullPath)
	log.Println("写目录", dir)
	if err := os.MkdirAll(dir, 0777); err != nil {
		log.Println("写目录" + dir + " 失败")
		return false
	}

	// 写到文件中
	file, err := os.Create(fullPath)
	defer file.Close()
	if err != nil {
		log.Println("写文件" + fullPath + " 失败")
		return false
	}
	file.WriteString(content)
	return true
}
