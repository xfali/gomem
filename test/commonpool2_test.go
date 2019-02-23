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
    commonPool2 "gomem/commonPool2"
    "testing"
    "time"
)

type b int

func (f *b) ActivateObject(i interface{}) { fmt.Printf("ActivateObject %v\n", i) }
func (f *b) DestroyObject(i interface{})  { fmt.Printf("DestroyObject %v\n", i) }
func (f *b) MakeObject() interface{} {
    fmt.Printf("MakeObject\n")
    return "test"
}
func (f *b) PassivateObject(i interface{}) { fmt.Printf("PassivateObject %v\n", i) }
func (f *b) ValidateObject(i interface{}) bool {
    fmt.Printf("ValidateObject %v\n", i)
    return true
}

func TestCommonPool2(t *testing.T) {
    //f := commonPool2.DummyFactory(func() interface{} {
    //    fmt.Println("create!")
    //    return "test"
    //})
    f := b(1)
    pb := commonPool2.CommonPool{
        MinIdle:                       10,
        MaxSize:                       20,
        BlockWhenExhausted:            true,
        TimeBetweenEvictionRunsMillis: 2 * time.Second,
        Factory:                       &f,
        MaxWaitMillis:                 time.Second * 10,
        TestOnReturn:                  true,
        TestWhileIdle:                 true,
        TestOnBorrow:                  true,
        TestOnCreate:                  true,
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
        fmt.Printf("value %v\n", buf)
        fmt.Printf("use time :%d ms\n", time.Since(now)/time.Millisecond)
        l.PushBack(buf)
    }
}

func TestCommonPool2_2(t *testing.T) {
    //f := commonPool2.DummyFactory(func() interface{} {
    //    fmt.Println("create!")
    //    return "test"
    //})
    f := b(1)
    pb := commonPool2.CommonPool{
        MinIdle:                       5,
        MaxSize:                       20,
        BlockWhenExhausted:            true,
        TimeBetweenEvictionRunsMillis: 1 * time.Second,
        MinEvictableIdleTimeMillis:    1 * time.Second,
        Factory:                       &f,
        MaxWaitMillis:                 time.Second * 10,
        TestOnReturn:                  true,
        TestWhileIdle:                 true,
        TestOnBorrow:                  true,
        TestOnCreate:                  true,
    }
    pb.Init()
    defer pb.Close()

    l := list.New()
    for i := 0; i < 10; i++ {
        buf := pb.Get()
        l.PushBack(buf)
        fmt.Printf("value %v\n", buf)
    }

    for e := l.Front(); e != nil; e = e.Next() {
        pb.Put(e.Value)
    }
    time.Sleep(time.Second)
    now := time.Now()
    time.Sleep(10 * time.Second)
    fmt.Printf("%d ms\n", time.Since(now)/time.Millisecond)
}

func TestCommonPool2_subloopTimeout(t *testing.T) {
    //f := commonPool2.DummyFactory(func() interface{} {
    //    fmt.Println("create!")
    //    return "test"
    //})
    f := b(1)
    pb := commonPool2.CommonPool{
        MinIdle:                       1,
        MaxSize:                       2,
        BlockWhenExhausted:            true,
        TimeBetweenEvictionRunsMillis: 1 * time.Second,
        MinEvictableIdleTimeMillis:    1 * time.Second,
        Factory:                       &f,
        MaxWaitMillis:                 time.Second * 10,
        TestOnReturn:                  true,
        TestWhileIdle:                 true,
        TestOnBorrow:                  true,
        TestOnCreate:                  true,
    }
    pb.Init()
    defer pb.Close()

    l := list.New()
    go func() {
        time.Sleep(3 * time.Second)
        pb.Put(l.Front().Value)
    }()
    for i := 0; i < 10; i++ {
        buf := pb.Get()
        l.PushBack(buf)
        fmt.Printf("value %v\n", buf)
    }

    for e := l.Front(); e != nil; e = e.Next() {
        pb.Put(e.Value)
    }
    time.Sleep(time.Second)
    now := time.Now()
    time.Sleep(10 * time.Second)
    fmt.Printf("%d ms\n", time.Since(now)/time.Millisecond)
}
