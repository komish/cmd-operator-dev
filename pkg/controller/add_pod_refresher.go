package controller

import (
	"github.com/komish/certmanager-operator/pkg/controller/podrefresher"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, podrefresher.Add)
}
