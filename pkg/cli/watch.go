package cli

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jpeach/wotcher/pkg/event"
	"github.com/jpeach/wotcher/pkg/k"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
)

const Progname = "wotcher"

func BuildResourceMapping(r *k.Reader) (map[string][]schema.GroupVersionResource, error) {
	resourceMap := map[string][]schema.GroupVersionResource{}

	_, resources, err := r.Discovery.ServerGroupsAndResources()
	if err != nil {
		return nil, fmt.Errorf("failed to query API resources: %w", err)
	}

	for _, resourceList := range resources {
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			return nil, err
		}

		for _, resource := range resourceList.APIResources {
			if k.IsSubResource(resource.Name) {
				continue
			}

			// Note that we can see the same kind multiple times due to most types having
			// status subresources. Conversely we can also see different resources with
			// the same Kind name.
			gvr := gv.WithResource(resource.Name)
			log.Printf("gvr is %q", gvr)
			// Index by resource name.
			resourceMap[resource.Name] = append(resourceMap[resource.Name], gvr)
			// Index by kind name.
			resourceMap[resource.Kind] = append(resourceMap[resource.Kind], gvr)
			// Index by GroupVersion.
			resourceMap[gv.String()] = append(resourceMap[gv.String()], gvr)
			// Index by just the Group.
			resourceMap[gv.Group] = append(resourceMap[gv.Group], gvr)
			// Index by all the short names.
			for _, short := range resource.ShortNames {
				resourceMap[short] = append(resourceMap[short], gvr)
			}
		}
	}

	return resourceMap, nil
}

func InformOnMatchingResources(r *k.Reader, matches []string) error {
	resourceMap, err := BuildResourceMapping(r)
	if err != nil {
		return err
	}

	// when we build the resource mapping, the same GVK can end
	gvrSeen := map[schema.GroupVersionResource]bool{}

	for _, a := range matches {
		resources, ok := resourceMap[a]
		if !ok || len(resources) == 0 {
			fmt.Fprintf(os.Stderr, "%s: no API group or kind match for %q\n", Progname, a)
			continue
		}

		for _, gvr := range resources {
			if gvrSeen[gvr] {
				continue
			}

			fmt.Printf("%s: informing on %q\n", Progname, gvr)
			r.InformOnResource(gvr, event.NewPrinter())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// NewWatcher ...
func NewWatcher() *cobra.Command {
	cmd := cobra.Command{
		Use:           Progname,
		Short:         "Watch stuff change in Kubernetes",
		SilenceUsage:  true, // Don't emit usage on generic errors.
		SilenceErrors: true, // Don't print errors twice.
		RunE: func(cmd *cobra.Command, args []string) error {
			s := k.NewScheme(capi.AddToScheme)

			r, err := k.NewReaderForScheme(s)
			if err != nil {
				return err
			}

			stopChan := make(chan struct{}, 2)
			sigChan := make(chan os.Signal, 2)

			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigChan
				close(stopChan)
			}()

			if err := InformOnMatchingResources(r, args); err != nil {
				return err
			}

			// TODO: inform on all events and check
			// whether they are for resource types that
			// we are matching.

			// TODO: inform on customresourcedefinitions
			// resources, and check whether new CRDs match
			// the resource types we are watching. In that
			// case we should start new informers.

			r.Run(stopChan)
			return nil
		},
	}

	return &cmd
}
