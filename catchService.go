package services

import (
	"bindo/models"
	"time"
)

type CatchService struct {
}

var catch models.Cache

//初始化
func (s *CatchService) Init(size string, timeInterval time.Duration) {
	catch = models.New(size, timeInterval)
}

//存入
func (s *CatchService) Set(key string, val interface{}, expire time.Duration) {
	catch.Set(key, val, expire)
}

// 获取
func (s *CatchService)Get(key string) (interface{}, bool) {
	return catch.Get(key)
}

// 删除
func (s *CatchService)Del(key string) bool {
	return catch.Del(key)
}

// 检测⼀一个值 是否存在
func (s *CatchService)Exists(key string) bool {
	return catch.Exists(key)
}

// 清空所有值
func (s *CatchService)Flush() bool {
	return catch.Flush()
}

//统计key的个数
func (c *CatchService) Keys() int64 {
	return catch.Keys()
}

//设置最大内存
func (c *CatchService) SetMaxMemory(size string) bool {
	return catch.SetMaxMemory(size)
}

//TODO：手动GC ,以s为单位
func (c *CatchService)StartGC(interval int) error{
	return catch.StartGC(interval)
}

//TODO：自动GC
func (c *CatchService)AutoGC(interval time.Duration)  {
	catch.AutoGC(interval)
}