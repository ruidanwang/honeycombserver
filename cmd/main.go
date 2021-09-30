package main

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/syndtr/goleveldb/leveldb"
	"honeyserver/pkg/websocket"
	"net/http"
)

func main() {

	r := gin.Default()

	//以下配置gin跨域配置
	config := cors.DefaultConfig()
	config.ExposeHeaders = []string{"Authorization"}
	config.AllowCredentials = true
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	r.Use(gzip.Gzip(gzip.DefaultCompression))
	//ws pool
	pool := websocket.NewPool()
	go pool.Start()

	db,_ := leveldb.OpenFile("./data", nil)
	initFile,_:=db.Get([]byte("key"),nil)
	if(initFile == nil){
		initdb(db)
	}

	r.GET("/ws", func(c *gin.Context) {
		username:= c.Query("name");
		fmt.Println(username)
		serveWs(pool, c,username)
	})

	r.POST("/get", func(c *gin.Context) {
		username:= c.Query("name");
		fmt.Println(username)
		value,_ := db.Get([]byte("key"),nil)
		c.JSON(http.StatusOK,string(value))
		//c.JSON(http.StatusOK,`[{"name":"Cell","index":"sheet_01","order":0,"status":1,"celldata":[{"r":0,"c":0,"v":{"v":1,"m":"1","ct":{"fa":"General","t":"n"}}}]},{"name":"Data","index":"sheet_02","order":1,"status":0},{"name":"Picture","index":"sheet_03","order":2,"status":0}]`);
	})

	r.POST("set", func(c *gin.Context) {
		value := c.PostForm("jsonExcel")
		db.Put([]byte("key"),[]byte(value),nil)
		c.String(http.StatusOK, "true")
	})



	r.Run("localhost:9000")
}

func initdb(db *leveldb.DB) {
	db.Put([]byte("key"),[]byte(`[{"name":"Cell","index":"sheet_01","order":0,"status":1,"celldata":[{"r":0,"c":0,"v":{"v":1,"m":"1","ct":{"fa":"General","t":"n"}}}]},{"name":"Data","index":"sheet_02","order":1,"status":0},{"name":"Picture","index":"sheet_03","order":2,"status":0}]`),nil)
	//return
}


//func serveWs(pool *websocket.Pool, w http.ResponseWriter, r *http.Request) {
func serveWs(pool *websocket.Pool, c *gin.Context,username string) {
	fmt.Println("WebSocket Endpoint Hit")
	conn, _ := websocket.Upgrade(c.Writer, c.Request)


	client := &websocket.Client{
		Conn: conn,
		Pool: pool,
		ID: username,
	}

	pool.Register <- client
	client.Read()
}


