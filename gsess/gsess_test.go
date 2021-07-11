package gsess_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mchirico/gsa/gsess"
)

var _ = Describe("Gsess", func() {

	

	gsa := gsess.NewAWS()

	BeforeEach(func() {

	})

	Describe("Add function", func() {

		Context("Adding fist and second", func() {
			It("should be equal to expected", func() {
				err := gsa.PutItem("cwxstat-test", "test/testItem", "Test data")
				Expect(err).To(BeNil())

				_,s, _ := gsa.GetItem("cwxstat-test", "test/testItem")
				Expect(s).To(BeEquivalentTo("Test data"))


			})
		})


	})

})
