package models

import (
	"bindo/utils"
	"fmt"
	"log"
	"sync"
	"time"
)

/**
Author:charlie
Description:一个简单的缓存系统--底层源码
*/

type Cache interface {
	//size 是⼀一个字符串串。⽀支持以下参数: 1KB，100KB，1MB，2MB，1GB 等
	SetMaxMemory(size string) bool
	// 设置⼀一个缓存项，并且在expire时间之后过期
	Set(key string, val interface{}, expire time.Duration)
	Get(key string) (interface{}, bool) // 获取⼀一个值
	// 删除⼀一个值
	Del(key string) bool
	// 检测⼀一个值 是否存在
	Exists(key string) bool
	// 清空所有值
	Flush() bool
	// 返回所有的key 多少
	Keys() int64

	//补充
	StartGC(interval int) error

	AutoGC(interval time.Duration)
}

//定义结构体存储数据
type CatchItem struct {
	Data       map[string]CatchItem //缓存中的数据以key-value形式存储
	Expired    time.Duration        //过期时间
	LastAccess time.Time            //最后一次接受时间
	Val        interface{}          //值
}

//Catch中的数据在数据结构体中进行补充
type CatchStrut struct {
	item *CatchItem //数据
	//size 是⼀一个字符串串。⽀支持以下参数: 1KB，100KB，1MB，2MB，1GB 等
	Size int64 //内存大小
	//防止内存溢出 -- 定时任务按时清理超出>>Expired time
	TimeInterval time.Duration //与过期时间保持一致，以s为单位
	Stop         chan bool     //定时清理GC，停止标识
	dataLock     sync.RWMutex  //读写锁
}

//存入
//TODO:若元素已经存在--更新值
//@param:key-键，val值，expire-过期时间
func (c *CatchStrut) Set(key string, val interface{}, expire time.Duration) {
	/**
	写锁
	*/
	log.Printf("开始存入 %s-%v", key, val)
	c.dataLock.Lock()         //加锁
	defer c.dataLock.Unlock() //解锁
	var item CatchItem
	item.LastAccess = time.Now() //开始存储的时间
	if expire > 0 {
		item.Expired = expire //过期时间
	}
	item.Val = val
	var data = make(map[string]CatchItem)
	data[key] = item
	item.Data = data
	c.item = &item
	log.Println("存入成功")
}

// 获取
//@param:key-键
//@response:值，值得有效性
//TODO：只获取有效期内的值
func (c *CatchStrut) Get(key string) (interface{}, bool) {
	log.Printf("开始获取 %s", key)
	/**
	加读锁,读锁>写锁，先读后写
	*/
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	val := c.item.Data[key].Val
	if val != "" {
		if c.item.Data[key].Expired < time.Now().Sub(c.item.Data[key].LastAccess) {
			//如果在有效期内
			return val, true
		}
	}
	return nil, false
}

// 删除
//@param:key-键
//@response:是否删除成功
func (c *CatchStrut) Del(key string) bool {
	c.dataLock.Lock()         //加锁
	defer c.dataLock.Unlock() //解锁
	log.Printf("开始删除 %s", key)
	//先排除空
	_, flag := c.item.Data[key]
	if !flag {
		//如果值不存在
		log.Printf("删除的 %s 不存在！", key)
		return false
	} else {
		delete(c.item.Data, key)
	}
	log.Printf("删除成功")
	return true
}

// 检测⼀一个值 是否存在
func (c *CatchStrut) Exists(key string) bool {
	/**
	加读锁,读锁>写锁，先读后写
	*/
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	log.Printf("开始检测 %s", key)
	value, flag := c.item.Data[key]
	if flag {
		if value.Expired < time.Now().Sub(value.LastAccess) {
			return flag
		}
	}
	return flag
}

// 清空所有值
func (c *CatchStrut) Flush() bool {
	c.dataLock.Lock()         //加锁
	defer c.dataLock.Unlock() //解锁
	log.Printf("开始清空")
	c.item.Data = make(map[string]CatchItem)
	log.Printf("清空完成")
	return true
}

//统计key的个数
func (c *CatchStrut) Keys() int64 {
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	var count int64
	for _, value := range c.item.Data {
		if value.Expired < time.Now().Sub(value.LastAccess) {
			count++
		}
	}
	return count
}

//设置最大内存
func (c *CatchStrut) SetMaxMemory(size string) bool {
	c.dataLock.Lock()         //加锁
	defer c.dataLock.Unlock() //解锁
	parse, err := utils.ParseSize(size)
	if err != nil {
		return false
	}
	c.Size = parse
	return true
}

//实现Catch接口 -- 实现所有的方法-->>ok
func New(size string, timeInterval time.Duration) Cache {
	parse, err := utils.ParseSize(size)
	if err != nil {
		log.Println("初始化失败", err)
		return nil
	}
	var instance = &CatchStrut{
		Size:         parse,
		item:         &CatchItem{},
		TimeInterval: timeInterval,
	}
	//可以在声明的时候，手动GC
	return instance
}

//TODO：手动GC ,以s为单位
func (c *CatchStrut) StartGC(interval int) error {
	if interval <= 0 {
		return fmt.Errorf("请传入时间间隔")
	}
	c.TimeInterval = time.Duration(interval) * time.Second
	go func() {
		c.dataLock.RLock()
		c.dataLock.RUnlock()
		for {
			<-time.After(c.TimeInterval) //防止协程堵塞
			if c.item == nil {
				return
			}
			if keys := c.expiredKeys(); len(keys) != 0 {
				c.clearItems(keys)
			}
		}
	}()
	return nil
}

func (c *CatchStrut) clearItems(keys []string) {
	c.dataLock.Lock()
	defer c.dataLock.Unlock()
	for _, value := range keys {
		delete(c.item.Data, value)
	}
}

//取得所有过期的值
func (c *CatchStrut) expiredKeys() []string {
	var keys = make([]string, 0)
	c.dataLock.RLock()
	defer c.dataLock.RUnlock()
	for key, itm := range c.item.Data {
		if time.Duration(itm.Expired) < time.Now().Sub(itm.LastAccess) {
			keys = append(keys, key)
		}
	}
	return keys
}

//TODO：自动GC
func (c *CatchStrut) AutoGC(interval time.Duration) {
	c.TimeInterval = interval
	go c.run()
}

func (c *CatchStrut) run() {
	ticker := time.NewTicker(c.TimeInterval)
	for {
		select {
		case <-ticker.C: //时间间隔一到，清理内存
			//清理过期时间
			if keys := c.expiredKeys(); len(keys) != 0 {
				c.clearItems(keys)
			}
		case <-c.Stop:
			//停止定时任务
			ticker.Stop()
		}
	}
}

//如果需要手动停止 可以调用回调函数
func (c *CatchStrut)StopAutoGC()  {
	c.Stop <- true
}
