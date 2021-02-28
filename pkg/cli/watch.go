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

		if _, ok := resourceMap[gv.String()]; !ok {
			resourceMap[gv.String()] = []schema.GroupVersionKind{}
		}

		if _, ok := resourceMap[gv.Group]; !ok {
			resourceMap[gv.Group] = []schema.GroupVersionKind{}
		}

		for _, k := range r.APIResources {
			if _, ok := resourceMap[k.Kind]; !ok {
				resourceMap[k.Kind] = []schema.GroupVersionKind{}
			}

			gvk := gv.WithKind(k.Kind)

			resourceMap[k.Kind] = append(resourceMap[k.Kind], gvk)
			resourceMap[gv.String()] = append(resourceMap[gv.String()], gvk)
			resourceMap[gv.Group] = append(resourceMap[gv.Group], gvk)
		}
	}

	return resourceMap, nil
}

func WatchMatchingResources(r *k.Reader, matches []string) error {
	resourceMap, _ := BuildResourceMapping(r)

	for _, a := range matches {
		resources, ok := resourceMap[a]
		if !ok || len(resources) == 0 {
			fmt.Fprintf(os.Stderr, "%s: no API group or kind match for %q\n", Progname, a)
			continue
		}

		for _, gvk := range resources {
			fmt.Printf("%s: informing on '%s'\n", Progname, gvk)
			_, _ = r.Cache.GetInformerForKind(context.Background(), gvk)
		}
	}

	return nil
}

// NewWatcher ...
func NewWatcher() *cobra.Command {
	cmd := cobra.Command{
		Use:   Progname,
		Short: "Watch stuff change in Kubernetes",
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
