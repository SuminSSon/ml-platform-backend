// main.go
package main

import (
	"ml-platform-backend/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	api := r.Group("/api")
	{
		api.GET("/mljobs", handlers.GetMLJobs)
		api.POST("/mljobs", handlers.CreateMLJob)
		api.PATCH("/mljobs/:jobName", handlers.PatchMLJob)

		api.GET("/pods", handlers.GetPods)
		api.POST("/pods", handlers.CreatePod)
		api.POST("/pods/:name/stop", handlers.StopPod)
		api.GET("/pods/:name/logs", handlers.PodLogStream)

		api.GET("/cluster/resources", handlers.GetClusterResources)

		// WebSocket
		api.GET("/ws/pods", handlers.PodStream)
		api.GET("/ws/logs/:podName", handlers.PodLogStream)

        api.GET("/pods/stream", handlers.PodStream)
	}

	r.Run(":8080")
}
