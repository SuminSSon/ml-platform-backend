package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ml-platform-backend/pkg/k8sclient"
)

/*───────────────────────────────*/
/* WebSocket 업그레이더 설정      */
/*───────────────────────────────*/
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

/*───────────────────────────────*/
/* Prometheus 쿼리 헬퍼 함수      */
/*───────────────────────────────*/
const prometheusURL = "http://prometheus.monitoring.svc.cluster.local:9090/api/v1/query"

// 단일 PromQL 쿼리를 실행해 float64 결과 한 개를 리턴
func queryPrometheus(query string) (float64, error) {
	// Prometheus HTTP API 호출
	resp, err := http.Get(prometheusURL + "?query=" + query)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	// 응답 구조체
	var result struct {
		Data struct {
			Result []struct {
				Value []interface{} `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}
	if len(result.Data.Result) == 0 {
		return 0, nil // 데이터 없음
	}

	valueStr := result.Data.Result[0].Value[1].(string)
	var value float64
	fmt.Sscanf(valueStr, "%f", &value)
	return value, nil
}

/*───────────────────────────────*/
/* 실제 WebSocket 핸들러          */
/*───────────────────────────────*/
func PodStream(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ns := "sumin"

	for {
		/* 1) 쿠버네티스에서 Pod 목록 조회 */
		pods, err := k8sclient.K8sClient.CoreV1().Pods(ns).List(c, metav1.ListOptions{})
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"cannot fetch pods"}`))
			time.Sleep(time.Second)
			continue
		}

		var out []map[string]interface{}

		for _, p := range pods.Items {
			/* 2) 요청/제한치(스펙) 파싱 */
			var cpuReq, memReq, gpuReq string
			if len(p.Spec.Containers) > 0 {
				req := p.Spec.Containers[0].Resources.Requests
				if q, ok := req[corev1.ResourceCPU]; ok {
					cpuReq = q.String()
				}
				if q, ok := req[corev1.ResourceMemory]; ok {
					memReq = q.String()
				}
				if q, ok := req[corev1.ResourceName("nvidia.com/gpu")]; ok {
					gpuReq = q.String()
				}
			}

			/* 3) 실시간 사용량 Prometheus 쿼리 */
			podName := p.Name
			cpuQ := fmt.Sprintf(
				`sum(rate(container_cpu_usage_seconds_total{namespace="%s",pod="%s",image!~"pause"}[1m]))*1000`,
				ns, podName)
			memQ := fmt.Sprintf(
				`sum(container_memory_usage_bytes{namespace="%s",pod="%s"})`,
				ns, podName)

			cpuUsed, _ := queryPrometheus(cpuQ)   // mCPU
			memUsed, _ := queryPrometheus(memQ)   // bytes

			/* 4) 부가 정보 */
			restarts := 0
			if len(p.Status.ContainerStatuses) > 0 {
				restarts = int(p.Status.ContainerStatuses[0].RestartCount)
			}
			ageDur := time.Since(p.CreationTimestamp.Time)
			ageStr := fmt.Sprintf("%dm %ds", int(ageDur.Minutes()), int(ageDur.Seconds())%60)

			/* 5) JSON 행 추가 */
			out = append(out, map[string]interface{}{
				"name":       podName,
				"status":     string(p.Status.Phase),
				"cpuRequest": cpuReq,
				"memRequest": memReq,
				"gpuRequest": gpuReq,
				"cpuUsage":   fmt.Sprintf("%.1fm", cpuUsed),
				"memUsage":   fmt.Sprintf("%.1fMi", memUsed/1024/1024),
				"restarts":   restarts,
				"age":        ageStr,
				"job":        p.Labels["job"],
			})
		}

		/* 6) WebSocket으로 결과 전송 */
		conn.WriteJSON(out)
		time.Sleep(1 * time.Second)
	}
}