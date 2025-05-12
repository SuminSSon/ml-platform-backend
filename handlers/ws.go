package handlers

import (
	"time"
	"net/http"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ml-platform-backend/pkg/k8sclient"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func PodStream(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		pods, err := k8sclient.K8sClient.CoreV1().Pods("sumin").List(c, metav1.ListOptions{})
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"cannot fetch pods"}`))
			continue
		}

		var result []map[string]interface{}

		for _, p := range pods.Items {
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

			//age := time.Since(p.CreationTimestamp.Time).Round(time.Minute).String()
			duration := time.Since(p.CreationTimestamp.Time)
			minutes := int(duration.Minutes())
			seconds := int(duration.Seconds()) % 60
			ageStr := fmt.Sprintf("%dm %ds", minutes, seconds)

			result = append(result, map[string]interface{}{
				"name":     p.Name,
				"status":   string(p.Status.Phase),
				"cpu":      cpu,
				"gpu":      gpu,
				"memory":   mem,
				"restarts": restarts,
				"age":      ageStr,
				"job":      p.Labels["job"],
			})
		}

		conn.WriteJSON(result)
		time.Sleep(1 * time.Second)
	}
}