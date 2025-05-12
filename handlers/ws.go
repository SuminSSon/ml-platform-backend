// handlers/ws.go
package handlers

import (
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	//corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	
	"ml-platform-backend/pkg/k8sclient"
)

var upgrader = websocket.Upgrader{ CheckOrigin: func(r *http.Request) bool { return true } }

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

		conn.WriteJSON(pods.Items)
		time.Sleep(3 * time.Second)
	}
}
