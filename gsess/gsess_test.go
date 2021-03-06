package gsess_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/mchirico/gsa/gsess"
)

var _ = Describe("Gsess", func() {

	gsa := gsess.NewAWS()
	bucket := "cwxstat-bozo-test"

	// BeforeEach(func() {
	// 	err := gsa.CreateBucket(bucket)
	// 	Expect(err).To(BeNil())

	// })

	// AfterEach(func() {
	// 	err := gsa.DeleteBucket(bucket)
	// 	Expect(err).To(BeNil())
	// })

	Describe("Put/Get Bucket Item", func() {

		Context("Put item", func() {
			It("should get", func() {

				data := "Test data"

				_, err := gsa.PutItem(bucket, "test/testItem", data)
				Expect(err).To(BeNil())

				_, s, _ := gsa.GetItem(bucket, "test/testItem")
				Expect(s).To(BeEquivalentTo(data))

			})
		})

	})

	Describe("Create/Delete Bucket", func() {

		Context("Create Bucket", func() {
			It("should create", func() {

				bucketName := "cwxstat-bucket-test-bozo"

				result, err := gsa.ListBuckets()
				Expect(err).To(BeNil())

				for _, bucket := range result.Buckets {
					if aws.StringValue(bucket.Name) == bucketName {
						gsa.DeleteBucket(bucketName)
					}
				}

				err = gsa.CreateBucket(bucketName)
				Expect(err).To(BeNil())

				item := "some/path/down"
				_, err = gsa.PutItem(bucketName, item, "data string")
				Expect(err).To(BeNil())

				err = gsa.DeleteItem(bucketName, item)
				Expect(err).To(BeNil())

				err = gsa.DeleteBucket(bucketName)
				Expect(err).To(BeNil())

				// Leave Created
				err = gsa.CreateBucket(bucketName)
				Expect(err).To(BeNil())

			})
		})

	})

})
