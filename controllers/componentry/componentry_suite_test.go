package componentry_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestComponentry(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Componentry Suite")
}
