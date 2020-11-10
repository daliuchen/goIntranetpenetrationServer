package netWorkPa

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func mai1n() {
	go startServer()
	time.Sleep(time.Second * 1)
	go startClient()
	time.Sleep(time.Hour * 1)
}

//开启客户端
func startClient() {
	log.Print("start client ....")
	port := ":7777"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", port)
	checkErr(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	//开启一个线程，输入读取到con，此线程要一直获取输入的值
	go io.Copy(conn, os.Stdin)
	//接受con的输出
	io.Copy(os.Stdout, conn)
}

//开启服务端
func startServer() {
	log.Print("start server ....")
	port := ":7777"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", port)
	checkErr(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	for {
		con, err := listener.Accept()
		checkErr(err)
		log.Println("客户端连接成功:", con.RemoteAddr().String())
		go handleConnection(con)
	}
}
func handleConnection(conn net.Conn) {
	//将 con 的数据读取到 buffer的数据，
	// 没有数据就一直堵塞。
	input := bufio.NewScanner(conn)
	// 不断读取 buffer的数据
	for input.Scan() {
		//将客户端的输入大写返回
		fmt.Fprintln(conn, "\t", strings.ToUpper(input.Text()))
		//修改1秒
		time.Sleep(time.Second * 1)
		// 原封不动的写出
		fmt.Fprintln(conn, "\t", input.Text())
		time.Sleep(time.Second * 1)
		//将客户端的输入小写返回
		fmt.Fprintln(conn, "\t", strings.ToLower(input.Text()))
	}
}
func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
