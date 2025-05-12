package handlers

import (
	"bufio"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ml-platform-backend/pkg/k8sclient"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func int64Ptr(i int64) *int64 { return &i }

// WebSocket /api/ws/logs/:podName
func PodLogStream(c *gin.Context) {
	podName := c.Param("podName")

	// 컨테이너 이름 확인
	pod, err := k8sclient.K8sClient.CoreV1().Pods("sumin").Get(c, podName, metav1.GetOptions{})
	if err != nil || len(pod.Spec.Containers) == 0 {
		log.Printf("파드 정보 조회 실패: %v\n", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	containerName := pod.Spec.Containers[0].Name

	// WebSocket 업그레이드
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket 업그레이드 실패:", err)
		return
	}
	defer conn.Close()

	log.Printf("WebSocket 로그 스트리밍 시작: %s (%s)\n", podName, containerName)

	// 이전 로그 전송
	req1 := k8sclient.K8sClient.CoreV1().
		Pods("sumin").
		GetLogs(podName, &corev1.PodLogOptions{
			Container:  containerName,
			Timestamps: true,
			TailLines:  int64Ptr(100),
		})

	stream1, err := req1.Stream(c)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("이전 로그를 불러오지 못했습니다"))
		log.Println("이전 로그 불러오기 실패:", err)
	} else {
		scanner := bufio.NewScanner(stream1)
		for scanner.Scan() {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(scanner.Text())); err != nil {
				log.Println("이전 로그 WebSocket 전송 실패:", err)
				break
			}
		}
		stream1.Close()
	}

	// 실시간 로그 스트리밍
	stream2, err := k8sclient.K8sClient.CoreV1().
		Pods("sumin").
		GetLogs(podName, &corev1.PodLogOptions{
			Container:  containerName,
			Timestamps: true,
			Follow:     true,
		}).Stream(c)

	if err != nil {
		log.Println("실시간 로그 스트림 실패:", err)
		conn.WriteMessage(websocket.TextMessage, []byte("실시간 로그를 가져올 수 없습니다"))
		return
	}
	defer stream2.Close()

	scanner := bufio.NewScanner(stream2)
	for scanner.Scan() {
		line := scanner.Text()
		if err := conn.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
			log.Println("WebSocket 전송 실패:", err)
			break
		}
	}
	if err := scanner.Err(); err != nil {
		log.Println("실시간 로그 스캐너 오류:", err)
	}

	log.Printf("WebSocket 로그 스트리밍 종료: %s\n", podName)
}