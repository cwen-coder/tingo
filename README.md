# tingo     
golang 实现抓取网页模板   
#### ubuntu14.04 用法
```
git clone https://github.com/cwen-coder/tingo   
cd $GOPATH/src/
mkdir test   
touch test.go  
```  

#### test.go  
``` 
package main 

import (
    "github.com/cwen-coder/tingo" 
)

func main() {
    tmp := tingo.NewTingo()
    tmp.Fetch("http://cwengo.com", "/home/yin_cwen/tingoTest")  //url  path
}
```  

