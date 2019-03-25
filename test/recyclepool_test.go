/**
 * Copyright (C) 2019, Xiongfa Li.
 * All right reserved.
 * @author xiongfa.li
 * @date 2019/2/21
 * @time 15:31
 * @version V1.0
 * Description: 
 */

package test

import (
    "container/list"
    "fmt"
    "github.com/xfali/gomem/recyclePool"
    "math/rand"
    "runtime"
    "testing"
    "time"
)

func TestPoolBuffer(t *testing.T) {
    pb := recyclePool.RecyclePool{
        New: func() interface{} {
            fmt.Println("create!")
            i := make([]byte, 1000)
            return i
        },
        Delete: func(i interface{}) {
            fmt.Println("delete!")
            i = nil
        },
        MinEvictableIdleTimeMillis: 10*time.Second,
        TimeBetweenEvictionRunsMillis: 10 * time.Second}
    get, give := pb.Init()
    defer pb.Close()

    var m runtime.MemStats
    pool := make([]interface{}, 20)
    for i := 0; i < 100; i++ {
        buf := <-get
        if buf == nil {
            fmt.Println("Max count!")
        }
        i := rand.Intn(len(pool))
        b := pool[i]
        if b != nil {
            give <- b
        }
        pool[i] = buf
        time.Sleep(time.Second)
        bytes := 0
        for i := 0; i < len(pool); i++ {
            if pool[i] != nil {
                bytes += len(pool[i].([]byte))
            }
        }

        runtime.ReadMemStats(&m)
        fmt.Printf("%d,%d,%d,%d,%d\n", m.HeapSys, bytes, m.HeapAlloc,
            m.HeapIdle, m.HeapReleased)
    }
}

func TestPoolBuffer2(t *testing.T) {
    pb := recyclePool.RecyclePool{
        New: func() interface{} {
            fmt.Println("create!")
            i := make([]byte, 1000)
            return i
        },
        Delete: func(i interface{}) {
            fmt.Println("delete!")
            i = nil
        },
        MinEvictableIdleTimeMillis: 2*time.Second,
        TimeBetweenEvictionRunsMillis: 2 * time.Second}
    get, give := pb.Init()
    defer pb.Close()

    l := list.New()
    for i := 0; i < 100; i++ {
        buf := <-get
        l.PushBack(buf)
    }
    time.Sleep(2 * time.Second)
    e := l.Front()
    for e != nil {
        give <- e.Value
        r := e
        e = e.Next()
        l.Remove(r)
        time.Sleep(time.Second)
    }

}

func TestTimer(t *testing.T) {
    timer := time.NewTimer(time.Second)
    for {
        select {
        case <-timer.C:
            fmt.Println("timeout")
            timer = time.NewTimer(1 * time.Second)

        }
    }
}
