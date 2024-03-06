package main

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

type PortScanner struct {
	ip   string
	lock *semaphore.Weighted
}

func ULimit() int64 {
	// out, err1 := exec.Command("ulimit -n").Output()
	// if err1 != nil {
	// 	panic(err1)
	// }
	s := strings.TrimSpace(string("100"))
	i, err := strconv.ParseInt(s, 10, 64)

	if err != nil {
		panic(err)
	}

	return i
}

func ScanPort(ip string, port int, timeout time.Duration) {
	target := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		if strings.Contains(err.Error(), "Too many open files") {
			time.Sleep(timeout)
			ScanPort(ip, port, timeout)
		}
		return
	}

	defer conn.Close()
	fmt.Println("Port:", port, "Open")
}

func (ps *PortScanner) Start(f, l int, timeout time.Duration) {
	wg := sync.WaitGroup{}
	defer wg.Wait()
	for port := f; port <= l; port++ {
		ps.lock.Acquire(context.TODO(), 1)
		wg.Add(1)

		go func(port int) {
			defer ps.lock.Release(1)
			defer wg.Done()
			ScanPort(ps.ip, port, timeout)
		}(port)
	}
}

func main() {
	ps := &PortScanner{
		ip:   "127.0.0.1",
		lock: semaphore.NewWeighted(ULimit()),
	}
	ps.Start(1, 65535, 1*time.Second)
}
