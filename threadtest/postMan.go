package threadtest

import (
	"log"
	"strconv"
	"sync"
	"time"
)

/**
主启动类
*/
func main1() {
	//创建 一个接耦类
	m := new(MainBox)
	//调用初始化函数，
	m.Init()
	//创建4个消息的接受者。go 关键字 表示开启线程。
	for i := 0; i < 4; i++ {
		go func() {
			//m的create方法 返回一个box ，则此线程就从此box读取数据
			create := m.Create()
			receiver(create)
		}()
	}
	//主线程sleep 1 秒，
	time.Sleep(time.Second * 1)
	log.Println("开始发送消息")
	//得到中间解耦类map的所有的key
	ids := m.GetAllId()
	//遍历所有的key，来开启线程来做。
	for index, id := range ids {
		/**
		请注意这里的赋值操作，
		因为 在循环体中开启线程是不能引用循环体中的变量的，得重新赋值一个变量，否则就只有拿到最后的值。
		*/
		i := index
		d := id
		//开启发送者线程
		go func() {
			b := m.Get(d)
			message := "消息" + strconv.Itoa(i)
			log.Print("发送的消息为 -------->:", message)
			b <- "消息" + message
		}()
	}
	//主线程等待，要注意，在go中主线程有绝对的控制权，主线程如果结束，一切都会over
	time.Sleep(time.Hour * 10)
}

//接受消息，此方法就是从 通道用拿数据。注意，此处的通道是没有缓存的通道，异步通道。
func receiver(box2 box) {
	boxf := <-box2
	log.Print("收到的值:", boxf)
}

//声明 box类型，他的实际类型为 存放string 的chan
type box chan string

//中间结耦类
type MainBox struct {
	//维护的map，key：自增需要，map：维护的 通道
	chanMap map[int]box
	//key的自增序号
	sid int
	//互斥锁
	sync.Mutex
}

/**
得到id的自增序列
*/
func (this *MainBox) getId() int {
	this.sid = this.sid + 1
	return this.sid
}

//这是我自己写的初始化函数，用来初始化
func (this *MainBox) Init() {
	//创建一个 map
	this.chanMap = make(map[int]box)
	//设置自增序列为的初始值为0
	this.sid = 0
}

//得到一个 chan
func (this *MainBox) Create() box {
	//lock不支持重入
	this.Lock()
	//在函数的最后调用 defer修饰的函数
	defer this.Unlock()
	//创造一个box
	strings := make(box)
	//得到一个自增序列
	id := this.getId()
	//设置kv
	this.chanMap[id] = strings
	return strings
}

/**
根据id拿到一个chan
*/
func (this *MainBox) Get(key int) box {
	//这里加锁是因为多个协程要访问同一个共享变量。
	this.Lock()
	defer this.Unlock()
	strings := this.chanMap[key]
	//从map中删除 ，通过key
	delete(this.chanMap, key)
	return strings
}

/**
得到map的所有key，要通过key才能拿到chan
*/
func (this *MainBox) GetAllId() []int {
	ints := make([]int, 0)
	for i, _ := range this.chanMap {
		ints = append(ints, i)
	}
	return ints
}

//init函数，是go中在使用包之前会默认调用的函数。
//这里就是设置日志的打印方式，go提供了简单的日志
func init() {
	log.SetFlags(log.Lshortfile)
}
func main() {
	//定义一个chan，
	strinChan := make(chan string, 10)
	//开一个发送线程
	go func() {
		//一直发送
		for {
			strinChan <- time.Now().String()
		}
	}()
	//主线程等待strinChan 里面的数据
	for {
		time.Sleep(time.Second * 1)
		//接受
		message := <-strinChan
		log.Println(message)
	}
}
