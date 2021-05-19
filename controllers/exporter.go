package controllers

import (
	"context"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	mdbv1 "github.com/mongodb/mongodb-kubernetes-operator/api/v1"
	"github.com/mongodb/mongodb-kubernetes-operator/pkg/kube/service"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	metricsUsername     = "metrics"
	prometheusNamespace = "monitoring"
	prometheusName      = "app-prometheus"
)

// buildExporterService creates a Service that will be used for the prometheus exporter
func buildExporterService(mdb mdbv1.MongoDBCommunity) corev1.Service {
	label := make(map[string]string)
	label["app"] = mdb.ServiceName()
	return service.Builder().
		SetName(mdb.Name + "-exporter-svc").
		SetNamespace(mdb.Namespace).
		SetLabels(label).
		SetSelector(label).
		SetServiceType(corev1.ServiceTypeClusterIP).
		SetClusterIP("None").
		SetPort(9216).
		SetPublishNotReadyAddresses(true).
		Build()
}

func (r *ReplicaSetReconciler) ensureExporterService(mdb mdbv1.MongoDBCommunity) error {
	svc := buildExporterService(mdb)
	err := r.client.Create(context.TODO(), &svc)
	if err != nil && apiErrors.IsAlreadyExists(err) {
		r.log.Infof("The exporter service already exists... moving forward: %s", err)
		return nil
	}
	return err
}

func (r *ReplicaSetReconciler) ensureExporterServiceMonitor(mdb mdbv1.MongoDBCommunity) error {
	svcMon := &monitoringv1.ServiceMonitor{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "monitoring.coreos.com/v1",
			Kind:       "ServiceMonitor",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mdb.Name + "-exporter",
			Namespace: prometheusNamespace,
			Labels: map[string]string{
				"prometheus": prometheusName,
			},
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: []monitoringv1.Endpoint{
				monitoringv1.Endpoint{
					Interval:      "30s",
					Path:          "/metrics",
					Scheme:        "http",
					ScrapeTimeout: "30s",
				},
			},
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{mdb.Namespace},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": mdb.ServiceName(),
				},
			},
		},
	}
	err := r.client.Create(context.TODO(), svcMon)
	if err != nil && apiErrors.IsAlreadyExists(err) {
		r.log.Infof("The service monitor already exists... moving forward: %s", err)
		return nil
	}
	return err
}

func insertMetricsUser(mdb *mdbv1.MongoDBCommunity) mdbv1.MongoDBUser {
	metricsUser := mdbv1.MongoDBUser{
		Name:              metricsUsername,
		DB:                "admin",
		PasswordSecretRef: mdbv1.SecretKeyReference{Name: mdb.Name + "-metrics-user"},
		Roles: []mdbv1.Role{
			mdbv1.Role{
				Name: "clusterMonitor",
				DB:   "admin",
			},
			mdbv1.Role{
				Name: "read",
				DB:   "local",
			},
			mdbv1.Role{
				Name: "find",
				DB:   "admin",
			},
		},
		ScramCredentialsSecretName: mdb.Name + "-metrics-user",
	}

	contains := false
	for i, value := range mdb.Spec.Users {
		if value.Name == metricsUsername {
			contains = true
			mdb.Spec.Users[i] = metricsUser
			break
		}
	}
	if contains == false {
		mdb.Spec.Users = append(mdb.Spec.Users, metricsUser)
	}

	return metricsUser

}
