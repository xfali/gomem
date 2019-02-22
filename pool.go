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

type Pool interface {
    Get() interface{}
    Put(interface{})
}
