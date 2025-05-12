package handlers

import (
	"context"
	"fmt"
	// "mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ml-platform-backend/pkg/k8sclient"
)

func CreatePod(c *gin.Context) {
	ns := "sumin" // 고정 네임스페이스

	// 폼 파라미터 읽기
	name := c.PostForm("name")
	// requirements := c.PostForm("requirements")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	// 파일이 있으면 처리 (예: ConfigMap 생성 등)
	file, _ := c.FormFile("file")
	var fileVolume *corev1.Volume
	var volumeMount *corev1.VolumeMount

	if file != nil {
		dst := fmt.Sprintf("/tmp/%s", file.Filename)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "file upload failed"})
			return
		}
		// 실제 사용 시 PVC 또는 ConfigMap으로 볼륨 구성 필요
		// 여기서는 단순 시뮬레이션
		fileVolume = &corev1.Volume{
			Name: "data",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		}
		volumeMount = &corev1.VolumeMount{
			Name:      "data",
			MountPath: "/mnt/data",
		}
	}

	pod := &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels: map[string]string{
				"job": "manual",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "trainer",
					Image: "ubuntu:latest", // 또는 사용자 정의 이미지
					Command: []string{"sleep", "3600"},
					Resources: corev1.ResourceRequirements{
						// 요구 사항 파싱해서 여기에 추가 가능
					},
				},
			},
		},
	}

	if fileVolume != nil && volumeMount != nil {
		pod.Spec.Volumes = []corev1.Volume{*fileVolume}
		pod.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{*volumeMount}
	}

	_, err := k8sclient.K8sClient.CoreV1().Pods(ns).Create(context.Background(), pod, v1.CreateOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "pod created"})
}
