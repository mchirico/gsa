package gsess_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/mchirico/gsa/gsess"
)

var _ = Describe("SQS", func() {

	gsa := gsess.NewAWS()
	qName := "cwt"

	msgAttrib := map[string]*sqs.MessageAttributeValue{
		"Title": &sqs.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String("The Whistler"),
		},
		"Author": &sqs.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String("John Grisham"),
		},
		"WeeksOn": &sqs.MessageAttributeValue{
			DataType:    aws.String("Number"),
			StringValue: aws.String("6"),
		},
	}

	msgBody := "Information about current NY Times fiction bestseller for week of 12/11/2016."

	Describe("SQS", func() {

		Context("Create Q", func() {
			It("should return url", func() {
				url, err := gsa.CreateSQS(qName)
				Expect(err).To(BeNil())
				Expect(url).To(ContainSubstring(qName))

			})
		})

		Context("Send Msg", func() {
			It("should return url", func() {

				messageID, err := gsa.SendSQS(qName, 0, msgAttrib, msgBody)
				Expect(err).To(BeNil())
				Expect(messageID).To(ContainSubstring("-"))

				err = gsa.ReceiveSQS(qName)
				Expect(err).To(BeNil())

			})
		})

	})

})