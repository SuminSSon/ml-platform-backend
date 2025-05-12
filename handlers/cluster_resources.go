// handlers/cluster_resources.go
package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

    "ml-platform-backend/pkg/k8sclient"
)

func GetClusterResources(c *gin.Context) {
    nodes, err := k8sclient.K8sClient.CoreV1().Nodes().List(c, metav1.ListOptions{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    totalCPU := corev1.ResourceList{}
    totalMem := corev1.ResourceList{}
    allocCPU := corev1.ResourceList{}
    allocMem := corev1.ResourceList{}

    for _, n := range nodes.Items {
        for k, v := range n.Status.Capacity {
            switch k {
            case corev1.ResourceCPU:
                q := totalCPU[k]
                q.Add(v)
                totalCPU[k] = q
            case corev1.ResourceMemory:
                q := totalMem[k]
                q.Add(v)
                totalMem[k] = q
            }
        }
        for k, v := range n.Status.Allocatable {
            switch k {
            case corev1.ResourceCPU:
                q := allocCPU[k]
                q.Add(v)
                allocCPU[k] = q
            case corev1.ResourceMemory:
                q := allocMem[k]
                q.Add(v)
                allocMem[k] = q
            }
        }
    }

    cpuTotal := totalCPU[corev1.ResourceCPU]
    memTotal := totalMem[corev1.ResourceMemory]
    cpuAlloc := allocCPU[corev1.ResourceCPU]
    memAlloc := allocMem[corev1.ResourceMemory]

    c.JSON(http.StatusOK, gin.H{
        "totalCPU":       cpuTotal.String(),
        "totalMemory":    memTotal.String(),
        "allocatableCPU": cpuAlloc.String(),
        "allocatableMem": memAlloc.String(),
    })
}
