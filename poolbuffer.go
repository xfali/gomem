/**
 * Copyright (C) 2019, Xiongfa Li.
 * All right reserved.
 * @author xiongfa.li
 * @date 2019/2/21
 * @time 14:58
 * @version V1.0
 * Description: 
 */

package gomem

import (
    "container/list"
    "time"
)

type Manager struct {
    get      chan interface{}
    give     chan interface{}
    stop     chan bool
    interval time.Duration
    New      func() interface{}
    Delete  func(interface{})
}

type poolObject struct {
    when time.Time
    obj  interface{}
}

func New(New func() interface{}, Delete func(interface{}), gcInterval time.Duration) *Manager {
    return &Manager{
        get:      make(chan interface{}),
        give:     make(chan interface{}),
        stop:     make(chan bool),
        interval: gcInterval,
        New:      New,
        Delete:  Delete}
}

func (m *Manager) Start() (<-chan interface{}, chan<- interface{}) {
    go func() {
        queue := list.New()
        timer := time.NewTimer(m.interval)
        for {
            select {
            case <-m.stop:
                return
            case <-timer.C:
                for e := queue.Front(); e != nil; e = e.Next() {
                    if time.Since(e.Value.(poolObject).when) > m.interval {
                        queue.Remove(e)
                        if m.Delete != nil {
                            m.Delete(e.Value.(poolObject).obj)
                        }
                        e.Value = nil
                    }
                }
                timer = time.NewTimer(m.interval)
            default:
            }

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
            }
        }
    }()
    return m.get, m.give
}

func (m *Manager) Stop() {
    close(m.stop)
}

func (m *Manager) Get() interface{} {
    return <-m.get
}

func (m *Manager)Put(i interface{}) {
    m.give <- i
}
