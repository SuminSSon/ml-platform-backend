package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"ml-platform-backend/pkg/k8sclient"
)

func StopPod(c *gin.Context) {
	ns := "sumin"
	name := c.Param("name")

	err := k8sclient.K8sClient.CoreV1().Pods(ns).Delete(context.Background(), name, v1.DeleteOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "파드를 중지할 수 없습니다: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "파드 중지됨", "name": name})
}