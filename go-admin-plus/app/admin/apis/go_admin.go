package apis

import "github.com/gin-gonic/gin"

const INDEX = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Go Admin Plus API</title>
  <style>
    body { margin: 0; font: 16px/1.6 system-ui, sans-serif; color: #1f2937; background: #f8fafc; }
    main { max-width: 720px; margin: 12vh auto; padding: 0 24px; }
    h1 { margin: 0 0 8px; font-size: 32px; }
    a { color: #0969da; }
  </style>
</head>
<body>
  <main>
    <h1>Go Admin Plus API</h1>
    <p>服务已启动。</p>
    <p><a href="/swagger/admin/index.html">Swagger</a> · <a href="/health/ready">就绪检查</a></p>
  </main>
</body>
</html>`

func GoAdmin(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, INDEX)
}
