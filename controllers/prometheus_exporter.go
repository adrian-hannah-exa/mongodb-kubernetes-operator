package controllers

import (
    "crypto/sha256"
    "fmt"
    "strings"

    "github.com/pkg/errors"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

    "github.com/mongodb/mongodb-kubernetes-operator/controllers/construct"
    "github.com/mongodb/mongodb-kubernetes-operator/pkg/automationconfig"

    "github.com/mongodb/mongodb-kubernetes-operator/pkg/kube/secret"

    "github.com/mongodb/mongodb-kubernetes-operator/pkg/kube/configmap"

    "github.com/mongodb/mongodb-kubernetes-operator/pkg/kube/podtemplatespec"
    "github.com/mongodb/mongodb-kubernetes-operator/pkg/kube/statefulset"

    apiErrors "k8s.io/apimachinery/pkg/api/errors"

    "github.com/mongodb/mongodb-kubernetes-operator/pkg/kube/deployment"
    "github.com/mongodb/mongodb-kubernetes-operator/pkg/kube/service"
    mdbv1 "github.com/mongodb/mongodb-kubernetes-operator/api/v1"
    corev1 "k8s.io/api/core/v1"
)

const (
    exporterUserName = "metrics"
    clusterDomain = "cluster.local"
    prometheusNamespace = "monitoring"
)

// buildExporterServiceAccount creates a ServiceAccount that will be used for the Prometheus Exporter
func buildExporterServiceAccount(mdb mdbv1.MongoDBCommunity) corev1.ServiceAccount {
    return corev1.ServiceAccount {
        ObjectMeta: metav1.ObjectMeta {
            Name: mdb.Name() + "-exporter"
            Namespace: mdb.Namespace()
            Labels: struct {
                "app": mdb.Name() + "-exporter"
            }
        }
    }
}

// buildMongoDbUriSecret creates a Secret with the URI of the MongoDB instance to monitor
func buildMongoDbUriSecret(mdb mdbv1.MongoDBCommunity) corev1.Secret {
    label := make(map[string]string)
    label["app"] = mdb.Name() + '-exporter'

    password := ''
    for _, user := range mdb.Users {
        if user.Name() == exporterUserName {
            secretsClient := r.client.CoreV1().Secrets(mdb.Namespace)
            passwordRef := corev1.Secret{}
            err := secretsClient.Get(context.TODO(), user.PasswordSecretRef.Name(), %passwordRef)
            password = passwordRef.Data[user.PasswordSecretRef.Key]
            break
        }
    }

    full_uri = fmt.Sprintf(
        "mongodb://%s:%s@%s.%s.%s.%s:%d",
        exporterUserName,
        password,
        m.ServiceName(),
        m.Namespace,
        clusterDomain,
        27017)

    return secret.Builder().
        SetName(mdb.Name() + '-exporter-uri').
        SetNamespace(mdb.Namespace).
        SetLabels(label).
        SetField('mongodb-uri', full_uri).
        Build()
}

func (r *ReplicaSetReconciler) ensureExporterDeployment(mdb mdbv1.MongoDBCommunity) error {
    dpl := buildExporterDeployment(mdb)
    err := r.client.Create(context.TODO(), &dpl)
    if err != nil && apiErrors.IsAlreadyExists(err) {
        r.log.Infof("The deployment already exists... moving forward: %s", err)
        return nil
    }
    return err
}

// buildExporterDeployment creates a Deployment for the Prometheus Exporter
func buildExporterDeployment(mdb mdbv1.MongoDBCommunity) appsv1.Deployment {
    label := make(map[string]string)
    label["app"] = mdb.Name() + '-exporter'

    ExporterServiceAccount := buildExporterServiceAccount(mdb)
    err := r.client.Create(context.TODO(), &ExporterServiceAccount)
    if err != nil && !apiErrors.IsAlreadyExists(err) {
        r.log.Infof("Error when building exporter service account: %s", err)
        return err
    }

    MongoDbUriSecret := buildMongoDbUriSecret(mdb)
    err := r.client.Create(context.TODO(), &MongoDbUriSecret)
    if err != nil && !apiErrors.IsAlreadyExists(err) {
        r.log.Infof("Error when building Mongo DB URI Secret: %s", err)
        return err
    }

    return deployment.Builder().
        SetName(mdb.Name() + '-exporter').
        SetNamespace(mdb.Namespace).
        SetLabel(label).
        SetAnnotations().
        SetReplicas(1).
        SetSelector(label).
        SetPodAnnotations().
        SetPodLabel(label).
        SetServiceAccount(ExporterServiceAccount.Name()).
        SetContainerMongoDBURI(MongoDbUriSecret.Name()).
        SetContainerEnvs().
        SetContainerImage().
        SetContainerImagePullPolicy().
        SetContainerListenerPort().
        SetContainerExtraArgs().
        SetContainerLivenessProbe().
        SetContainerReadinessProbe().
        SetContainerResources().
        SetContainerImagePullSecrets().
        SetContainerNodeSelector().
        SetContainerTolerations().
        Build()

func (r *ReplicaSetReconciler) ensureExporterService(mdb mdbv1.MongoDBCommunity) error {
    svc := buildExporterService(mdb)
    err := r.client.Create(context.TODO(), &svc)
    if err != nil && apiErrors.IsAlreadyExists(err) {
        r.log.Infof("The service already exists... moving forward: %s", err)
        return nil
    }
    return err
}

// buildExporterService creates a Service that will be used for the Prometheus Exporter
func buildExporterService(mdb mdbv1.MongoDBCommunity) corev1.Service {
    label := make(map[string]string)
    label["app"] = mdb.Name() + '-exporter'
    return service.Builder().
        SetName(mdb.Name() + '-exporter-svc').
        SetNamespace(mdb.Namespace).
        SetLabel(label).
        SetSelector(label).
        SetServiceType(corev1.ServiceTypeClusterIP).
        SetClusterIP("None").
        SetPort(9216).
        SetPublishNotReadyAddresses(true).
        Build()
}

func (r *ReplicaSetReconciler) ensureExporterServiceMonitor(mdb mdbv1.MongoDBCommunity) error {
    svcMon := buildExporterServiceMonitor(mdb)
    data, err := r.client.RESTClient().
        Get().
        AbsPath("/apis/monitoring.coreos.com/v1").
        Namespace(prometheusNamespace).
        Resource("ServiceMonitor").
        Body(svcMon).
        DoRaw(context.TODO())

// buildExporterServiceMonitor creates a ServiceMonitor that will be used for the Prometheus Exporter
func buildExporterServiceMonitor(mdb mdbv1.MongoDBCommunity) &ServiceMonitor {
    return &ServiceMonitor {
        TypeMeta: metav1.TypeMeta {
            APIVersion: "monitoring.coreos.com/v1"
            Kind: "ServiceMonitor"
        }
        ObjectMeta: metav1.ObjectMeta {
            Name: mdb.Name() + "-exporter"
            Namespace: prometheusNamespace
            Labels: struct {
                "prometheus": "prometheus"
            }
        }
        Spec: struct {
            Endpoints: struct {
                Interval: "30s"
                Path: "/metrics"
                Scheme: "http"
                ScrapeTimeout: "30s"
            }
            NamespaceSelector: struct {
                MatchNames: [m.Namespace()]
            }
            Selector: metav1.LabelSelector {
                MatchLabels: struct {
                    "app": mdb.Name() + '-exporter'
                }
            }
        }
    }
}
