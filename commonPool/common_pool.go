/**
 * Copyright (C) 2019, Xiongfa Li.
 * All right reserved.
 * @author xiongfa.li
 * @date 2019/2/21
 * @time 17:59
 * @version V1.0
 * Description: 
 */

package commonPool

import (
    "gomem"
    "math"
    "sync"
    "time"
)

type CommonPool struct {
    MaxIdle     int
    MaxSize     int
    WaitTimeout time.Duration
    New         func() interface{}
    queue       chan interface{}
    curCount    int
    mutex       sync.Mutex

    init bool
}

func (p *CommonPool) initDefault() {
    if p.init {
        return
    }
    p.init = true
    if p.MaxIdle == 0 {
        p.MaxIdle = 16
    }
    if p.MaxSize == 0 {
        p.MaxSize = 32
        if p.MaxIdle < p.MaxSize {
            p.MaxIdle = p.MaxSize
        }
    }
    if p.WaitTimeout == 0 {
        p.WaitTimeout = time.Duration(math.MaxInt64)
    }
    if p.queue == nil {
        p.queue = make(chan interface{}, p.MaxIdle)
    }
    p.curCount = 0
}

func (p *CommonPool) Build() gomem.Pool {
    p.initDefault()
    return p
}

func (p *CommonPool) make() interface{} {
    p.mutex.Lock()
    defer p.mutex.Unlock()

    if p.curCount < p.MaxSize {
        p.curCount++
        return p.New()
    }
    return nil
}

func (p *CommonPool) Get() interface{} {
    var ret interface{}

    if len(p.queue) == 0 {
        ret = p.make()
        if ret != nil {
            return ret
        }
    }
    select {
    case ret = <-p.queue:
        return ret
    case <-time.After(p.WaitTimeout):
        break
    }
    if ret == nil {
        return p.make()
    }
    return ret
}

func (p *CommonPool) Put(i interface{}) {
    p.queue <- i
}
