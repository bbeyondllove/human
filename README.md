# golang + gin
这是一个基于[golang](https://golang.org/) + [gin](https://gin-gonic.com/)的基础Web框架。项目是完全按照Gopher
公认的[项目标准结构](https://github.com/golang-standards/project-layout)定制

包含如下特性：
* 基于[gin](https://gin-gonic.com/)的轻量级web框架，拥有更加出色的性能。
    * gin-validator表单验证。底层实现:https://gopkg.in/go-playground/validator.v8
    * middleware拦截功能：可以自由的实现一些拦截handler
    * swagger API文档
    * pprof：性能分析工具
    * 优雅退出机制

* 基于[toml](https://github.com/toml-lang/toml)的配置文件，真的是谁用谁知道。
* 参考[Kratos](https://github.com/bilibili/kratos)实现的log组件。支持配置日志位置、按文件大小滚动日志、缓冲大小等。   
* 参考[Kratos](https://github.com/bilibili/kratos)实现的mysql组件。读写分离、支持事务、慢查询日志记录等。
* 基于[redigo](https://github.com/gomodule/redigo)封装的redis组件。这也是golang官方推荐的redis包。
* 基于net/http封装的http连接池组件。
* 经典的错误码设计理念，与golang的error处理机制完美结合。
* and so on...

# 项目结构


# 更新 2022-10-26
* 新加基于 [hystrix-go](https://github.com/afex/hystrix-go) 的熔断器。
* 优化 http Client 组件，并集成 hystrix，配置 demo 如下
```toml
[httpClient]
    [httpClient.abc]
        addr = "http://api.abc.com"
        [httpClient.abc.clientConf]
            maxTotal = 10
            maxPerHost  = 10
            keepAlive = "5s"
            dialTimeout = "1s"
            timeout = "1s"
            [httpClient.abc.clientConf.breaker]
                namespace = "abc"
                timeout = "3s"
                maxConcurrentRequests = 5
                requestVolumeThreshold= 1
                sleepWindow = "5s"
                errorPercentThreshold = 50
```
* 新加 gRPC Client 组件，并集成 hystrix，配置 demo 如下
```toml
[grpcClient]
    [grpcClient.sayHello]
        addr = "0.0.0.0:9101"
        [grpcClient.sayHello.clientConf]
            dialTimeout = "1s"
            timeout = "1s"
            poolSize = 4
            [grpcClient.sayHello.clientConf.breaker]
                namespace = "sayHello"
                timeout = "1s"
                maxConcurrentRequests = 1000
                requestVolumeThreshold= 10
                sleepWindow = "5s"
                errorPercentThreshold = 60
```
* 完善相关组件单元测试

# 使用说明

## 安装编译
```$xslt 
$ go get 本项目

# go get代理，org包要科学上网
$ export GOPROXY=https://goproxy.io

# 开启GO111MODULE模式
$ export GO111MODULE=on

# 使用GO MODULE模式管理包
$ go mod init

# 编译
$ go build

# 运行
$ ./human -conf ../configs/application.toml 
或者指定端口
$ ./human -conf ../configs/application.toml --http.port=80
```

## Swagger

### 安装swagger
```$xslt
$ go get -u github.com/swaggo/swag/cmd/swag
```
若 $GOPATH/bin 没有加入$PATH中，你需要执行将其可执行文件移动到$GOBIN下
```$xslt
$ mv $GOPATH/bin/swag /usr/local/go/bin
```

### 验证是否安装成功
```$xslt
$ swag -v
swag version vxxx
```

### 编写API注释
```$xslt
// @Summary 注册
// @Produce json
// @Param username query string true "用户名"
// @Param password query string true "密码"
// @Param config_password query string true "确认密码"
// @Success 200 {object} render.JSON
// @Router /api/register [post]
```

### 生成
我们进入到的根目录，执行初始化命令
```$xslt
$ swag init
```

### 验证
大功告成，访问：http://localhost/swagger/index.html

## pprof性能分析工具

### 安装pprof
请参考：https://github.com/DeanThompson/ginpprof

### 验证
大功告成，访问：http://localhost/debug/pprof/

### pprof实战
请参考：https://blog.wolfogre.com/posts/go-ppof-practice/

## 代码片段

### 基于toml的配置文件
用起来不要太爽，谁用谁知道...

### Router配置
基于gin的Router高效、强大、简单


### Middleware拦截器
基于gin的middleware，可以实现日志打印、权限验证等等..

### 表单Validator功能
很实用的表单validation功能，文档：https://gopkg.in/go-playground/validator.v8

### 优雅退出

### 经典优雅的错误码设计
定义Codes，并直接errors.Cause(e).(Codes)进行强转判断，完美兼容golang的error显性处理机制


### DB事务操作

### Redis操作

### Http连接池请求连接复用

# 后续规划
* [x] DB、Redis、Http、gRPC组件
* [x] 熔断器
* [ ] Router限速器功能

# 参考项目
* https://github.com/gin-gonic/gin
* https://github.com/EDDYCJY/go-gin-example
* https://github.com/bilibili/kratos
