package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"database/sql"
	"fmt"

	_ "github.com/denisenkom/go-mssqldb"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

// ============================================================
// 配置：從環境變數讀取，支援 .env 檔
// ============================================================
//
// 啟動時會自動載入同目錄下的 .env（若存在）。
// 所有變數都帶有 fallback，若未設定則使用預設值，
// 方便本地開發直接 `go run main.go`，也方便部署時覆寫。
//
// 完整環境變數清單見 .env.example。
//
// ============================================================

// getEnv 取得環境變數，若未設定則回傳 fallback
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getEnvInt 取得整數型環境變數，轉換失敗時回傳 fallback
func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
		log.Printf("環境變數 %s=%s 不是合法整數，使用 fallback=%d\n", key, v, fallback)
	}
	return fallback
}

func main() {
	// --------------------------------------------------------
	// 載入 .env（若存在）。找不到檔案不報錯，方便正式環境直接吃系統環境變數。
	// --------------------------------------------------------
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 檔，使用系統環境變數 / fallback 值")
	} else {
		log.Println("已載入 .env 檔")
	}

	// ---- 讀取設定 ----
	serverAddr := getEnv("SERVER_ADDR", ":8080")

	redisAddr := getEnv("REDIS_ADDR", "redis-cluster.h1-redis-dev:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := getEnvInt("REDIS_DB", 0)

	mssqlServer := getEnv("MSSQL_SERVER", "daydb-svc.h1-db-dev")
	mssqlPort := getEnvInt("MSSQL_PORT", 1433)
	mssqlUser := getEnv("MSSQL_USER", "mobile_api")
	mssqlPassword := getEnv("MSSQL_PASSWORD", "")
	mssqlDatabase := getEnv("MSSQL_DATABASE", "HKNetGame_HJ")

	router := gin.Default()

	// --------------------------------------------------------
	// 靜態文件服務
	//    設定靜態文件目錄，訪問 http://localhost:8080/static/xxx
	// --------------------------------------------------------
	router.Static("/static", "./static")

	// 設定首頁為靜態 HTML (改用 c.File 避免與 Vue 的 {{}} 語法衝突)
	router.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// --------------------------------------------------------
	// 範例 2: 基本 POST 請求 - 會員登入
	//    取得參數: c.ShouldBindJSON(&struct)
	//    回應 JSON: c.JSON(http.StatusOK, gin.H{...})
	// --------------------------------------------------------
	router.POST("/api/Member/login", func(c *gin.Context) {
		// 可用以下方式取得參數:
		// var req struct {
		//     Username string `json:"username"`
		//     Password string `json:"password"`
		// }
		// c.ShouldBindJSON(&req)

		// 回應 JSON
		c.JSON(http.StatusOK, gin.H{
			"message": "login",
			"status":  "success",
		})
	})

	// --------------------------------------------------------
	// 範例 3: GET 請求 - 健康檢查 API
	//    展示如何檢查 Redis 與 MSSQL 連線狀態
	//    回應格式: JSON
	// --------------------------------------------------------
	router.GET("/healthcheck", func(c *gin.Context) {
		ctx := context.Background()
		result := gin.H{
			"status": "healthy",
			"redis":  "unknown",
			"mssql":  "unknown",
		}
		statusCode := http.StatusOK

		// ---- 檢查 Redis 連線 ----
		rdb := redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: redisPassword,
			DB:       redisDB,
		})

		_, err := rdb.Ping(ctx).Result()
		if err != nil {
			result["redis"] = "disconnected"
			result["status"] = "unhealthy"
			statusCode = http.StatusServiceUnavailable
			log.Printf("Redis error: %v\n", err)
		} else {
			result["redis"] = "connected"
		}
		defer rdb.Close()

		// ---- 檢查 MSSQL 連線 ----
		connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;",
			mssqlServer, mssqlUser, mssqlPassword, mssqlPort, mssqlDatabase)

		db, err := sql.Open("sqlserver", connString)
		if err != nil {
			result["mssql"] = "disconnected"
			result["status"] = "unhealthy"
			statusCode = http.StatusServiceUnavailable
			log.Printf("MSSQL connection error: %v\n", err)
		} else {
			defer db.Close()
			err = db.Ping()
			if err != nil {
				result["mssql"] = "disconnected"
				result["status"] = "unhealthy"
				statusCode = http.StatusServiceUnavailable
				log.Printf("MSSQL ping error: %v\n", err)
			} else {
				result["mssql"] = "connected"
			}
		}

		// ---- 回應 JSON ----
		c.JSON(statusCode, result)
	})

	// --------------------------------------------------------
	// 額外範例: RESTful API 路由分組
	// --------------------------------------------------------
	// api := router.Group("/api")
	// {
	//     api.GET("/users", func(c *gin.Context) { ... })       // 取得用戶列表
	//     api.GET("/users/:id", func(c *gin.Context) { ... })    // 取得單一用戶
	//     api.POST("/users", func(c *gin.Context) { ... })       // 建立用戶
	//     api.PUT("/users/:id", func(c *gin.Context) { ... })    // 更新用戶
	//     api.DELETE("/users/:id", func(c *gin.Context) { ... }) // 刪除用戶
	// }

	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	log.Printf("Server listening on %s\n", serverAddr)

	go func() {
		// 服務連線
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待中斷信號以優雅地關閉服務器（設置 5 秒的超時時間）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
