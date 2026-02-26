# web.go
```markdown
使用 Go 建立 Web 與靜態前端
```

# 初始化
```
go mod init github.com/chiisen/web.go
```

**為什麼 `go mod init` 後面要加 `github.com/chiisen/web.go`？**

在 `go mod init github.com/chiisen/web.go` 這個指令中，`github.com/chiisen/web.go` 被稱為這個 Go 專案的 **「模組路徑」 (Module Path)**。

加上 `github.com/chiisen` 主要有以下幾個重要意義與目的：

1. **唯一識別碼 (Namespace) 防止命名衝突**：
   加上網域和你的使用者名稱作為前綴，建立了一個獨一無二的命名空間，確保你的套件與全世界其他人的套件不會發生名稱衝突。

2. **套件下載定位 (Location for `go get`)**：
   Go 語言的套件管理與網址是直接綁定的。當其他開發者想要使用這個套件時，只要執行 `go get github.com/chiisen/web.go`，Go 編譯器就會直接根據這個路徑，去 GitHub 尋找並下載程式碼。

3. **專案內部的 Import 路徑基準**：
   專案內部互相引用程式碼時，也必須以這個完整路徑為開頭。例如：`import "github.com/chiisen/web.go/utils"`。

*💡 最佳實踐：在實際開發中，強烈建議使用 `[程式碼代管平台]/[你的帳號]/[專案名稱]` 的格式來命名，這符合 Go 社群的標準規範。*

# 執行
```
go run main.go
```
# 瀏覽
http://localhost:8080/

# 靜態前端
靜態檔案放在 `static/` 目錄下，前端畫面使用了 **Vue 3 (CDN 版本)** 來處理資料與互動：
- 進入點為 `static/index.html`。
- 透過 Vue 的 Composition API (`setup`) 將動態資料（如 API 列表）綁定到畫面上。

# API 文件

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | 歡迎訊息 |
| POST | `/api/Member/login` | 會員登入 |
| GET | `/healthcheck` | 服務健康檢查 (測試 Redis 與 MSSQL 連線) |


