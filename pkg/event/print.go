package event

import (
	"fmt"
	"time"

	"github.com/jpeach/wotcher/pkg/k"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"
)

func timestamp(when metav1.Time) string {
	return when.Format(time.RFC3339)
}

func printOp(op string, when metav1.Time, u *unstructured.Unstructured) {
	fmt.Printf("%s %s %s %s %s\n",
		timestamp(when),
		op,
		u.GetObjectKind().GroupVersionKind().GroupKind().Kind,
		u.GetObjectKind().GroupVersionKind().GroupVersion(),
		k.NamespacedNameOf(u),
	)
}

func NewPrinter() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			u := obj.(*unstructured.Unstructured)

			// Show the object creation timestamp so
			// that events we see during the initial
			// informer sync make more sense.
			printOp("ADD", u.GetCreationTimestamp(), u)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if !equality.Semantic.DeepEqual(oldObj, newObj) {
				printOp("MOD", metav1.Now(), oldObj.(*unstructured.Unstructured))
			}
		},
		DeleteFunc: func(obj interface{}) {
			printOp("DEL", metav1.Now(), obj.(*unstructured.Unstructured))
		},
	}
}
