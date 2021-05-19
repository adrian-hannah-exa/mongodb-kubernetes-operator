package cronjob

import (
	"fmt"
	"hash/fnv"

	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type builder struct {
	name      string
	namespace string
	schedule  string
	labels    map[string]string

	concurrencyPolicy          batchv1beta1.ConcurrencyPolicy
	ownerReferences            []metav1.OwnerReference
	podTemplateSpec            corev1.PodTemplateSpec
	readinessProbePerContainer map[string]*corev1.Probe
	volumeClaimsTemplates      []corev1.PersistentVolumeClaim
	volumeMountsPerContainer   map[string][]corev1.VolumeMount
}

func (b *builder) SetLabels(labels map[string]string) *builder {
	b.labels = labels
	return b
}

func (b *builder) SetConcurrencyPolicy(policy batchv1beta1.ConcurrencyPolicy) *builder {
	b.concurrencyPolicy = policy
	return b
}

func (b *builder) SetName(name string) *builder {
	b.name = name
	return b
}

func (b *builder) SetNamespace(namespace string) *builder {
	b.namespace = namespace
	return b
}

func (b *builder) SetOwnerReferences(ownerReferences []metav1.OwnerReference) *builder {
	b.ownerReferences = ownerReferences
	return b
}

func (b *builder) SetPodTemplateSpec(podTemplateSpec corev1.PodTemplateSpec) *builder {
	b.podTemplateSpec = podTemplateSpec
	return b
}

func (b *builder) GenerateSchedule() *builder {
	hashKey := fnv.New64()
	hashKey.Write([]byte(b.name))
	hour := hashKey.Sum64() % 12
	return b.SetSchedule(fmt.Sprintf("0 %d,%d * * *", hour, hour+12))
}

func (b *builder) SetSchedule(schedule string) *builder {
	b.schedule = schedule
	return b
}

func (b builder) Build() batchv1beta1.CronJob {
	return batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:            b.name,
			Namespace:       b.namespace,
			OwnerReferences: b.ownerReferences,
			Labels:          b.labels,
		},
		Spec: batchv1beta1.CronJobSpec{
			ConcurrencyPolicy: b.concurrencyPolicy,
			Suspend:           &[]bool{true}[0], // TODO Temp until works
			Schedule:          b.schedule,
			JobTemplate: batchv1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: b.podTemplateSpec,
				},
			},
		},
	}
}

func Builder() *builder {
	return &builder{
		concurrencyPolicy:          batchv1beta1.ForbidConcurrent,
		ownerReferences:            []metav1.OwnerReference{},
		labels:                     map[string]string{},
		podTemplateSpec:            corev1.PodTemplateSpec{},
		readinessProbePerContainer: map[string]*corev1.Probe{},
		volumeClaimsTemplates:      []corev1.PersistentVolumeClaim{},
		volumeMountsPerContainer:   map[string][]corev1.VolumeMount{},
	}
}
