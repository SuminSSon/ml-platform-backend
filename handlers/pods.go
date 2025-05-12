// handlers/pods.go
package handlers

import (
    "net/http"
    "time"
    "fmt"

    "github.com/gin-gonic/gin"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

    "ml-platform-backend/pkg/k8sclient"
)

func GetPods(c *gin.Context) {
	ns := c.DefaultQuery("namespace", "sumin")

	podList, err := k8sclient.K8sClient.CoreV1().Pods(ns).List(c, metav1.ListOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var out []gin.H
	for _, p := range podList.Items {
		var cpu, mem, gpu string
		if len(p.Spec.Containers) > 0 {
			req := p.Spec.Containers[0].Resources.Requests
			if q, ok := req[corev1.ResourceCPU]; ok {
				cpu = q.String()
			}
			if q, ok := req[corev1.ResourceMemory]; ok {
				mem = q.String()
			}
			if q, ok := req[corev1.ResourceName("nvidia.com/gpu")]; ok {
				gpu = q.String()
			}
		}

		restarts := 0
		if len(p.Status.ContainerStatuses) > 0 {
			restarts = int(p.Status.ContainerStatuses[0].RestartCount)
		}

        //age := time.Since(p.CreationTimestamp.Time).String()
        duration := time.Since(p.CreationTimestamp.Time)
        minutes := int(duration.Minutes())
        seconds := int(duration.Seconds()) % 60
        ageStr := fmt.Sprintf("%dm %ds", minutes, seconds)


		out = append(out, gin.H{
			"name":     p.Name,
			"status":   string(p.Status.Phase),
			"cpu":      cpu,
			"gpu":      gpu,
			"memory":   mem,
			"restarts": restarts,
			"age":      ageStr,
		})
	}

	c.JSON(http.StatusOK, out)
}