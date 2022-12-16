package http

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"human/library/log"
	"human/library/net/netutil/breaker"
	xtime "human/library/time"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

const (
	_minRead       = 16 * 1024 // 16kb
	_contentType   = "Content-Type"
	_urlencoded    = "application/json; charset=utf-8"
	_userAgent     = "User-Agent"
	_authorization = "Authorization"
)

var (
	_noKickUserAgent = "469841047@qq.com"
)

func init() {
	n, err := os.Hostname()
	if err == nil {
		_noKickUserAgent = _noKickUserAgent + " " + runtime.Version() + " " + n
	}

}

// ClientConfig is http client conf.
type ClientConfig struct {
	MaxTotal    int
	MaxPerHost  int
	KeepAlive   xtime.Duration
	DialTimeout xtime.Duration
	Timeout     xtime.Duration
	Breaker     *breaker.Config
}

type HttpClient struct {
	GetAddr         string
	UploadStartAddr string
	UploadDoneAddr  string
	RunAddr         string
	ClientConf      *ClientConfig
}

// Client is http client.
type Client struct {
	conf      *HttpClient
	client    *http.Client
	dialer    *net.Dialer
	transport http.Transport
	mutex     sync.RWMutex
	breaker   *breaker.Group
}

// NewClient new a http client pool
func NewClient(c *HttpClient) *Client {
	if c.ClientConf.DialTimeout <= 0 || c.ClientConf.Timeout <= 0 {
		panic("must config http timeout")
	}

	client := new(Client)
	client.breaker = breaker.NewGroup(c.ClientConf.Breaker)
	client.conf = c
	client.dialer = &net.Dialer{
		Timeout:   time.Duration(c.ClientConf.DialTimeout),
		KeepAlive: time.Duration(c.ClientConf.KeepAlive),
	}
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DialContext:         client.dialer.DialContext,
		MaxIdleConns:        c.ClientConf.MaxTotal,
		MaxIdleConnsPerHost: c.ClientConf.MaxPerHost,
		IdleConnTimeout:     time.Duration(c.ClientConf.KeepAlive),
	}
	client.client = &http.Client{
		Transport: transport,
		Timeout:   time.Duration(c.ClientConf.Timeout),
	}

	return client
}

// SetTransport set client transport
func (client *Client) SetTransport(t http.Transport) {
	client.transport = t
}

// SetConfig set client config.
func (client *Client) SetConfig(c *ClientConfig) {
	client.mutex.Lock()
	if c.MaxTotal > 0 {
		client.conf.ClientConf.MaxTotal = c.MaxTotal
	}
	if c.MaxPerHost > 0 {
		client.conf.ClientConf.MaxPerHost = c.MaxPerHost
	}
	if c.Timeout > 0 {
		client.client.Timeout = time.Duration(c.Timeout)
		client.conf.ClientConf.Timeout = c.Timeout
	}
	if c.DialTimeout > 0 {
		client.dialer.Timeout = time.Duration(c.DialTimeout)
		client.conf.ClientConf.DialTimeout = c.DialTimeout
	}
	if c.KeepAlive > 0 {
		client.dialer.KeepAlive = time.Duration(c.KeepAlive)
		client.conf.ClientConf.KeepAlive = c.KeepAlive
	}
	if c.Breaker != nil {
		client.conf.ClientConf.Breaker = c.Breaker
		client.breaker.Reload(c.Breaker)
	}
	client.mutex.Unlock()
}

func (client *Client) DownloadToLocal(url string, filename string) (int64, error) {
	// Get the data
	ret := int64(0)
	resp, err := http.Get(url)
	if err != nil {
		return ret, err
	}
	defer resp.Body.Close()
	// Create output file
	out, err := os.Create(filename)
	if err != nil {
		return ret, err
	}
	defer out.Close()
	// copy stream
	ret, err = io.Copy(out, resp.Body)
	return ret, err
}

func (client *Client) Download(url string, user_id int, filename string) (string, error) {
	//发起网络请求
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	// defer 释放资源
	defer res.Body.Close()
	//定义文件名字
	path := strings.Split(url, "/")
	name := path[len(path)-1]
	namelist := strings.Split(name, "?")
	if len(namelist) > 0 {
		name = namelist[0]
	}

	curtime := time.Now()
	curdate := fmt.Sprintf("%04d%02d%02d%02d%02d%02d", curtime.Year(), curtime.Month(), curtime.Day(), curtime.Hour(), curtime.Minute(), curtime.Second())

	destfile := curdate + "_" + name
	if filename != "" {
		destfile = filename
	}
	//创建文件
	out, err := os.Create(destfile)
	if err != nil {
		return "", err
	}
	// defer延迟调用 关闭文件，释放资源
	defer out.Close()
	//添加缓冲 bufio 是通过缓冲来提高效率。
	wt := bufio.NewWriter(out)
	_, _ = io.Copy(wt, res.Body)
	//将缓存的数据写入到文件中
	_ = wt.Flush()
	return destfile, nil
}

func (client *Client) NewRequest(method, uri string, params url.Values, token string) (req *http.Request, err error) {
	// NewRequest new http request with method, uri, values and headers.
	ru := uri
	if params != nil {
		ru = uri + "?" + params.Encode()
	}
	req, err = http.NewRequest(method, ru, nil)
	if err != nil {
		err = errors.Wrapf(err, "method:%s,uri:%s", method, ru)
		return
	}
	if token != "" {
		token := fmt.Sprintf("Bearer %s", token)
		req.Header.Set(_authorization, token)
	}
	if method == http.MethodPost {
		req.Header.Set(_contentType, _urlencoded)
	}
	//req.Header.Set(_userAgent, _noKickUserAgent)
	return
}

// TimeoutDialer 连接超时和传输超时
func TimeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}

func createReqBody(fileName string, field map[string]string) (string, io.Reader, error) {
	buf := new(bytes.Buffer)
	bw := multipart.NewWriter(buf) // body writer
	defer bw.Close()

	fw, err := bw.CreateFormFile("file", fileName)
	if err != nil {
		return "", nil, err
	}

	f, _ := os.Open(fileName)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()

	_, err = io.Copy(fw, f)
	if err != nil {
		return "", nil, err
	}

	if field != nil {
		for k, v := range field {
			err = bw.WriteField(k, v)
			if err != nil {
				fmt.Printf("WriteField: %s:%s %v\n", k, v, err)
			}
		}
	}

	return bw.FormDataContentType(), buf, nil
}

// UploadFile 上传文件 rURL为第三方接口url,b为文件内容,header为自定义header头
func (client *Client) Upload(method string, token string, id string, filename string, starturl, doneurl, runurl string) (interface{}, error) {
	//获取上传url
	postret, posterr := client.ReqData(method, starturl, nil, token)
	if posterr != nil {
		return nil, posterr
	}

	log.Info("Upload ReqData return :%s\n", postret)
	retMap := make(map[string]interface{}, 0)
	posterr = json.Unmarshal([]byte(postret), &retMap)
	if posterr != nil {
		return nil, posterr
	}

	if retMap["method"] == nil || retMap["url"] == nil {
		return retMap, posterr
	}
	reqMethod := retMap["method"].(string)
	reqUrl := retMap["url"].(string)

	//上传文件nil
	contType, reader, err := createReqBody(filename, nil)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(reqMethod, reqUrl, reader)
	// add headers
	req.Header.Set("Content-Type", contType)
	connectTimeout := 600 * time.Second
	readWriteTimeout := 5184000 * time.Millisecond
	cli := http.Client{
		//忽略证书验证
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
			Dial: TimeoutDialer(connectTimeout, readWriteTimeout),
		},
	}

	resp, err := cli.Do(req)
	if err != nil || resp == nil {
		return nil, err
	}
	defer resp.Body.Close()
	etag := resp.Header["Etag"]
	log.Info("upload Etag=%v\n", etag)
	if etag==nil || len(etag)==0{
		return nil, nil
	}
	
	//调用upload done接口
	doneurl = doneurl + etag[0]
	go client.done_run(method, doneurl, runurl, token)
	return retMap, nil
}

func (client *Client) done_run(method string, doneurl, runurl string, token string) {

	postret, posterr := client.ReqData(method, doneurl, nil, token)
	log.Info("upload done:%s[%v]\n", postret, posterr)
	//调用run接口
	postret, posterr = client.ReqData(method, runurl, nil, token)
	log.Info("upload run:%s[%v]\n", postret, posterr)

}

// UploadFile 上传文件 rURL为第三方接口url,b为文件内容,header为自定义header头
func (client *Client) TranFile(filename string, starturl, runurl, geturl string) (string, string, error) {
	field := make(map[string]string, 0)
	field["key"] = "2f0e5cfdd6d5b"
	field["app"] = "conversion"

	//上传文件
	contType, reader, err := createReqBody(filename, field)
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequest(http.MethodPost, starturl, reader)
	// add headers
	req.Header.Set("Content-Type", contType)
	connectTimeout := 120 * time.Second
	readWriteTimeout := 5184000 * time.Millisecond
	cli := http.Client{
		//忽略证书验证
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
			Dial: TimeoutDialer(connectTimeout, readWriteTimeout),
		},
	}

	resp, err := cli.Do(req)
	if err != nil || resp == nil {
		return "", "", err
	}
	defer resp.Body.Close()

	res, _ := ioutil.ReadAll(resp.Body)
	log.Info("tran start:%s\n", string(res))
	retMap := make(map[string]interface{}, 0)
	err = json.Unmarshal(res, &retMap)
	if err != nil {
		return "", "", err
	}

	log.Info("tran ret map:%s\n", retMap)
	if retMap["id"] == nil {
		return "", "", nil
	}

	id := retMap["id"].(string)
	// 带数据 json 类型
	urlValues := map[string]interface{}{
		"compression": false,
		"format":      "OBJ",
		"id":          id,
	}

	b1, _ := json.Marshal(&urlValues)
	req, err = http.NewRequest(http.MethodPost, runurl, bytes.NewReader(b1))
	req.Header.Set("Content-Type", _urlencoded)
	resp, err = cli.Do(req)
	defer resp.Body.Close()
	postret, _ := ioutil.ReadAll(resp.Body)
	log.Info("tran run:%s\n", postret)
	geturl += "?id=" + id

	//获取下载链接
	down_url := ""
	for true {
		rsp, serr := client.ReqData(http.MethodGet, geturl, nil, "")
		if serr != nil || rsp == "" {
			return "", "", err
		}
		log.Info("tran get:%s\n", rsp)
		retmap := make(map[string]interface{}, 0)
		err = json.Unmarshal([]byte(rsp), &retmap)
		if err != nil {
			return "", "", err
		}

		if _, ok := retmap["state"]; !ok {
			err = errors.New("get dowdload url error")
			return "", "", err
		}

		status := retmap["state"].(string)
		log.Info("tran status:%s\n", status)
		if status != "Expired" {
			time.Sleep(time.Second * 10)
			continue
		}

		down_url = retmap["payload"].(string)
		break
	}

	return down_url, id, nil
}

// Post issues a Post to the specified URL.
func (client *Client) ReqData(method string, uri string, params url.Values, token string) (string, error) {
	req, err := client.NewRequest(method, uri, params, token)
	if err != nil {
		return "", err
	}

	cli := http.Client{}
	resp, err := cli.Do(req)
	if resp == nil {
		fmt.Fprintln(os.Stderr, err)
		return "", err
	}
	defer resp.Body.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return "", err
	}

	res, _ := ioutil.ReadAll(resp.Body)
	return string(res), nil
}

// JSON sends an HTTP request and returns an HTTP json response.
func (client *Client) JSON(c context.Context, req *http.Request, res interface{}, v ...string) (err error) {
	var bs []byte
	if bs, err = client.Raw(c, req, v...); err != nil {
		return
	}
	if res != nil {
		var json = jsoniter.ConfigCompatibleWithStandardLibrary
		if err = json.Unmarshal(bs, res); err != nil {
			err = errors.Wrapf(err, "host:%s, url:%s", req.URL.Host, realURL(req))
		}
	}
	return
}

// Raw sends an HTTP request and returns bytes response
func (client *Client) Raw(c context.Context, req *http.Request, v ...string) (bs []byte, err error) {
	var (
		resp *http.Response
		uri  = fmt.Sprintf("%s://%s%s", req.URL.Scheme, req.Host, req.URL.Path)
	)

	// NOTE fix prom & config uri key.
	if len(v) == 1 {
		uri = v[0]
	}

	err = client.breaker.Do(uri, func() error {
		if resp, err = client.client.Do(req); err != nil {
			return errors.Wrapf(err, "host:%s, url:%s", req.URL.Host, realURL(req))
		}
		defer resp.Body.Close()
		if resp.StatusCode >= http.StatusBadRequest {
			return errors.Errorf("incorrect http status:%d host:%s, url:%s", resp.StatusCode, req.URL.Host, realURL(req))
		}
		if bs, err = readAll(resp.Body, _minRead); err != nil {
			return errors.Wrapf(err, "host:%s, url:%s", req.URL.Host, realURL(req))
		}
		return nil
	})
	return
}

// Do sends an HTTP request and returns an HTTP json response.
func (client *Client) Do(c context.Context, req *http.Request, res interface{}, v ...string) (err error) {
	var bs []byte
	if bs, err = client.Raw(c, req, v...); err != nil {
		return
	}
	if res != nil {
		var json = jsoniter.ConfigCompatibleWithStandardLibrary
		if err = json.Unmarshal(bs, res); err != nil {
			err = errors.Wrapf(err, "host:%s, url:%s", req.URL.Host, realURL(req))
		}
	}
	return
}

// realUrl return url with http://host/params.
func realURL(req *http.Request) string {
	if req.Method == http.MethodGet {
		return req.URL.String()
	} else if req.Method == http.MethodPost {
		ru := req.URL.Path
		if req.Body != nil {
			rd, ok := req.Body.(io.Reader)
			if ok {
				buf := bytes.NewBuffer([]byte{})
				buf.ReadFrom(rd)
				ru = ru + "?" + buf.String()
			}
		}
		return ru
	}
	return req.URL.Path
}

// readAll reads from r until an error or EOF and returns the data it read
// from the basic buffer allocated with a specified capacity.
func readAll(r io.Reader, capacity int64) (b []byte, err error) {
	buf := bytes.NewBuffer(make([]byte, 0, capacity))
	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	_, err = buf.ReadFrom(r)
	return buf.Bytes(), err
}
