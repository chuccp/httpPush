package user

import (
	"fmt"
	"testing"
	"time"
)

func TestNamePerformance(t *testing.T) {
	//testAddPerformance(100000, 1000000)
	// 测试遍历操作
	testIterationPerformance(1000000, 1000000)
}

// 测试添加操作
func testAddPerformance(size, iterations int) {
	// 创建map和切片用于测试
	//var mapData map[int]int = make(map[int]int)
	var sliceData []int

	// 开始计时
	startTime := time.Now()

	// 添加元素到map和切片中
	for i := 0; i < iterations; i++ {
		//mapData[i] = i * 2 // 使用简单的映射关系进行填充
		sliceData = append(sliceData, i*2) // 添加到切片中
	}

	// 计算并打印执行时间
	elapsedTime := time.Since(startTime)
	//fmt.Printf("添加操作（map）：%v\n", elapsedTime)
	fmt.Printf("添加操作（切片）：%v\n", elapsedTime)
}

// 测试遍历操作
func testIterationPerformance(size, iterations int) {
	// 为map和切片预填充数据
	var mapData map[int]int = makeTestMap(size)
	//var sliceData []int = makeTestSlice(size)

	// 开始计时
	startTime := time.Now()

	// 遍历map和切片并执行某些操作（例如求和）
	sumMap := 0
	//sumSlice := 0
	for _, value := range mapData {
		sumMap += value // 对map的每个值进行求和操作
	}
	//for _, value := range sliceData {
	//	sumSlice += value // 对切片的每个值进行求和操作
	//}

	// 计算并打印执行时间
	elapsedTimeMap := time.Since(startTime)
	//fmt.Printf("遍历操作（map）：%v\n", elapsedTimeMap)
	elapsedTimeSlice := time.Since(startTime) - elapsedTimeMap // 减去map的遍历时间，因为它们是同时进行的
	fmt.Printf("遍历操作（切片）：%v\n", elapsedTimeSlice)
}

// 创建具有一定规模的数据的测试map和切片
func makeTestMap(size int) map[int]int {
	var data map[int]int = make(map[int]int)
	for i := 0; i < size; i++ {
		data[i] = i * 2 // 使用简单的映射关系进行填充
	}
	return data
}
func makeTestSlice(size int) []int {
	var data []int = make([]int, size) // 使用默认值初始化切片（全为0）
	for i := 0; i < size; i++ {        // 填充切片数据（与map相同）
		data[i] = i * 2 // 使用简单的映射关系进行填充，这会导致额外的内存分配和拷贝操作，因为切片的长度大于初始容量时，底层数组可能会被重新分配和拷贝数据。这可能会影响性能，因此在实际应用中，建议预先分配足够的内存空间或使用其他方法避免频繁的内存分配。但在本测试中，我们主要关注切片的基本性能，因此忽略了这个因素。
	}
	return data
}
