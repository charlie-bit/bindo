# bindo实操题 
## 一、内存缓存系统
1. 支持设定过期时间，精度为秒级。
2. 支持设定最大内存，当内存超出时候做出合理的处理。
3. 支持并发安全。

### 思路：
#### 1.model层面：
数据（有效时间（只会过期时间，精度为秒），最后一次的存储时间、值）、读写锁（支持并发安全）、内存大小、内存清理时间；
#### 2.基础功能实现
##### 1.Get
从map中获取指定key的值（排除过期）
##### 2.Set
设置键值对，过期时间，creatAt(在没有判断key是否存在的情况，有key的前提就更新，没有就存储)
##### 3.Del
在map中删除指定key对应的值，调用内置函数delete方法
##### 4.Exists
在map中判断是否具有key值，去map返回的有效标识符
##### 5.Flush
初始化map
##### 6.Keys
统计map中的key数量，排除过期时间的key
##### 7.SetMaxMemory
设置最大内存，需要转换内存单位

### 拓展功能 -- 设定最大内存之后,内存超出的处理
##### 1.定时GC，清理过期的元素 
##### 2.手动GC，清理过期的元素  采用ticker完成定时清理功能，但是增加了手动停止的选项
##### 3.拓展最大内存（不推荐）


