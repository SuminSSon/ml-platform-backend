package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	//"k8s.io/apimachinery/pkg/runtime/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"ml-platform-backend/pkg/k8sclient"
)

type CreateMLJobRequest struct {
	CPU    string `json:"cpu" binding:"required"`
	GPU    string `json:"gpu"`
	Memory string `json:"memory" binding:"required"`
}

func CreateMLJob(c *gin.Context) {
	ns := c.DefaultQuery("namespace", "sumin")

	var req CreateMLJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mljob := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "ai.mljob-controller/v1",
			"kind":       "MLJob",
			"metadata": map[string]interface{}{
				"generateName": "mljob-",
				"namespace":    ns,
			},
			"spec": map[string]interface{}{
				"resources": map[string]interface{}{
					"requests": map[string]interface{}{
						"cpu":             req.CPU,
						"memory":          req.Memory,
						"nvidia.com/gpu":  req.GPU,
					},
					"limits": map[string]interface{}{
						"cpu":             req.CPU,
						"memory":          req.Memory,
						"nvidia.com/gpu":  req.GPU,
					},
				},
			},
		},
	}

	result, err := k8sclient.DynamicClient.
		Resource(MLJobGVR).
		Namespace(ns).
		Create(context.Background(), mljob, metav1.CreateOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result.Object)
}