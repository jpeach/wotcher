package k

import (
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	DefaultResyncPeriod = 10 * time.Minute
)

// Reader provides read-only access to Kubernetes resources.
type Reader struct {
	Conf            *rest.Config
	Discovery       *discovery.DiscoveryClient
	InformerFactory dynamicinformer.DynamicSharedInformerFactory
}

func (r *Reader) InformOnResource(gvr schema.GroupVersionResource, handler cache.ResourceEventHandler) {
	i := r.InformerFactory.ForResource(gvr)
	i.Informer().AddEventHandler(handler)
}

func (r *Reader) Run(stopChan <-chan struct{}) {
	r.InformerFactory.Start(stopChan)

	// Block until the channel closes and all the informers are done.
	<-stopChan
}

func IsSubResource(resource string) bool {
	return strings.Contains(resource, "/")
}

func NamespacedNameOf(u *unstructured.Unstructured) types.NamespacedName {
	return types.NamespacedName{
		Namespace: u.GetNamespace(),
		Name:      u.GetName(),
	}
}

// NewScheme ...
func NewScheme(addTo ...func(*runtime.Scheme) error) *runtime.Scheme {
	s := runtime.NewScheme()

	// Add default types to the scheme.
	scheme.AddToScheme(s)

	for _, a := range addTo {
		a(s)
	}

	return s
}

// NewReaderForScheme returns a new informer cache that uses the given scheme.
func NewReaderForScheme(s *runtime.Scheme) (*Reader, error) {
	var err error
	var reader Reader

	reader.Conf, err = ctrl.GetConfig()
	if err != nil {
		return nil, err
	}

	reader.Discovery, err = discovery.NewDiscoveryClientForConfig(reader.Conf)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(reader.Conf)
	if err != nil {
		return nil, err
	}

	reader.InformerFactory = dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		dynamicClient,
		DefaultResyncPeriod,
		corev1.NamespaceAll,
		func(options *metav1.ListOptions) {
		},
	)

	return &reader, nil
}
