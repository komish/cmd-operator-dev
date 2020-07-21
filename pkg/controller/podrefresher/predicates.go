package podrefresher

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	// PodRefresherPredicateFuncs help guide the events we want the podrefresh-controller
	// to activate upon.
	PodRefresherPredicateFuncs = predicate.Funcs{
		UpdateFunc:  predicate.ResourceVersionChangedPredicate{}.Update,
		CreateFunc:  func(e event.CreateEvent) bool { return false },
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		GenericFunc: func(e event.GenericEvent) bool { return false },
	}
)
