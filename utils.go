package utils

import (
	"fmt"
	"strconv"
)

/**
Author:将单位进制全部换算为B单位
 */
func ParseSize(size string) (int64,error) {
	if len(size) <= 2 {
		return 0,fmt.Errorf("请输入指定内存格式！exp:1K,1M,1G")
	}
	suffix := size[len(size)-1 : ] //获取单位后缀
	sizeStr,_ := strconv.ParseInt(size[:len(size)-1], 10, 64) //数值
	switch suffix {
	case "K":
		return sizeStr << 10,nil
	case "M":
		return sizeStr << 20,nil
	case "G":
		return sizeStr << 30,nil
	default:
		return 0, fmt.Errorf(suffix + "单位类型待拓展")
	}
}
