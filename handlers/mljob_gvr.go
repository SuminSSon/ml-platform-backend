// handlers/mljob_gvr.go
package handlers

import "k8s.io/apimachinery/pkg/runtime/schema"

var MLJobGVR = schema.GroupVersionResource{
    Group:    "ai.mljob-controller",
    Version:  "v1",
    Resource: "mljobs",
}
