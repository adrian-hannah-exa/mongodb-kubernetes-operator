package cronjob

import (
	batchv1 "k8s.io/api/batch/v1beta1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Getter interface {
	GetCronJob(objectKey client.ObjectKey) (batchv1.CronJob, error)
}

type Updater interface {
	UpdateCronJob(cm batchv1.CronJob) error
}

type Creator interface {
	CreateCronJob(cm batchv1.CronJob) error
}

type Deleter interface {
	DeleteCronJob(key client.ObjectKey) error
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

type GetUpdateCreateDeleter interface {
	Getter
	Updater
	Creator
	Deleter
}

const (
	lineSeparator     = "\n"
	keyValueSeparator = "="
)

// CreateOrUpdate creates the given CronJob if it doesn't exist,
// or updates it if it does.
func CreateOrUpdate(getUpdateCreator GetUpdateCreator, cm batchv1.CronJob) error {
	_, err := getUpdateCreator.GetCronJob(types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return getUpdateCreator.CreateCronJob(cm)
		}
		return err
	}
	return getUpdateCreator.UpdateCronJob(cm)
}
