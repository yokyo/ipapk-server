package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/phinexdaz/ipapk-server/conf"
	"github.com/phinexdaz/ipapk-server/api"
	"github.com/phinexdaz/ipapk-server/models"
	"github.com/phinexdaz/ipapk-server/templates"
	"github.com/phinexdaz/ipapk-server/utils"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func Init() {
	_, err := os.Stat(".data")
	if os.IsNotExist(err) {
		os.MkdirAll(".data", 0755)
	}

	if err := utils.InitCA(); err != nil {
		log.Fatal(err)
	}

	if err := conf.InitConfig("config.json"); err != nil {
		log.Fatal(err)
	}

	if err := models.InitDB(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	Init()

	router := gin.Default()
	router.SetFuncMap(templates.TplFuncMap)
	router.LoadHTMLGlob("public/views/*")

	router.Static("app", ".data/app")
	router.Static("icon", ".data/icon")
	router.Static("static", "public/static")
	router.StaticFile("myCA.cer", ".ca/myCA.cer")

	router.POST("/upload", api.Upload)
	bundle := router.Group("/bundle")
	{
		bundle.GET("/:uuid", api.GetBundle)
		bundle.GET("/:uuid/changelog", api.GetChangelog)
		bundle.GET("/:uuid/qrcode", api.GetQRCode)
		bundle.GET("/:uuid/plist", api.GetPlist)
		bundle.GET("/:uuid/download", api.DownloadAPP)
		bundle.GET("/:uuid/versions", api.GetVersions)
		bundle.GET("/:uuid/versions/:version", api.GetBuilds)
	}

	srv := &http.Server{
		Addr:    conf.AppConfig.Addr(),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServeTLS(".ca/mycert1.cer", ".ca/mycert1.key"); err != nil {
			log.Printf("listen: %v\n", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
}
