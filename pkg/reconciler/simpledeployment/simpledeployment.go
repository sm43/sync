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
	"fmt"
	"time"

	mf "github.com/manifestival/manifestival"
	"github.com/sm43/sync/pkg/apis/samples/v1alpha1"
	simpledeploymentreconciler "github.com/sm43/sync/pkg/client/injection/reconciler/samples/v1alpha1/simpledeployment"
	"github.com/sm43/sync/pkg/sync"
	"github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
)

var reconcileDuration = 10 * time.Second

type Reconciler struct {
	kubeclient   kubernetes.Interface
	manifest     *mf.Manifest
	tektonClient versioned.Interface
	enqueueAfter func(obj interface{}, after time.Duration)
	syncer       *sync.Manager
}

// Check that our Reconciler implements Interface
var _ simpledeploymentreconciler.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, d *v1alpha1.SimpleDeployment) reconciler.Event {
	logger := logging.FromContext(ctx)

	logger.Info("Reconciling for ", d.Name)

	acquire, _, msg, err := r.syncer.TryAcquire(d)
	if err != nil {
		fmt.Println("Error at start := err ", err)
		return nil
	}

	if !acquire {
		d.Status.Phase = v1alpha1.PHASEPENDING
		fmt.Println("dude no lock available, come back later !", msg)
		r.enqueueAfter(d, reconcileDuration)
		return nil
	}

	fmt.Println("acquired lock - ", d.Name)

	d.Status.Phase = v1alpha1.PHASERUNNING

	// check if pipelineRun exist with name of simple deployment
	pipelineRun, err := r.tektonClient.TektonV1beta1().PipelineRuns(d.Namespace).Get(ctx, d.Name, v1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		logger.Error("failed to get pipelineRun ", err)
		return err
	}

	// not found err
	if err != nil {
		//create pipelineRun
		trns, err := r.transformer(d)
		if err != nil {
			logger.Error("failed to transformed manifest ", err)
		}

		logger.Info("creating pipelineRun for sd: ", d.Name)
		if err := trns.Apply(); err != nil {
			logger.Error("failed to apply ", err)
			return err
		}
		logger.Info("created pipelineRun for sd: ", d.Name)

		return nil
	}

	logger.Info("Checking status of pipelineRun for SD ", d.Name)

	// found then check status
	succeeded := pipelineRun.Status.GetCondition(apis.ConditionSucceeded)
	if succeeded == nil {
		logger.Info("waiting for pipelineRun to complete")
		r.enqueueAfter(d, reconcileDuration)
		return nil
	}

	if succeeded.Status == corev1.ConditionTrue {
		d.Status.MarkTRReady()
	} else {
		logger.Info("waiting for pipelineRun to complete")
		r.enqueueAfter(d, reconcileDuration)
		return nil
	}

	d.Status.MarkReady()

	// release the lock for next one
	r.syncer.Release(d)

	d.Status.Phase = v1alpha1.PHASECOMPLETED

	logger.Info("Reconciling completed for ", d.Name)

	return nil
}

func (r *Reconciler) transformer(d *v1alpha1.SimpleDeployment) (trns mf.Manifest, err error) {

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
