package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
	"websocket-ssh-server/config"
	views "websocket-ssh-server/controller"

	"github.com/gin-gonic/gin"
)

func init() {
	//设置Flats为 日期 时间 微秒 文件名:行号
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}

func main() {
	log.Println("start...")
	var configFilePath string
	flag.StringVar(&configFilePath, "c", "config.yml", "Configuration file path")
	flag.Parse()

	// 读取配置
	config.ReadFile(configFilePath)

	host := config.CONFIG.App.Host
	port := strconv.Itoa(config.CONFIG.App.Port)

	gin.SetMode(gin.ReleaseMode)
	server := gin.New()
	server.Use(gin.Recovery())
	server.Use(JSONAppErrorReporter())
	server.Use(CORSMiddleware())
	server.GET("/wsss/ws", views.ShellWs)
	server.GET("/wsss/socket", views.SshdSocket)
	server.Run(host + ":" + port)
}

// 对产生的任何error进行处理
func JSONAppErrorReporter() gin.HandlerFunc {
	return jsonAppErrorReporterT(gin.ErrorTypeAny)
}

func jsonAppErrorReporterT(errType gin.ErrorType) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		detectedErrors := c.Errors.ByType(errType)
		if len(detectedErrors) > 0 {
			err := detectedErrors[0].Err
			var parsedError *views.ApiError
			switch err.(type) {
			//如果产生的error是自定义的结构体,转换error,返回自定义的code和msg
			case *views.ApiError:
				parsedError = err.(*views.ApiError)
			default:
				parsedError = &views.ApiError{
					Code:    http.StatusInternalServerError,
					Message: err.Error(),
				}
			}
			c.IndentedJSON(parsedError.Code, parsedError)
			return
		}

	}
}

// 设置所有跨域请求
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
