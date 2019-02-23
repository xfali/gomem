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
    "container/list"
    "time"
)

type PooledObjectFactory interface {
    //对象被激活时调用
    ActivateObject(interface{})
    //对象处于idle状态，被定时释放时调用
    DestroyObject(interface{})
    //创建对象时调用
    MakeObject() interface{}
    //对象回收之后进入idle状态时被调用
    PassivateObject(interface{})
    //验证对象是否有效
    ValidateObject(interface{}) bool
}

type CommonPool struct {
    //池中最小保留的idle对象的数量，默认8
    MinIdle int
    //最大对象数量,默认32
    MaxSize int
    //获取资源的等待时间,BlockWhenExhausted 为 true 时有效。-1 代表无时间限制，一直阻塞直到有可用的资源
    MaxWaitMillis time.Duration
    //对象空闲的最小时间，达到此值后空闲对象将可能会被移除。-1 表示不移除；默认 30 分钟
    MinEvictableIdleTimeMillis time.Duration
    //资源回收协程执行一次回收操作，回收资源的数量。默认 3
    NumTestsPerEvictionRun int
    //创建对象时是否调用 Factory.ValidateObject 方法，默认 false
    TestOnCreate bool
    //获取对象时是否调用 Factory.ValidateObject 方法，默认 false
    TestOnBorrow bool
    //释放对象时是否调用 Factory.ValidateObject 方法，默认 false
    TestOnReturn bool
    //对象空闲时是否调用 Factory.ValidateObject 方法，默认 false
    TestWhileIdle bool
    //回收资源协程的执行周期，默认 -1 表示不定时回收
    TimeBetweenEvictionRunsMillis time.Duration
    //资源耗尽时，是否阻塞等待获取资源，默认 false
    BlockWhenExhausted bool
    //是否接受外部Object，默认false(暂时没有作用)
    AcceptExternalObj bool
    //对象工厂
    Factory PooledObjectFactory

    //inner vars
    getChan  chan interface{}
    putChan  chan interface{}
    stop     chan bool
    curCount int
    init     bool
}

const (
    IDLE       = iota //在池中，处于空闲状态
    ALLOCATED         //被使用中
    EVICTION          //正在被逐出器验证
    VALIDATION        //正在验证
    INVALID           //驱逐测试或验证失败并将被销毁
    ABANDONED         //对象被客户端拿出后，长时间未返回池中，或没有调用 use 方法，即被标记为抛弃的
    READY             //可以被给客户端使用
)

type poolObject struct {
    when  time.Time
    State int
    obj   interface{}
}

func (p *CommonPool) initDefault() {
    if p.Factory == nil {
        panic("Factory is Empty")
    }
    if p.init {
        return
    }
    p.init = true
    if p.MinIdle == 0 {
        p.MinIdle = 8
    }
    if p.MaxSize == 0 {
        p.MaxSize = 32
    }
    if p.MaxWaitMillis == 0 {
        p.MaxSize = -1
    }
    if p.MinEvictableIdleTimeMillis == 0 {
        p.MinEvictableIdleTimeMillis = 30 * time.Minute
    }
    if p.TimeBetweenEvictionRunsMillis == 0 {
        p.TimeBetweenEvictionRunsMillis = -1
    }
    if p.NumTestsPerEvictionRun == 0 {
        p.NumTestsPerEvictionRun = 3
    }

    p.getChan = make(chan interface{})
    p.putChan = make(chan interface{})
    p.stop = make(chan bool)

    p.curCount = 0
}

//支持获取channel，但对factory的支持以及获取超时时间的配置项失效；支持回收channel，但对factory的支持失效，不建议使用
func (p *CommonPool) Init() (<-chan interface{}, chan<- interface{}) {
    p.initDefault()
    go func() {
        queue := list.New()
        var timer *time.Timer
        if p.TimeBetweenEvictionRunsMillis == -1 {
            timer = &time.Timer{C: make(chan time.Time)}
        } else {
            timer = time.NewTimer(p.TimeBetweenEvictionRunsMillis)
        }

        for {
            //fmt.Println("main loop")
            if queue.Len() == 0 {
                o := p.make()
                //到达对象池上限
                if o == nil {
                    //等待用户归还对象
                    got := false
                    for !got {
                        select {
                        case <-p.stop:
                            return
                        case b := <-p.putChan:
                            if p.idleObj(b) {
                                queue.PushBack(poolObject{time.Now(), IDLE, b})
                                got = true
                            }
                        case <-timer.C:
                            //fmt.Println("in sub loop")
                            e := queue.Front()
                            next := e
                            for e != nil && queue.Len() > p.MinIdle {
                                next = e.Next()
                                if p.MinEvictableIdleTimeMillis > 0 && time.Since(e.Value.(poolObject).when) > p.MinEvictableIdleTimeMillis {
                                    queue.Remove(e)
                                    p.destoryObj(e.Value.(poolObject).obj)
                                    e.Value = nil
                                }
                                e = next
                            }
                            timer = time.NewTimer(p.TimeBetweenEvictionRunsMillis)
                        }
                    }
                } else {
                    queue.PushBack(poolObject{time.Now(), ALLOCATED, o})
                }
            }
            e := queue.Front()
            state := e.Value.(poolObject).State
            if state == IDLE || state == ALLOCATED {
                p.Factory.ActivateObject(e.Value.(poolObject).obj)
            }
            select {
            case <-p.stop:
                return
            case b := <-p.putChan:
                if p.idleObj(b) {
                    queue.PushBack(poolObject{ time.Now(), IDLE, b})
                }
            case p.getChan <- e.Value.(poolObject).obj:
                queue.Remove(e)
            case <-timer.C:
                e := queue.Front()
                next := e
                for e != nil && queue.Len() > p.MinIdle {
                    next = e.Next()
                    if p.MinEvictableIdleTimeMillis > 0 && time.Since(e.Value.(poolObject).when) > p.MinEvictableIdleTimeMillis {
                        queue.Remove(e)
                        p.destoryObj(e.Value.(poolObject).obj)
                        e.Value = nil
                    }
                    e = next
                }
                timer = time.NewTimer(p.TimeBetweenEvictionRunsMillis)
            }
        }
    }()

    return p.getChan, p.putChan
}

func (p *CommonPool) Close() {
    close(p.stop)
}

func (p *CommonPool) syncMake() interface{} {
    if p.curCount < p.MaxSize {
        o := p.Factory.MakeObject()
        p.curCount++
        return o
    }
    return nil
}

func (p *CommonPool) idleObj(i interface{}) bool {
    if i == nil {
        return false
    }

    if p.TestWhileIdle {
        if !p.Factory.ValidateObject(i) {
            return false
        }
    }
    p.Factory.PassivateObject(i)
    return true
}

func (p *CommonPool) destoryObj(i interface{}) {
    if i != nil {
        p.Factory.DestroyObject(i)
    }
}

func (p *CommonPool) make() interface{} {
    i := p.syncMake()
    if i != nil {
        if p.TestOnCreate {
            if !p.Factory.ValidateObject(i) {
                return nil
            }
        }
    }
    return i
}

func (p *CommonPool) Get() interface{} {
    var ret interface{}
    if !p.BlockWhenExhausted {
        select {
        case ret = <-p.getChan:
            if p.TestOnBorrow {
                if !p.Factory.ValidateObject(ret) {
                    return nil
                }
            }
            return ret
        default:
            return nil
        }
    }

    if p.MaxWaitMillis == -1 {
        ret := <-p.getChan
        if p.TestOnBorrow {
            if !p.Factory.ValidateObject(ret) {
                return nil
            }
        }
        return ret
    }

    select {
    case ret = <-p.getChan:
        if p.TestOnBorrow {
            if !p.Factory.ValidateObject(ret) {
                return nil
            }
        }
        return ret
    case <-time.After(p.MaxWaitMillis):
        break
    }

    return ret
}

func (p *CommonPool) Put(i interface{}) {
    if p.TestOnReturn {
        if !p.Factory.ValidateObject(i) {
            return
        }
    }
    p.putChan <- i
}

type DummyFactory func() interface{}

func (f *DummyFactory) ActivateObject(interface{})      {}
func (f *DummyFactory) DestroyObject(interface{})       {}
func (f *DummyFactory) MakeObject() interface{}         { return (*f)() }
func (f *DummyFactory) PassivateObject(interface{})     {}
func (f *DummyFactory) ValidateObject(interface{}) bool { return true }

type DefaultFactory struct {
    Activate  func(interface{})
    Destroy   func(interface{})
    Make      func() interface{}
    Passivate func(interface{})
    Validate  func(interface{}) bool
}

func (f *DefaultFactory) ActivateObject(i interface{}) {
    if f.Activate != nil {
        f.Activate(i)
    }
}
func (f *DefaultFactory) DestroyObject(i interface{}) {
    if f.Destroy != nil {
        f.Destroy(i)
    }
}
func (f *DefaultFactory) MakeObject() interface{} {
    if f.Make == nil {
        panic("Make func is nil")
    }
    return f.Make()
}
func (f *DefaultFactory) PassivateObject(i interface{}) {
    if f.Passivate != nil {
        f.Passivate(i)
    }
}
func (f *DefaultFactory) ValidateObject(i interface{}) bool {
    if f.Validate != nil {
        return f.Validate(i)
    }
    return true
}
