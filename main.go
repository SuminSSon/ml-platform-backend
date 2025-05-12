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
        api.PATCH("/mljobs/:jobName", handlers.PatchMLJob)
        api.GET("/pods", handlers.GetPods)
        api.GET("/pods/:podName/logs", handlers.GetPodLogs)
        api.GET("/cluster/resources", handlers.GetClusterResources)

        api.GET("/ws/pods", handlers.PodStream)
        api.GET("/ws/logs/:podName", handlers.PodLogStream)
    }

    r.Run(":8080")
}

