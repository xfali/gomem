/**
 * Copyright (C) 2019, Xiongfa Li.
 * All right reserved.
 * @author xiongfa.li
 * @date 2019/2/22
 * @time 9:20
 * @version V1.0
 * Description: 
 */

package gomem

//对象池接口，比系统自带sync.Pool功能强，性能待测试
type Pool interface {
    /*
     初始化对象池
     return ：
        1、获得对象的channel
        2、回收对象的channel
     注意：直接使用两个channel，可能使Pool内部的缓存机制及附加功能失效，更推荐使用Get、Put方法
        如果需要自己控制超时时间，可尝试使用此方法。可查看不同Pool的方法注释及具体实现，看是否支持直接使用channel的方式及产生的影响。
        内置的Pool有：
        1、recyclePool: 可以使用
        2、commonPool: 禁止使用
        3、commonPool2: 不建议使用
     */
    Init() (<-chan interface{}, chan<- interface{})
    /*
     关闭对象池，回收资源
     */
    Close()
    /*
     获得一个对象，根据实现可能会阻塞
     */
    Get() interface{}
    /*
     回收一个对象，根据实现可能会阻塞
     */
    Put(interface{})
}
