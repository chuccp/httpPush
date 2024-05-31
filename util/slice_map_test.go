package util

import (
	"log"
	"runtime"
	"strconv"
	"testing"
)

func BenchmarkName(b *testing.B) {
	for i := 0; i < b.N; i++ {

	}
}
func BenchmarkNameMap(t *testing.B) {
	for i := 0; i < 10000; i++ {
		sliceMap := new(SliceMap[string])

		for i := 0; i < 2; i++ {
			sliceMap.Put(strconv.Itoa(i), strconv.Itoa(i))
		}
	}
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	log.Printf("Alloc:%d(bytes) HeapIdle:%d(bytes) HeapReleased:%d(bytes)", ms.Alloc, ms.HeapIdle, ms.HeapReleased)

}
func BenchmarkNameMap2(t *testing.B) {

	for i := 0; i < 10000; i++ {
		sliceMap := make(map[string]string)

		for i := 0; i < 2; i++ {
			sliceMap[strconv.Itoa(i)] = strconv.Itoa(i)
		}

	}

	//t.ReportAllocs()
	/**
	2024/05/31 14:56:48 Alloc:291344(bytes) HeapIdle:3088384(bytes) HeapReleased:302
	2848(bytes)
	2024/05/31 14:56:48 Alloc:291344(bytes) HeapIdle:3088384(bytes) HeapReleased:302
	2848(bytes)
	2024/05/31 14:56:48 Alloc:294192(bytes) HeapIdle:3088384(bytes) HeapReleased:302
	2848(bytes)
	2024/05/31 14:56:48 Alloc:294192(bytes) HeapIdle:3088384(bytes) HeapReleased:301
	4656(bytes)
	*/

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	log.Printf("Alloc:%d(bytes) HeapIdle:%d(bytes) HeapReleased:%d(bytes)", ms.Alloc, ms.HeapIdle, ms.HeapReleased)
}
