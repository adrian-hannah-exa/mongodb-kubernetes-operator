package deployment

import (
        appsv1 "k8s.io/api/apps/v1"
        corev1 "k8s.io/api/core/v1"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type builder struct {
        name                  string
        namespace             string
        labels                map[string]string
        ownerReferences       []metav1.OwnerReference
        selector              map[string]string
        annotations           map[string]string
        replicas              int
        podTemplateSpec       corev1.PodTemplateSpec
        deploymentStrategy    appsv1.DeploymentStrategy
}

func (b *builder) SetLabels(labels map[string]string) *builder {
        b.labels = labels
        return b
}

func (b *builder) SetAnnotations(annotations map[string]string) *builder {
        b.annotations = annotations
        return b
}

func (b *builder) SetSelector(selector map[string]string) *builder {
        b.selector = selector
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

func (b *builder) SetReplicas(replicas int) *builder {
        b.replicas = replicas
        return b
}

func (b *builder) SetPodTemplate(podTemplate corev1.PodTemplate) *builder {
        b.podTemplate = podTemplate
        return b
}

func (b *builder) SetDeploymentStrategy(deploymentStrategy appsv1.DeploymentStrategy) *builder {
        b.deploymentStrategy = deploymentStrategy
        return b
}

func (b *builder) SetPodAnnotations(annotations map[string]string) *builder {
    b.podTemplate.Annotations = annotations
    return b
}

func (b *builder) SetPodLabel(labels map[string]string) *builder {}
    b.podTemplate.Labels = labels
    return b
}

func (b *builder) SetServiceAccount() *builder {}
func (b *builder) SetContainerMongoURISecret(mongoURI string) *builder {

}
func (b *builder) SetContainerEnvs() *builder {}
func (b *builder) SetContainerImage() *builder {}
func (b *builder) SetContainerImagePullPolicy() *builder {}
func (b *builder) SetContainerListenerPort() *builder {}
func (b *builder) SetContainerExtraArgs() *builder {}
func (b *builder) SetContainerLivenessProbe() *builder {}
func (b *builder) SetContainerReadinessProbe() *builder {}
func (b *builder) SetContainerResources() *builder {}
func (b *builder) SetContainerImagePullSecrets() *builder {}
func (b *builder) SetContainerNodeSelector() *builder {}
func (b *builder) SetContainerTolerations() *builder {}

func (b *builder) Build() appsv1.Deployment {
    return appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:            b.name,
            Namespace:       b.namespace,
            Labels:          b.labels,
            OwnerReferences: b.ownerReferences,
            Annotations:     b.annotations,
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: b.replicas,
            Selector: b.selector,
            Template: b.podTemplate,
            Strategy: b.deploymentStrategy,
        },
    }
}

func Builder() *builder {
    return &builder{
        labels:             map[string]string{},
        ownerReferences:    []metav1.OwnerReference{},
        selector:           map[string]string{},
        annotations:        map[string]string{},
        replicas:           1,
        podTemplate:        corev1.PodTemplate{},
        deploymentStrategy: appsv1.DeploymentStrategy{},
    }
}
