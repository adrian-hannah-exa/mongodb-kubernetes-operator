package deployment

import (
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Getter interface {
	GetDeployment(objectKey client.ObjectKey) (appsv1.Deployment, error)
}

type Updater interface {
	UpdateDeployment(secret appsv1.Deployment) error
}

type Creator interface {
	CreateDeployment(secret appsv1.Deployment) error
}

type GetUpdater interface {
	Getter
	Updater
}

type GetUpdateCreator interface {
	Getter
	Updater
	Creator
}

// Merge merges `source` into `dest`. Both arguments will remain unchanged
// a new deployment will be created and returned.
// The "merging" process is arbitrary and it only handle specific attributes
func Merge(dest appsv1.Deployment, source appsv1.Deployment) appsv1.Deployment {
	for k, v := range source.ObjectMeta.Annotations {
		dest.ObjectMeta.Annotations[k] = v
	}

	for k, v := range source.ObjectMeta.Labels {
		dest.ObjectMeta.Labels[k] = v
	}

	dest.Spec.Type = source.Spec.Type
	return dest
}
