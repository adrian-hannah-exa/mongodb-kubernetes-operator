package controllers

import (
	"context"
	"os"
	"strings"

	"github.com/mongodb/mongodb-kubernetes-operator/pkg/kube/container"

	"github.com/mongodb/mongodb-kubernetes-operator/pkg/kube/podtemplatespec"

	mdbv1 "github.com/mongodb/mongodb-kubernetes-operator/api/v1"
	"github.com/mongodb/mongodb-kubernetes-operator/pkg/kube/cronjob"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	backupUsername              = "backup"
	MongodbBackupRootEnv        = "MONGODB_BACKUP_ROOT"
	MongodbBackupImageEnv       = "MONGODB_BACKUP_IMAGE"
	MongodbBackupPullSecretsEnv = "MONGODB_BACKUP_IMAGE_PULL_SECRETS"

	gcpCredsSecretName = "data-backuper-credentials"
	gcpCredsSecretKey  = "GOOGLE_SERVICE_ACCOUNT_JSON_KEY"
)

func (r *ReplicaSetReconciler) ensureBackupCronJob(mdb mdbv1.MongoDBCommunity) error {
	cj := buildBackupCronJob(mdb)
	err := r.client.Create(context.TODO(), &cj)
	if err != nil && apiErrors.IsAlreadyExists(err) {
		r.log.Infof("The backup cronjob already exists... moving forward: %s", err)
		return nil
	}
	return err
}

// buildBackupCronJob creates a CronJob that will create a backup of the mongo database
func buildBackupCronJob(mdb mdbv1.MongoDBCommunity) batchv1beta1.CronJob {
	backupImg := os.Getenv(MongodbBackupImageEnv)
	backupPullSecrets := os.Getenv(MongodbBackupPullSecretsEnv)
	backupRoot := os.Getenv(MongodbBackupRootEnv)

	pullSecrets := podtemplatespec.NOOP()
	if backupPullSecrets != "" {
		for _, i := range strings.Split(backupPullSecrets, ",") {
			pullSecrets = podtemplatespec.Apply(pullSecrets, podtemplatespec.WithImagePullSecrets(i))
		}
	}

	var podSpec corev1.PodTemplateSpec
	mods := podtemplatespec.Apply(
		podtemplatespec.WithContainer(
			"mongodb-backup",
			container.Apply(
				container.WithImage(backupImg),
				container.WithArgs([]string{
					"/bin/sh",
					"-c",
					"mkdir bkps;/usr/bin/mongodump $MONGODB_URI -o bkps/; gcloud auth activate-service-account --key-file=$GOOGLE_APPLICATION_CREDENTIALS; gsutil -m cp -r bkps/* " + backupRoot + mdb.Namespace + "/" + mdb.Name + "/" + mdb.Name + "-$(date +%Y%m%d-%H%M%S)/",
				}),
				container.WithVolumeMounts([]corev1.VolumeMount{
					corev1.VolumeMount{
						Name:      "creds-json",
						ReadOnly:  true,
						MountPath: "/etc/gcp",
					},
				}),
				container.WithEnvs(
					corev1.EnvVar{
						Name: "MONGODB_URI",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: mdb.Name + "-backup-uri",
								},
								Key: "mongodb-uri",
							},
						},
					},
					corev1.EnvVar{
						Name:  "GOOGLE_APPLICATION_CREDENTIALS",
						Value: "/etc/gcp/creds.json",
					},
				),
			),
		),
		pullSecrets,
		podtemplatespec.WithVolume(corev1.Volume{
			Name: "creds-json",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: gcpCredsSecretName,
					Items: []corev1.KeyToPath{
						corev1.KeyToPath{
							Key:  gcpCredsSecretKey,
							Path: "creds.json",
						},
					},
				},
			},
		}),
		podtemplatespec.WithRestartPolicy(corev1.RestartPolicyOnFailure),
	)
	mods(&podSpec)

	label := make(map[string]string)
	label["app"] = mdb.ServiceName()
	return cronjob.Builder().
		SetName(mdb.Name + "-backup").
		SetNamespace(mdb.Namespace).
		SetLabels(label).
		GenerateSchedule().
		SetPodTemplateSpec(podSpec).
		Build()
}

func insertBackupUser(mdb *mdbv1.MongoDBCommunity) mdbv1.MongoDBUser {
	backupUser := mdbv1.MongoDBUser{
		Name:              backupUsername,
		DB:                "admin",
		PasswordSecretRef: mdbv1.SecretKeyReference{Name: mdb.Name + "-backup-user"},
		Roles: []mdbv1.Role{
			mdbv1.Role{
				Name: "backup",
				DB:   "admin",
			},
		},
		ScramCredentialsSecretName: mdb.Name + "-backup-user",
	}

	contains := false
	for i, value := range mdb.Spec.Users {
		if value.Name == backupUsername {
			contains = true
			mdb.Spec.Users[i] = backupUser
			break
		}
	}
	if contains == false {
		mdb.Spec.Users = append(mdb.Spec.Users, backupUser)
	}

	return backupUser

}
