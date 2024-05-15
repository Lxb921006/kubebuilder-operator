package v2

import (
	appV1 "github.com/Lxb921006/kubebuild-go/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
	"strconv"
)

func (src *App) ConvertTo(dstRaw conversion.Hub) error {
	//fmt.Println("ConvertTo  >>>> v2 -> v1")
	dst := dstRaw.(*appV1.App)
	dst.Spec.Replicas = src.Spec.Replicas
	dst.Spec.Image = src.Spec.Image
	dst.Spec.EnableService = src.Spec.EnableService
	dst.Spec.EnableIngress = src.Spec.EnableIngress

	if dst.Annotations == nil {
		dst.Annotations = make(map[string]string)
	}

	dst.ObjectMeta = src.ObjectMeta
	dst.Annotations["apps.v2.buildcrd.k8s.example.io/enable_pod"] = strconv.FormatBool(src.Spec.EnablePod)

	return nil
}

func (dst *App) ConvertFrom(srcRaw conversion.Hub) error {
	//fmt.Println("ConvertFrom >>>> v1 -> v2")
	src := srcRaw.(*appV1.App)
	dst.Spec.Replicas = src.Spec.Replicas
	dst.Spec.Image = src.Spec.Image
	dst.Spec.EnableService = src.Spec.EnableService
	dst.Spec.EnableIngress = src.Spec.EnableIngress
	enablePod, found := src.Annotations["apps.v2.buildcrd.k8s.example.io/enable_pod"]
	if found {
		if enablePod == "true" {
			dst.Spec.EnablePod = true
		} else {
			dst.Spec.EnablePod = false
		}
	} else {
		dst.Spec.EnablePod = false
	}

	dst.ObjectMeta = src.ObjectMeta

	return nil
}
