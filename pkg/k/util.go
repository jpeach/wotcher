package k

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// Reader provides read-only access to Kubernetes resources.
type Reader struct {
	Conf      *rest.Config
	Cache     cache.Cache
	Mapper    meta.RESTMapper
	Discovery *discovery.DiscoveryClient
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

// NewClientForScheme returns a new Kubernetes client that uses the given scheme.
func NewClientForScheme(s *runtime.Scheme) (client.Client, error) {
	conf, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}

	mapper, err := apiutil.NewDiscoveryRESTMapper(conf)
	if err != nil {
		return nil, err
	}

	return client.New(conf,
		client.Options{
			Scheme: s,
			Mapper: mapper,
		},
	)
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

	gr, err := restmapper.GetAPIGroupResources(reader.Discovery)
	if err != nil {
		return nil, err
	}

	reader.Mapper = restmapper.NewDiscoveryRESTMapper(gr)

	reader.Cache, err = cache.New(reader.Conf,
		cache.Options{
			Scheme:    s,
			Mapper:    reader.Mapper,
			Namespace: metav1.NamespaceAll,
		})
	if err != nil {
		return nil, err
	}

	return &reader, nil
}
