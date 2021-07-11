package gsess_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGsess(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gsess Suite")
}
