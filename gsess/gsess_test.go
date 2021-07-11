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

	Describe("Put/Get Bucket Item", func() {

		Context("Put item", func() {
			It("should get", func() {

                data := "Test data"

				_, err := gsa.PutItem("cwxstat-test", "test/testItem", data)
				Expect(err).To(BeNil())

				_,s, _ := gsa.GetItem("cwxstat-test", "test/testItem")
				Expect(s).To(BeEquivalentTo(data))


			})
		})


	})

})
