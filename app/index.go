package app

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joshDamian/web3storage-service/app/providers"
	"github.com/gin-contrib/cors"
)

func App() *gin.Engine {
	providers.Uploader.APIKey = os.Getenv("MORALIS_API_KEY")
	router := gin.Default()
	router.Use(cors.Default())
	router.POST("/upload-file", providers.UploadSingleFile)
	router.POST("/upload-files", providers.UploadMultipleFiles)
	return router
}
