2026-07-02 20:34:36	deployments/app	2026-07-02 12:34:36	INFO	router/router.go:25	incoming request	{"method": "POST", "path": "/api/v1/auth/register", "remote_addr": "172.18.0.1"}
2026-07-02 20:34:36	deployments/postgres	2026-07-02 12:34:36.091 UTC [41] ERROR:  relation "users" does not exist at character 15
2026-07-02 20:34:36	deployments/postgres	2026-07-02 12:34:36.091 UTC [41] STATEMENT:  SELECT * FROM "users" WHERE email = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2
2026-07-02 20:34:36	deployments/app	2026-07-02 12:34:36	ERROR	auth/register.go:32	failed to check email uniqueness	{"error": "failed to get user by email: ERROR: relation \"users\" does not exist (SQLSTATE 42P01)"}
2026-07-02 20:34:36	deployments/app	github.com/EricStone1900/ecommerce-backend/internal/usecase/auth.(*AuthUseCase).Register
2026-07-02 20:34:36	deployments/app		/build/internal/usecase/auth/register.go:32
2026-07-02 20:34:36	deployments/app	github.com/EricStone1900/ecommerce-backend/internal/interface/http/handler.(*AuthHandler).Register
2026-07-02 20:34:36	deployments/app		/build/internal/interface/http/handler/auth.go:42
2026-07-02 20:34:36	deployments/app	github.com/gin-gonic/gin.(*Context).Next
2026-07-02 20:34:36	deployments/app		/go/pkg/mod/github.com/gin-gonic/gin@v1.10.0/context.go:185
2026-07-02 20:34:36	deployments/app	github.com/EricStone1900/ecommerce-backend/internal/interface/http/router.NewRouter.func1
2026-07-02 20:34:36	deployments/app		/build/internal/interface/http/router/router.go:30
2026-07-02 20:34:36	deployments/app	github.com/gin-gonic/gin.(*Context).Next
2026-07-02 20:34:36	deployments/app		/go/pkg/mod/github.com/gin-gonic/gin@v1.10.0/context.go:185
2026-07-02 20:34:36	deployments/app	github.com/gin-gonic/gin.CustomRecoveryWithWriter.func1
2026-07-02 20:34:36	deployments/app		/go/pkg/mod/github.com/gin-gonic/gin@v1.10.0/recovery.go:102
2026-07-02 20:34:36	deployments/app	github.com/gin-gonic/gin.(*Context).Next
2026-07-02 20:34:36	deployments/app		/go/pkg/mod/github.com/gin-gonic/gin@v1.10.0/context.go:185
2026-07-02 20:34:36	deployments/app	github.com/gin-gonic/gin.(*Engine).handleHTTPRequest
2026-07-02 20:34:36	deployments/app		/go/pkg/mod/github.com/gin-gonic/gin@v1.10.0/gin.go:633
2026-07-02 20:34:36	deployments/app	github.com/gin-gonic/gin.(*Engine).ServeHTTP
2026-07-02 20:34:36	deployments/app		/go/pkg/mod/github.com/gin-gonic/gin@v1.10.0/gin.go:589
2026-07-02 20:34:36	deployments/app	net/http.serverHandler.ServeHTTP
2026-07-02 20:34:36	deployments/app		/usr/local/go/src/net/http/server.go:3210
2026-07-02 20:34:36	deployments/app	net/http.(*conn).serve
2026-07-02 20:34:36	deployments/app		/usr/local/go/src/net/http/server.go:2092
2026-07-02 20:34:36	deployments/app	
2026-07-02 20:34:36	deployments/app	2026/07/02 12:34:36 /build/internal/infrastructure/persistence/gorm/user_repo.go:82 ERROR: relation "users" does not exist (SQLSTATE 42P01)
2026-07-02 20:34:36	deployments/app	[0.566ms] [rows:0] SELECT * FROM "users" WHERE email = 'usertest@qq.com' AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT 1