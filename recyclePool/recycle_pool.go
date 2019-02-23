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
    //对象空闲的最小时间，达到此值后空闲对象将可能会被移除。-1 表示不移除；默认 30 分钟
    MinEvictableIdleTimeMillis time.Duration
    //回收资源协程的执行周期，默认 -1 表示不定时回收
    TimeBetweenEvictionRunsMillis time.Duration
    //创建对象函数
    New      func() interface{}
    //释放对象函数
    Delete   func(interface{})

    get  chan interface{}
    give chan interface{}
    stop chan bool
}

type poolObject struct {
    when time.Time
    obj  interface{}
}

//支持直接使用获取、回收channel，可以使用
func (m *RecyclePool) Init() (<-chan interface{}, chan<- interface{}) {
    if m.MinEvictableIdleTimeMillis == 0 {
        m.MinEvictableIdleTimeMillis = 30*time.Minute
    }
    if m.TimeBetweenEvictionRunsMillis == 0 {
        m.TimeBetweenEvictionRunsMillis = -1
    }
    m.get = make(chan interface{})
    m.give = make(chan interface{})
    m.stop = make(chan bool)

    go func() {
        queue := list.New()
        var timer *time.Timer
        if m.TimeBetweenEvictionRunsMillis <= 0 {
            timer = &time.Timer{C: make(chan time.Time)}
        } else {
            timer = time.NewTimer(m.TimeBetweenEvictionRunsMillis)
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
                    if m.MinEvictableIdleTimeMillis > 0 && time.Since(e.Value.(poolObject).when) > m.MinEvictableIdleTimeMillis  {
                        queue.Remove(e)
                        if m.Delete != nil {
                            m.Delete(e.Value.(poolObject).obj)
                        }
                        e.Value = nil
                    }
                    e = next
                }
                timer = time.NewTimer(m.TimeBetweenEvictionRunsMillis)
            }
        }
    }()
    return m.get, m.give
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
