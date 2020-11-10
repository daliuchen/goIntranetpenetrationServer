package threadtest

import (
	"log"
	"strconv"
	"sync"
)

//此结构体是为了同步数据使用的。数组下标和数据
type Test struct {
	Index int
	Value interface{}
}

func init() {
	log.SetFlags(log.Lshortfile)
}
func main() {
	d := new(DataBase)
	d.Init()
	//得到数据集合
	list := d.getDataList()
	//读取线程交互的通道
	chanInterface := make(chan Test, 20)

	//这是为了主线程等待子线程结束和特别设计的。
	c := make(chan struct{})
	//开启的线程数量
	count := 6

	//开启读线程读取
	go func() {
		//遍历集合
		for i, value := range list {
			//给每一个数据 + 1；数据做操作
			one := dataAddOne(value)
			test := Test{
				Index: i,
				Value: one,
			}
			chanInterface <- test
			//此线程结束之后，就给这个线程结束
			c <- struct{}{}
		}
	}()

	for i := 0; i < 5; i++ {
		//开启更改线程
		go func() {
			//读取通道中的信息，
			for test := range chanInterface {
				//做更新操作
				d.UpdateData(test.Index, test.Value)
			}
			//如果通道中没有消息了，这个线程也就结束了，就给c通道给个消息
			//关闭通信通道
			close(chanInterface)
			c <- struct{}{}
		}()
	}
	//主线程等待子线程结束/
	for range c {
		count--
		if count == 0 {
			log.Print(d.dataList)
			close(c)
		}
	}
}

type DataBase struct {
	//数据集合
	dataList [10]interface{}
	sync.Mutex
}

func (this *DataBase) Init() {
	//赋值
	for i := 0; i < 10; i++ {
		this.dataList[i] = strconv.Itoa(i)
	}
}
func (this *DataBase) getDataList() [10]interface{} {
	return this.dataList
}

//更新数据，通过下标和数据
func (this *DataBase) UpdateData(index int, param interface{}) {
	this.Lock()
	defer this.Unlock()
	this.dataList[index] = param
}

//数据修改，传进来的数据+1 返回
func dataAddOne(param interface{}) interface{} {
	s := param.(string)
	atoi, _ := strconv.Atoi(s)
	return strconv.Itoa(atoi + 1)
}
