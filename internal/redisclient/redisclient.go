// Package redisclient 提供全域 lazy-initialized Redis client。
//
// 用 sync.Once 保證整個 process 內只會建立一次連線，避免
// 原本「每個 request 都 NewClient + Close」造成的連線洩漏與
// 大量短連線的浪費。PoolSize / MinIdleConns 等參數可在建立時
// 透過 Options 設定，環境變數讀取則交給 caller（main.go）。
package redisclient

import (
	"sync"

	"github.com/go-redis/redis/v8"
)

var (
	client *redis.Client
	once   sync.Once
)

// Init 以給定的 Options 建立全域 client。整個 process 只會生效一次，
// 之後重複呼叫會直接傳回同一個 instance。
//
// 建議在 main 啟動時呼叫一次即可；執行期間不應再傳入不同 Options。
func Init(opts *redis.Options) *redis.Client {
	once.Do(func() {
		client = redis.NewClient(opts)
	})
	return client
}

// Get 取得已建立的 client。若尚未 Init 過，會回傳 nil（呼叫端需自行判斷）。
func Get() *redis.Client {
	return client
}

// Close 關閉底層連線池。建議在 graceful shutdown 階段呼叫。
func Close() error {
	if client == nil {
		return nil
	}
	return client.Close()
}
