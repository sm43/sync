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
	"os"
	"path/filepath"

	"github.com/go-logr/zapr"
	mfc "github.com/manifestival/client-go-client"
	mf "github.com/manifestival/manifestival"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/logging"

	pipClient "github.com/tektoncd/pipeline/pkg/client/injection/client"
	taskruninformer "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1beta1/taskrun"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	"knative.dev/sample-controller/pkg/apis/samples/v1alpha1"
	simpledeploymentinformer "knative.dev/sample-controller/pkg/client/injection/informers/samples/v1alpha1/simpledeployment"
	simpledeploymentreconciler "knative.dev/sample-controller/pkg/client/injection/reconciler/samples/v1alpha1/simpledeployment"
)

// NewController creates a Reconciler and returns the result of NewImpl.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {

	logger := logging.FromContext(ctx)

	simpledeploymentInformer := simpledeploymentinformer.Get(ctx)
	taskrunInformer := taskruninformer.Get(ctx)

	mfclient, err := mfc.NewClient(injection.GetConfig(ctx))
	if err != nil {
		logger.Fatalf("failed to init client")
	}
	mflogger := zapr.NewLogger(logger.Named("manifestival").Desugar())

	manifest, err := mf.ManifestFrom(mf.Slice{}, mf.UseClient(mfclient), mf.UseLogger(mflogger))
	if err != nil {
		logger.Fatalw("Error creating initial manifest", zap.Error(err))
	}

	koDataDir := os.Getenv("KO_DATA_PATH")
	if koDataDir == "" {
		logger.Fatalw("failed to get ko data")
	}
	resPath := filepath.Join(koDataDir, "imp")

	result, err := mf.NewManifest(resPath)
	if err != nil {
		logger.Fatalw("failed to read resource from path")
	}

	manifest = manifest.Append(result)

	if len(manifest.Resources()) == 0 {
		logger.Fatalw("no resources in manifest")
	}

	r := &Reconciler{
		kubeclient:    kubeclient.Get(ctx),
		taskRunLister: taskrunInformer.Lister(),
		manifest:      &manifest,
		tektonClient:  pipClient.Get(ctx),
	}

	impl := simpledeploymentreconciler.NewImpl(ctx, r)

	r.enqueueAfter = impl.EnqueueAfter

	simpledeploymentInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	taskrunInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterController(&v1alpha1.SimpleDeployment{}),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}
