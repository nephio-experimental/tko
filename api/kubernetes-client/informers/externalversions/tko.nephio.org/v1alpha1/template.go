// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	tkonephioorgv1alpha1 "github.com/nephio-experimental/tko/api/krm/tko.nephio.org/v1alpha1"
	versioned "github.com/nephio-experimental/tko/api/kubernetes-client/clientset/versioned"
	internalinterfaces "github.com/nephio-experimental/tko/api/kubernetes-client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/nephio-experimental/tko/api/kubernetes-client/listers/tko.nephio.org/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// TemplateInformer provides access to a shared informer and lister for
// Templates.
type TemplateInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.TemplateLister
}

type templateInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewTemplateInformer constructs a new informer for Template type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewTemplateInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredTemplateInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredTemplateInformer constructs a new informer for Template type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredTemplateInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.TkoV1alpha1().Templates().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.TkoV1alpha1().Templates().Watch(context.TODO(), options)
			},
		},
		&tkonephioorgv1alpha1.Template{},
		resyncPeriod,
		indexers,
	)
}

func (f *templateInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredTemplateInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *templateInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&tkonephioorgv1alpha1.Template{}, f.defaultInformer)
}

func (f *templateInformer) Lister() v1alpha1.TemplateLister {
	return v1alpha1.NewTemplateLister(f.Informer().GetIndexer())
}
