package event

import (
	"fmt"
	"os"
	"time"

	"github.com/jpeach/wotcher/pkg/k"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"
)

func timestamp(when time.Time) string {
	return when.Format(time.RFC3339)
}

func printOp(op string, obj interface{}) {
	kobj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		fmt.Fprint(os.Stderr, "Failed to cast obj to metav1.Object\n")
		return
	}

	fmt.Printf("%s %s %s %s %s\n",
		timestamp(time.Now()),
		op,
		kobj.GetObjectKind().GroupVersionKind().GroupKind().Kind,
		kobj.GetObjectKind().GroupVersionKind().GroupVersion(),
		k.NamespacedNameOf(kobj),
	)
}

func NewPrinter() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			printOp("ADD", obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			printOp("MOD", oldObj)
		},
		DeleteFunc: func(obj interface{}) {
			printOp("DEL", obj)
		},
	}
}
