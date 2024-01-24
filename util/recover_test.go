package util

import (
	"testing"
	"time"
)

func TestRecoverGo(t *testing.T) {

	//go func() {
	//	var arr [3]int
	//	defer func() {
	//		err := recover()               // recover() 捕获panic异常，获得程序执行权。
	//		fmt.Println("recover()后的内容！！") // recover()后的内容会正常打印
	//		if err != nil {
	//			fmt.Println(err) // runtime error: index out of range
	//		}
	//	}()
	//	index := 10
	//	arr[index] = 10 // 会抛出panic异常 (数组下标越界)
	//
	//	fmt.Println("异常发生后的内容！！") // 异常之后的内容不会打印
	//}()
	//
	//time.Sleep(time.Second * 5)

	RecoverGo(func() {
		time.Sleep(time.Second * 2)
		panic("1111")

	})

	time.Sleep(time.Second * 10)

}
func TestRecoverGo001(t *testing.T) {
	Go(func() {

		panic("1111")
	})

	time.Sleep(time.Second * 10)
	t.Log("1111111111")
}
