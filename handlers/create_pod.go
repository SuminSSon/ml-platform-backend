package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ml-platform-backend/pkg/k8sclient"
)

func CreatePod(c *gin.Context) {
    ns := "sumin" // 고정 네임스페이스

    // 파일이 존재하는지 확인
    file, _ := c.FormFile("file")

    // 파일이 없으면 오류 처리
    if file == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
        fmt.Println("파일이 없습니다. 요청이 잘못되었습니다.")
        return
    }

    // 파일을 서버의 /tmp 디렉터리에 저장
    dst := fmt.Sprintf("/tmp/%s", file.Filename)
    if err := c.SaveUploadedFile(file, dst); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "file upload failed"})
        fmt.Println("파일 업로드 실패:", err)
        return
    }

    fmt.Println("파일 업로드 성공:", file.Filename)

    // 파일을 파드에 마운트
    fileVolume := &corev1.Volume{
        Name: "data",
        VolumeSource: corev1.VolumeSource{
            EmptyDir: &corev1.EmptyDirVolumeSource{},
        },
    }
    volumeMount := &corev1.VolumeMount{
        Name:      "data",
        MountPath: "/mnt/data", // 파드 내에서 이 경로에 파일을 마운트
    }

    // 파드 정의
    pod := &corev1.Pod{
        ObjectMeta: v1.ObjectMeta{
            Name:      "dynamic-pod", // 파드 이름은 필요하므로 임시로 'dynamic-pod'로 설정
            Namespace: ns,
            Labels: map[string]string{
                "job": "manual",
            },
        },
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{
                {
                    Name:  "trainer",
                    Image: "ubuntu:latest", // 사용할 기본 이미지
                    Command: []string{"sleep", "3600"}, // 파드가 종료되지 않도록 'sleep' 명령어 실행
                    VolumeMounts: []corev1.VolumeMount{*volumeMount}, // 파일을 파드 내에서 마운트
                },
            },
            Volumes: []corev1.Volume{*fileVolume}, // 파드 내 볼륨 설정
        },
    }

    // Kubernetes 클러스터에 파드 생성
    _, err := k8sclient.K8sClient.CoreV1().Pods(ns).Create(context.Background(), pod, v1.CreateOptions{})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        fmt.Println("파드 생성 실패:", err)
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "pod created"})
    fmt.Println("파드 생성 성공")
}
