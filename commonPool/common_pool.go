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
    "math"
    "sync"
    "time"
)

type CommonPool struct {
    //对象池缓存大小，当该值比MaxSize还小时，将自动调整为MaxSize。当回收的对象数量大于该值，则Put方法会阻塞
    MaxIdle     int
    //对象池最大对象数
    MaxSize     int
    //当资源耗尽时的等待资源时间
    WaitTimeout time.Duration
    //创建对象函数
    New         func() interface{}

    queue       chan interface{}
    curCount    int
    mutex       sync.Mutex

    init bool
}

//不支持获取channel、支持回收channel，禁止使用
func (p *CommonPool) Init() (<-chan interface{}, chan<- interface{}) {
    if p.init {
        return p.queue, p.queue
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

    return p.queue, p.queue
}

func (p *CommonPool) Close() {
    //close(p.queue)
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
