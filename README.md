# gomem

## Pool Interface
*  初始化对象池

    Init() (<-chan interface{}, chan<- interface{})

*  关闭对象池，回收资源

    Close()

*  获得一个对象，根据实现可能会阻塞

    Get() interface{}

*  回收一个对象，根据实现可能会阻塞

    Put(interface{})

## 内置三种对象池
* ### RecyclePool
    简单的带回收的对象池。

* ### CommonPool
    基于带缓存Channel实现的对象池。

* ### CommonPool2
    类似Apache CommonPool2实现的Go对象池，机制与Recycle Pool一致，但功能更丰富。

