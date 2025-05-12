// handlers/patch_mljobs.go
package handlers

import (
    "context"
    "fmt"
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

    patch := fmt.Sprintf(`{
  "spec": {
    "resources": {
      "requests": {
        "cpu": "%s",
        "memory": "%s",
        "nvidia.com/gpu": "%s"
      },
      "limits": {
        "cpu": "%s",
        "memory": "%s",
        "nvidia.com/gpu": "%s"
      }
    }
  }
}`, req.CPU, req.Memory, req.GPU, req.CPU, req.Memory, req.GPU)

    out, err := k8sclient.DynamicClient.
        Resource(MLJobGVR).
        Namespace(ns).
        Patch(context.Background(),
            name,
            types.MergePatchType,
            []byte(patch),
            metav1.PatchOptions{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, out.Object)
}
