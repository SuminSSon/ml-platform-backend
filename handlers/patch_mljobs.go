package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ml-platform-backend/pkg/k8sclient"
)

type PatchReq struct {
	CPU    string `json:"cpu"`
	GPU    string `json:"gpu"`
	Memory string `json:"memory"`
}

func PatchMLJob(c *gin.Context) {
	ns := c.DefaultQuery("namespace", "sumin")
	name := c.Param("jobName")

	var req PatchReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resources := make(map[string]map[string]string)

	if req.CPU != "" || req.Memory != "" || req.GPU != "" {
		requests := map[string]string{}
		limits := map[string]string{}

		if req.CPU != "" {
			requests["cpu"] = req.CPU
			limits["cpu"] = req.CPU
		}
		if req.Memory != "" {
			requests["memory"] = req.Memory
			limits["memory"] = req.Memory
		}
		if req.GPU != "" {
			requests["nvidia.com/gpu"] = req.GPU
			limits["nvidia.com/gpu"] = req.GPU
		}

		resources["requests"] = requests
		resources["limits"] = limits
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "변경할 자원이 없습니다"})
		return
	}

	patchBody := map[string]interface{}{
		"spec": map[string]interface{}{
			"resources": resources,
		},
	}

	jsonPatch, err := json.Marshal(patchBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "패치 생성 실패"})
		return
	}

	result, err := k8sclient.DynamicClient.
		Resource(MLJobGVR).
		Namespace(ns).
		Patch(context.Background(),
			name,
			types.MergePatchType,
			jsonPatch,
			metav1.PatchOptions{})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "패치 실패: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, result.Object)
}
