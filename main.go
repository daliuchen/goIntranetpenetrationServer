package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

/**
服务端：
主要功能
- 接受http请求
- 和客户端webSocket
- 将http请求传给客户端，
- 将客户端的响应返回给http请求
*/
func main() {
	//启动webSocket
	go startSocket()
	go startHttpService()
	time.Sleep(time.Hour * 2)
}

const (
	applicationJson string = "application/json"
)

/**
socket连接  客户端返回给服务端对象
*/
type ResponseBody struct {
	/**
	请求头
	*/
	MyHeader map[string][]string `json:"myHeader"`
	/**
	请求体
	*/
	MyBody string `json:"myBody"`
}

//解析客户端返回给服务端数据，解析为 ResponseBody 对象
func UnTransformJson(str string) *ResponseBody {
	//发送请求获取返回值返回
	var s ResponseBody
	//使用go自带的json解析
	json.Unmarshal([]byte(str), &s)
	return &s
}

/*
定义全部变量，socket线程和http线程通信的 通道
*/
var (
	//http 线程 发送消息给 socket 线程的通道
	contextsChan chan UrlContext
	//socket线程 发送消息给 http线程的通道
	responseParam chan string
)

/**
socket连接，服务端发给客户端的包装对象
*/
type UrlContext struct {
	Method           string                 `json:"method"`
	RequestUrl       string                 `json:"requestUrl"`
	RequestParam     map[string][]string    `json:"requestParam"`
	RequestHeader    map[string][]string    `json:"requestHeader"`
	JsonRequestParam map[string]interface{} `json:"jsonRequestParam"`
	ApplicationType  string                 `json:"applicationType"`
}

/**
自定义http转发的统一处理
*/
type MyMux struct {
}

/*
自定义handle处理类
*/
func (p *MyMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestHeadType := r.Header.Get("Content-Type")
	/**
	构建 UrlContext 对象
	*/
	urlContext := UrlContext{
		Method:        r.Method,
		RequestUrl:    r.URL.Path,
		RequestHeader: r.Header,
	}
	switch requestHeadType {
	case applicationJson:
		b, _ := ioutil.ReadAll(r.Body)
		m := make(map[string]interface{})
		json.Unmarshal(b, &m) //第二个参数要地址传递
		urlContext.JsonRequestParam = m
		urlContext.ApplicationType = applicationJson
	default:
		//通过socket 给客户端
		r.ParseMultipartForm(32 << 20)
		urlContext.RequestParam = r.Form
	}

	//赛到通道里面，socket线程就 监听到有数据了，就开始给客户端发送消息了。
	contextsChan <- urlContext
	log.Println("消息转发给socket线程，数据为:", urlContext)
	/*
		等待socket线程 返回给http的消息。
	*/
	for s := range responseParam {
		log.Println("http线程监听到 socket 线程的数据了。")
		//解析为 ResponseBody 因为这个是客户端发送给服务端的消息
		transformJson := UnTransformJson(s)
		//设置 请求头和请求体，并且返回数据
		setResponseHeaderAndBody(transformJson, w)
		return
	}
}

//组装响应的response返回
func setResponseHeaderAndBody(transformJson *ResponseBody, w http.ResponseWriter) {
	//设置responseHeader
	for k, v := range transformJson.MyHeader {
		for _, i2 := range v {
			w.Header().Set(k, i2)
		}
	}
	//设置body 并且返回
	w.Write([]byte(transformJson.MyBody))
}

//开启http服务 等待调用者
func startHttpService() {
	log.Println("开启http服务,监听端口为 9090")
	mux := &MyMux{}
	err := http.ListenAndServe(":9090", mux)
	checkErr(err)
}

//开启socket
func startSocket() {
	log.Println("socket线程开始启动,socket 端口为 7777")
	port := ":7777"
	addr, err := net.ResolveTCPAddr("tcp4", port)
	checkErr(err)
	listen, err := net.ListenTCP("tcp", addr)
	for {
		log.Println("socket线程 ------->启动成功，监听端口为:", port)
		clientCon, _ := listen.Accept()
		log.Println("socket线程 ------->客户端连接成功,客户端地址为:", clientCon.RemoteAddr().String())
		//初始化两个通道
		contextsChan = make(chan UrlContext)
		responseParam = make(chan string)
		go handleConnection(clientCon)
	}
}

//处理socket线程的客户端连接
func handleConnection(conn net.Conn) {
	for {
		//接受http线程给的消息
		context := <-contextsChan
		//使用go自带的 json格式化
		b, _ := json.Marshal(context)
		//发送给socket 线程的客户端
		fmt.Fprintln(conn, string(b))
		// 该线程堵塞，一直等待 socket 线程的 消息
		scanner := bufio.NewScanner(conn)
		//如果有消息
		if scanner.Scan() {
			text := scanner.Text()
			log.Println("socket线程，客户端返回的消息为:", text)
			//socket线程发送消息给 http线程
			responseParam <- text
			log.Println("socket线程发送消息给 http线程 成功")
		}
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
