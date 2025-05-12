// handlers/pod_logs.go
package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"ml-platform-backend/pkg/k8sclient"
    "io"
    "strings"
	
	"k8s.io/client-go/tools/remotecommand"
	corev1 "k8s.io/api/core/v1"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var wsUpgrader = websocket.Upgrader{ CheckOrigin: func(r *http.Request) bool { return true } }

func PodLogStream(c *gin.Context) {
	podName := c.Param("podName")
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	req := k8sclient.K8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace("sumin").
		Name(podName).
		SubResource("log").
		VersionedParams(&corev1.PodLogOptions{
			Follow:    true,
			Container: "trainer",
			Timestamps: true,
		}, clientgoscheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(k8sclient.RestConfig, "POST", req.URL())
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("로그 연결 실패"))
		return
	}

	reader, writer := io.Pipe()
	defer writer.Close()

	go func() {
		_ = exec.Stream(remotecommand.StreamOptions{
			Stdout: writer,
			Stderr: writer,
		})
	}()

	buf := make([]byte, 4096)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			break
		}
		conn.WriteMessage(websocket.TextMessage, buf[:n])
	}
}

func GetPodLogs(c *gin.Context) {
    podName := c.Param("name")
    podLogOpts := corev1.PodLogOptions{
        Container:  "trainer",
        Timestamps: true,
    }

    req := k8sclient.K8sClient.CoreV1().Pods("sumin").GetLogs(podName, &podLogOpts)

    stream, err := req.Stream(c)
    if err != nil {
        c.String(http.StatusInternalServerError, "로그를 가져올 수 없습니다")
        return
    }
    defer stream.Close()

    buf := new(strings.Builder)
    io.Copy(buf, stream)

    c.String(http.StatusOK, buf.String())
}
