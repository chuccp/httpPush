package util

import (
	"fmt"
	"runtime/debug"
	"sync"
)

func RecoverGo(handle func()) {
	go func() {
		wg := new(sync.WaitGroup)
		for {
			wg.Add(1)
			go func() {
				defer func() {
					if err := recover(); err != nil {
						s := string(debug.Stack())
						fmt.Printf("err=%v, stack=%s\n", err, s)
						wg.Done()
					}
				}()
				handle()
			}()
			wg.Wait()
		}
	}()
}

func Go(handle func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				s := string(debug.Stack())
				fmt.Printf("err=%v, stack=%s\n", err, s)

			}
		}()
		handle()
	}()
}
