// handlers/get_mljobs.go
package handlers

import (
    "context"
    "net/http"

    "github.com/gin-gonic/gin"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

    "ml-platform-backend/pkg/k8sclient"
)

func GetMLJobs(c *gin.Context) {
    ns := c.DefaultQuery("namespace", "sumin")

    list, err := k8sclient.DynamicClient.
        Resource(MLJobGVR).
        Namespace(ns).
        List(context.Background(), metav1.ListOptions{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, list.Items)
}
