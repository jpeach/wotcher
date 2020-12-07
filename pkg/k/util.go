package k

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

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

// NewCacheForScheme returns a new informer cache that uses the given scheme.
func NewCacheForScheme(s *runtime.Scheme) (cache.Cache, error) {
	conf, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}

	mapper, err := apiutil.NewDiscoveryRESTMapper(conf)
	if err != nil {
		return nil, err
	}

	return cache.New(conf,
		cache.Options{
			Scheme: s,
			Mapper: mapper,
		},
	)

}
