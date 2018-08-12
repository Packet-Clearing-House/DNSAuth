package utils

import (
	"runtime"
	"sync"
	"time"
)

var statsMux sync.Mutex
var startTime = time.Now().Unix()

func GetMemory() uint64 {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return mem.Sys
}

func GetUptime() uint64 {
	return uint64(time.Now().Unix() - startTime)
}

func GetNumGoRoutine() uint64 {
	statsMux.Lock()
	defer statsMux.Unlock()
	return uint64(runtime.NumGoroutine())
}
