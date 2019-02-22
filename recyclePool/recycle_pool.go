/**
 * Copyright (C) 2019, Xiongfa Li.
 * All right reserved.
 * @author xiongfa.li
 * @date 2019/2/21
 * @time 14:58
 * @version V1.0
 * Description: 
 */

package recyclePool

import (
    "container/list"
    "time"
)

type RecyclePool struct {
    Interval time.Duration
    New      func() interface{}
    Delete   func(interface{})

    get  chan interface{}
    give chan interface{}
    stop chan bool
}

type poolObject struct {
    when time.Time
    obj  interface{}
}

func (m *RecyclePool) Start() (<-chan interface{}, chan<- interface{}) {
    m.get = make(chan interface{})
    m.give = make(chan interface{})
    m.stop = make(chan bool)

    go func() {
        queue := list.New()
        var timer *time.Timer
        if m.Interval == 0 {
            timer = &time.Timer{C: make(chan time.Time)}
        } else {
            timer = time.NewTimer(m.Interval)
        }
        for {
            if queue.Len() == 0 {
                queue.PushBack(poolObject{when: time.Now(), obj: m.New()})
            }
            e := queue.Front()

            select {
            case <-m.stop:
                return
            case b := <-m.give:
                //timer.Stop()
                queue.PushBack(poolObject{when: time.Now(), obj: b})
            case m.get <- e.Value.(poolObject).obj:
                //timer.Stop()
                queue.Remove(e)
            case <-timer.C:
                e := queue.Front()
                next := e
                for e != nil {
                    next = e.Next()
                    if time.Since(e.Value.(poolObject).when) > m.Interval {
                        queue.Remove(e)
                        if m.Delete != nil {
                            m.Delete(e.Value.(poolObject).obj)
                        }
                        e.Value = nil
                    }
                    e = next
                }
                timer = time.NewTimer(m.Interval)
            }
        }
    }()
    return m.get, m.give
}

//支持直接使用获取、回收channel，可以使用
func (m *RecyclePool) Init() (<-chan interface{}, chan<- interface{}) {
    return m.Start()
}

func (m *RecyclePool) Close() {
    close(m.stop)
}

func (m *RecyclePool) Get() interface{} {
    return <-m.get
}

func (m *RecyclePool) Put(i interface{}) {
    m.give <- i
}
