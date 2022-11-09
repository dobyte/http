/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2021/8/26 4:51 下午
 * @Desc: TODO
 */

package rand

import (
	"math/rand"
	"time"
)

const seed = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Str generate a string of specified length.
func Str(length int) (lastStr string) {
	rand.Seed(time.Now().UnixNano())

	pos, seedLen := 0, len(seed)
	for i := 0; i < length; i++ {
		pos = rand.Intn(seedLen)
		lastStr += seed[pos : pos+1]
	}

	return lastStr
}
