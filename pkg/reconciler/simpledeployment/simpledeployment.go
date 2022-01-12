/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package simpledeployment

import (
	"context"
	"time"

	mf "github.com/manifestival/manifestival"
	"github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	v1beta1Taskrun "github.com/tektoncd/pipeline/pkg/client/listers/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"

	samplesv1alpha1 "knative.dev/sample-controller/pkg/apis/samples/v1alpha1"
	simpledeploymentreconciler "knative.dev/sample-controller/pkg/client/injection/reconciler/samples/v1alpha1/simpledeployment"
)

var reconcileDuration = 10 * time.Second

//var freq = map[string]int64{
//	"repo-abc": 3,
//	"repo-xyz": 2,
//}

type Reconciler struct {
	kubeclient    kubernetes.Interface
	taskRunLister v1beta1Taskrun.TaskRunLister
	manifest      *mf.Manifest
	tektonClient  versioned.Interface
	enqueueAfter  func(obj interface{}, after time.Duration)
}

// Check that our Reconciler implements Interface
var _ simpledeploymentreconciler.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, d *samplesv1alpha1.SimpleDeployment) reconciler.Event {
	logger := logging.FromContext(ctx)

	logger.Info("Reconciling for ", d.Name)

	// check if taskRun exist with name of simple deployment

	taskrun, err := r.tektonClient.TektonV1beta1().TaskRuns(d.Namespace).Get(ctx, d.Name, v1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		logger.Error("failed to get taskRun ", err)
	}

	// not found err
	if err != nil {
		//create taskRun
		trns, err := r.transformer(d)
		if err != nil {
			logger.Error("failed to transformed manifest ", err)
		}

		logger.Info("creating taskRun for sd: ", d.Name)
		if err := trns.Apply(); err != nil {
			logger.Error("failed to apply ", err)
			return err
		}
		logger.Info("created taskRun for sd: ", d.Name)

		return nil
	}

	logger.Info("Checking status of TaskRun for SD ", d.Name)

	// found then check status
	succeeded := taskrun.Status.GetCondition(apis.ConditionSucceeded)
	if succeeded == nil {
		logger.Info("waiting for taskRun to complete")
		r.enqueueAfter(d, reconcileDuration)
		return nil
	}

	if succeeded.Status == corev1.ConditionTrue {
		d.Status.MarkTRReady()
	} else {
		logger.Info("waiting for taskRun to complete")
		r.enqueueAfter(d, reconcileDuration)
		return nil
	}

	d.Status.MarkReady()

	logger.Info("Reconciling completed for ", d.Name)
	return nil
}

func (r *Reconciler) transformer(d *samplesv1alpha1.SimpleDeployment) (trns mf.Manifest, err error) {

	trns, err = r.manifest.Transform(mf.InjectNamespace(d.Namespace), mf.InjectOwner(d), injectName(d.GetName()))
	if err != nil {
		return mf.Manifest{}, err
	}
	return trns, nil
}

func injectName(name string) mf.Transformer {
	return func(u *unstructured.Unstructured) error {
		u.SetName(name)
		return nil
	}
}
