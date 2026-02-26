package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"

	"database/sql"
	"fmt"

	_ "github.com/denisenkom/go-mssqldb"

	"github.com/go-redis/redis/v8"
)

// ============================================================
// Gin API 開發教學
// ============================================================
//
// 1. 建立 Router
//    router := gin.Default()  // 建立預設路由器（包含 Logger 和 Recovery 中間件）
//
// 2. 定義路由
//    router.METHOD("/path", handler)
//
//    METHOD 可用:
//    - GET    : 取得資源
//    - POST   : 新增資源
//    - PUT    : 更新資源（完整）
//    - PATCH  : 更新資源（部分）
//    - DELETE : 刪除資源
//
// 3. Handler 函數參數
//    func(c *gin.Context)
//    - c: Gin 上下文物件，包含請求與回應資訊
//
// 4. 常用回應方法
//    c.String(http.StatusOK, "訊息")           // 回傳文字
//    c.JSON(http.StatusOK, gin.H{"key": "value"}) // 回傳 JSON
//    c.XML(http.StatusOK, gin.H{"key": "value"})  // 回傳 XML
//    c.HTML(htmlTemplate, "data")              // 回傳 HTML
//
// 5. 取得請求參數
//    c.Query("name")           // GET 參數 ?name=xxx
//    c.PostForm("name")        // POST 表單參數
//    c.Param("id")             // URL 參數 /user/:id
//    c.ShouldBindJSON(&obj)    // JSON Body 綁定
//
// 6. 回應狀態碼
//    http.StatusOK         = 200
//    http.StatusCreated    = 201
//    http.StatusBadRequest = 400
//    http.StatusUnauthorized = 401
//    http.StatusForbidden  = 403
//    http.StatusNotFound   = 404
//    http.StatusInternalServerError = 500
//
// ============================================================

func main() {

	router := gin.Default()

	// 載入 HTML 模板
	router.LoadHTMLGlob("static/*")

	// --------------------------------------------------------
	// 靜態文件服務
	//    設定靜態文件目錄，訪問 http://localhost:8080/static/xxx
	// --------------------------------------------------------
	router.Static("/static", "./static")

	// 設定首頁為靜態 HTML
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
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
	// 範例 3: GET 請求 + Redis + MSSQL 連線測試
	//    展示如何在 API 中使用資料庫
	// --------------------------------------------------------
	router.GET("/healthcheck", func(c *gin.Context) {
		// ---- Redis 連線範例 ----
		rdb := redis.NewClient(&redis.Options{
			Addr:     "redis-cluster.h1-redis-dev:6379",
			Password: "h1devredis1688", // no password set
			DB:       0,                // use default DB
		})

		ctx := context.Background()

		val, err := rdb.Ping(ctx).Result()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(val)
		}

		// ---- MSSQL 連線範例 ----
		connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;",
			"daydb-svc.h1-db-dev", "mobile_api", "a:oY%~^E+VU0", 1433, "HKNetGame_HJ")

		db, err := sql.Open("sqlserver", connString)
		if err != nil {
			log.Fatal("Open connection failed:", err.Error())
		}
		defer db.Close()

		rows, err := db.Query("SELECT name, state_desc FROM sys.databases WHERE name = 'HKNetGame_HJ';")
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		defer rows.Close()

		var name, state_desc string
		for rows.Next() {
			err := rows.Scan(&name, &state_desc)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("name: %s, state_desc: %s\n", name, state_desc)
		}
		str := fmt.Sprintf("val: %v, name: %s, state_desc: %s", val, name, state_desc)
		c.String(http.StatusOK, str)
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
		Addr:    ":8080",
		Handler: router,
	}

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
