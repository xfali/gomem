/**
 * Copyright (C) 2019, Xiongfa Li.
 * All right reserved.
 * @author xiongfa.li
 * @date 2019/2/22
 * @time 10:42
 * @version V1.0
 * Description: 
 */

package test

import (
    "container/list"
    "fmt"
    "gomem/commonPool"
    "testing"
    "time"
)

func TestCommonPool(t *testing.T) {
    pb := commonPool.CommonPool{
        MaxIdle: 10,
        MaxSize: 20,
        New: func() interface{} {
            fmt.Println("create!")
            i := make([]byte, 1000)
            return i
        },
        WaitTimeout: time.Second * 10,
    }
    pb.Init()
    defer pb.Close()

    l := list.New()
    go func() {
        time.Sleep(time.Second)
        i := 0
        e := l.Front()
        for e != nil {
            pb.Put(e.Value)
            r := e
            e = e.Next()
            l.Remove(r)
            time.Sleep(time.Second)
            if i < 5 {
                i++
            } else {
                break
            }
        }
    }()
    for i := 0; i < 30; i++ {
        now := time.Now()
        buf := pb.Get()
        fmt.Printf("use time :%d ms\n", time.Since(now)/time.Millisecond)
        l.PushBack(buf)
    }
}
