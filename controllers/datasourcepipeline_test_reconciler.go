package controllers

import (
	"context"
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/gomega"
	"github.com/redhatinsights/xjoin-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

type DatasourcePipelineTestReconciler struct {
	Namespace         string
	Name              string
	K8sClient         client.Client
	createdDatasource v1alpha1.XJoinDataSource
}

func (d *DatasourcePipelineTestReconciler) newXJoinDataSourcePipelineReconciler() *XJoinDataSourcePipelineReconciler {
	return NewXJoinDataSourcePipelineReconciler(
		d.K8sClient,
		scheme.Scheme,
		testLogger,
		record.NewFakeRecorder(10),
		d.Namespace,
		true)
}

func (d *DatasourcePipelineTestReconciler) createValidDataSourcePipeline() {
	ctx := context.Background()

	datasourceSpec := v1alpha1.XJoinDataSourcePipelineSpec{
		Name:             d.Name,
		Version:          "1234",
		AvroSchema:       "{}",
		DatabaseHostname: &v1alpha1.StringOrSecretParameter{Value: "dbHost"},
		DatabasePort:     &v1alpha1.StringOrSecretParameter{Value: "8080"},
		DatabaseUsername: &v1alpha1.StringOrSecretParameter{Value: "dbUsername"},
		DatabasePassword: &v1alpha1.StringOrSecretParameter{Value: "dbPassword"},
		DatabaseName:     &v1alpha1.StringOrSecretParameter{Value: "dbName"},
		DatabaseTable:    &v1alpha1.StringOrSecretParameter{Value: "dbTable"},
		Pause:            false,
	}

	datasource := &v1alpha1.XJoinDataSourcePipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.Name,
			Namespace: d.Namespace,
		},
		Spec: datasourceSpec,
		TypeMeta: metav1.TypeMeta{
			APIVersion: "xjoin.cloud.redhat.com/v1alpha1",
			Kind:       "XJoinDataSourcePipeline",
		},
	}

	Expect(d.K8sClient.Create(ctx, datasource)).Should(Succeed())

	//validate datasource spec is created correctly
	datasourceLookupKey := types.NamespacedName{Name: d.Name, Namespace: d.Namespace}
	createdDataSourcePipeline := &v1alpha1.XJoinDataSourcePipeline{}

	Eventually(func() bool {
		err := d.K8sClient.Get(ctx, datasourceLookupKey, createdDataSourcePipeline)
		if err != nil {
			return false
		}
		return true
	}, 10*time.Second, 100*time.Millisecond).Should(BeTrue())

	Expect(createdDataSourcePipeline.Spec.Name).Should(Equal(d.Name))
	Expect(createdDataSourcePipeline.Spec.Version).Should(Equal("1234"))
	Expect(createdDataSourcePipeline.Spec.Pause).Should(Equal(false))
	Expect(createdDataSourcePipeline.Spec.AvroSchema).Should(Equal("{}"))
	Expect(createdDataSourcePipeline.Spec.DatabaseHostname).Should(Equal(&v1alpha1.StringOrSecretParameter{Value: "dbHost"}))
	Expect(createdDataSourcePipeline.Spec.DatabasePort).Should(Equal(&v1alpha1.StringOrSecretParameter{Value: "8080"}))
	Expect(createdDataSourcePipeline.Spec.DatabaseUsername).Should(Equal(&v1alpha1.StringOrSecretParameter{Value: "dbUsername"}))
	Expect(createdDataSourcePipeline.Spec.DatabasePassword).Should(Equal(&v1alpha1.StringOrSecretParameter{Value: "dbPassword"}))
	Expect(createdDataSourcePipeline.Spec.DatabaseName).Should(Equal(&v1alpha1.StringOrSecretParameter{Value: "dbName"}))
	Expect(createdDataSourcePipeline.Spec.DatabaseTable).Should(Equal(&v1alpha1.StringOrSecretParameter{Value: "dbTable"}))
}

func (d *DatasourcePipelineTestReconciler) ReconcileNew() v1alpha1.XJoinDataSourcePipeline {
	d.registerNewMocks()
	d.createValidDataSourcePipeline()
	createdDataSourcePipeline := &v1alpha1.XJoinDataSourcePipeline{}
	result := d.reconcile()
	Expect(result).To(Equal(reconcile.Result{Requeue: false, RequeueAfter: 30000000000}))

	datasourceLookupKey := types.NamespacedName{Name: d.Name, Namespace: d.Namespace}
	Eventually(func() bool {
		err := d.K8sClient.Get(context.Background(), datasourceLookupKey, createdDataSourcePipeline)
		if err != nil {
			return false
		}
		return true
	}, K8sGetTimeout, K8sGetInterval).Should(BeTrue())

	return *createdDataSourcePipeline
}

func (d *DatasourcePipelineTestReconciler) reconcile() reconcile.Result {
	xjoinDataSourcePipelineReconciler := d.newXJoinDataSourcePipelineReconciler()
	datasourceLookupKey := types.NamespacedName{Name: d.Name, Namespace: d.Namespace}
	result, err := xjoinDataSourcePipelineReconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: datasourceLookupKey})
	checkError(err)
	return result
}

func (d *DatasourcePipelineTestReconciler) registerNewMocks() {
	httpmock.Reset()
	httpmock.RegisterNoResponder(httpmock.InitialTransport.RoundTrip) //disable mocks for unregistered http requests

	httpmock.RegisterResponder(
		"GET",
		"http://apicurio:1080/apis/ccompat/v6/subjects/xjoindatasourcepipeline.test-data-source-pipeline.1234-value/versions/1",
		httpmock.NewStringResponder(404, `{"message":"No version '1' found for artifact with ID 'xjoindatasourcepipeline.test-data-source-pipeline.1234-value' in group 'null'.","error_code":40402}`).Times(1))

	httpmock.RegisterResponder(
		"POST",
		"http://apicurio:1080/apis/ccompat/v6/subjects/xjoindatasourcepipeline.test-data-source-pipeline.1234-value/versions",
		httpmock.NewStringResponder(200, `{"createdBy":"","createdOn":"2022-07-27T17:28:11+0000","modifiedBy":"","modifiedOn":"2022-07-27T17:28:11+0000","id":1,"version":1,"type":"AVRO","globalId":1,"state":"ENABLED","groupId":"null","contentId":1,"references":[]}`))

	httpmock.RegisterResponder(
		"GET",
		"http://apicurio:1080/apis/ccompat/v6/subjects/xjoindatasourcepipeline.test-data-source-pipeline.1234-value/versions/latest",
		httpmock.NewStringResponder(200, `{"schema":"{\"name\":\"Value\",\"namespace\":\"xjoindatasourcepipeline.test-data-source-pipeline\"}","schemaType":"AVRO","references":[]}`))
}