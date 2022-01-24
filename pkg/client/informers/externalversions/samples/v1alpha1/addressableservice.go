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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	samplesv1alpha1 "github.com/sm43/sync/pkg/apis/samples/v1alpha1"
	versioned "github.com/sm43/sync/pkg/client/clientset/versioned"
	internalinterfaces "github.com/sm43/sync/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/sm43/sync/pkg/client/listers/samples/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// AddressableServiceInformer provides access to a shared informer and lister for
// AddressableServices.
type AddressableServiceInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.AddressableServiceLister
}

type addressableServiceInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewAddressableServiceInformer constructs a new informer for AddressableService type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewAddressableServiceInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredAddressableServiceInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredAddressableServiceInformer constructs a new informer for AddressableService type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredAddressableServiceInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SamplesV1alpha1().AddressableServices(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SamplesV1alpha1().AddressableServices(namespace).Watch(context.TODO(), options)
			},
		},
		&samplesv1alpha1.AddressableService{},
		resyncPeriod,
		indexers,
	)
}

func (f *addressableServiceInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredAddressableServiceInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *addressableServiceInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&samplesv1alpha1.AddressableService{}, f.defaultInformer)
}

func (f *addressableServiceInformer) Lister() v1alpha1.AddressableServiceLister {
	return v1alpha1.NewAddressableServiceLister(f.Informer().GetIndexer())
}
