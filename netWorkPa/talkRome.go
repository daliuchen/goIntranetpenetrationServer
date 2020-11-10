package netWorkPa

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

//定义一个 存放消息的chan
type cChan chan<- string

var (
	//有客户端连接的chan
	entering = make(chan cChan)
	//有客户端离开的chan
	leaving = make(chan cChan)
	//消息通信的通道
	message = make(chan string)
)

func main() {
	//
	port := ":7777"
	addr, err := net.ResolveTCPAddr("tcp4", port)
	checkErr(err)
	listen, err := net.ListenTCP("tcp", addr)
	checkErr(err)
	for {
		connection, err := listen.Accept()
		checkErr(err)
		log.Println("客户端连接成功:客户端地址为:", connection.RemoteAddr().String())
		//处理connecton
		go handleConnection(connection)
	}

}
func handleConnection(conn net.Conn) {
	//客户端ip
	clientIp := conn.RemoteAddr().String()
	clientChan := make(chan string)
	//开启线程给客户端响应
	go responseClient(conn, clientChan)

	msg := "you ip " + clientIp
	clientChan <- msg
	//加入到队列里面
	message <- clientIp + "entering"
	entering <- clientChan
	//读取消息
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := scanner.Text()
		//发送消息
		message <- clientIp + ":" + text
	}
	//上面会堵塞，程序如果到这里来就说明客户端断开了连接
	leaving <- clientChan
	message <- clientIp + ": leaving"
	conn.Close()
}

//给客户端响应
func responseClient(conn net.Conn, c <-chan string) {
	for msg := range c {
		fmt.Fprintf(conn, "%s \n", msg)
	}
}
