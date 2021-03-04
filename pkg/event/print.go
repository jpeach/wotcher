package event

import (
	"fmt"
	"time"

	"github.com/jpeach/wotcher/pkg/k"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"
)

func timestamp(when time.Time) string {
	return when.Format(time.RFC3339)
}

func printOp(op string, obj *unstructured.Unstructured) {
	fmt.Printf("%s %s %s %s %s\n",
		timestamp(time.Now()),
		op,
		obj.GetObjectKind().GroupVersionKind().GroupKind().Kind,
		obj.GetObjectKind().GroupVersionKind().GroupVersion(),
		k.NamespacedNameOf(obj),
	)
}

func NewPrinter() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			printOp("ADD", obj.(*unstructured.Unstructured))
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if !equality.Semantic.DeepEqual(oldObj, newObj) {
				printOp("MOD", oldObj.(*unstructured.Unstructured))
			}
		},
		DeleteFunc: func(obj interface{}) {
			printOp("DEL", obj.(*unstructured.Unstructured))
		},
	}
}
