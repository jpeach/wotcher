package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/jpeach/wotcher/pkg/k"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
)

const Progname = "wotcher"

func BuildResourceMapping(r *k.Reader) (map[string][]schema.GroupVersionKind, error) {
	resourceMap := map[string][]schema.GroupVersionKind{}

	_, resources, err := r.Discovery.ServerGroupsAndResources()
	if err != nil {
		return nil, fmt.Errorf("failed to query API resources: %w", err)
	}

	for _, r := range resources {
		gv, err := schema.ParseGroupVersion(r.GroupVersion)
		if err != nil {
			return nil, err
		}

		for _, k := range r.APIResources {
			// Note that we can see the same kind multiple times due to most types having
			// status subresources. Conversely we can also see different resources with
			// the same Kind name.
			gvk := gv.WithKind(k.Kind)
			// Index by resource name.
			resourceMap[k.Name] = append(resourceMap[k.Name], gvk)
			// Index by kind name.
			resourceMap[k.Kind] = append(resourceMap[k.Kind], gvk)
			// Index by GroupVersion.
			resourceMap[gv.String()] = append(resourceMap[gv.String()], gvk)
			// Index by just the Group.
			resourceMap[gv.Group] = append(resourceMap[gv.Group], gvk)
		}
	}

	return resourceMap, nil
}

func WatchMatchingResources(r *k.Reader, matches []string) error {
	resourceMap, _ := BuildResourceMapping(r)

	// when we build the resource mapping, the same GVK can end
	gvkSeen := map[schema.GroupVersionKind]bool{}

	for _, a := range matches {
		resources, ok := resourceMap[a]
		if !ok || len(resources) == 0 {
			fmt.Fprintf(os.Stderr, "%s: no API group or kind match for %q\n", Progname, a)
			continue
		}

		for _, gvk := range resources {
			if gvkSeen[gvk] {
				continue
			}

			fmt.Printf("%s: informing on %q\n", Progname, gvk)
			_, err := r.Cache.GetInformerForKind(context.Background(), gvk)
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

			if err := WatchMatchingResources(r, args); err != nil {
				return err
			}

			return r.Cache.Start(context.Background())
		},
	}

	return &cmd
}
