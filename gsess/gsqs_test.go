package gsess_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/mchirico/gsa/gsess"

	"fmt"
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

		Context("Send Msg and Receive", func() {
			It("should return url", func() {

				messageID, err := gsa.SendSQS(qName, 0, msgAttrib, msgBody)
				Expect(err).To(BeNil())
				Expect(messageID).To(ContainSubstring("-"))

				rmsgBody, rmsgStr, m, err := gsa.ReceiveSQS(qName)
				Expect(err).To(BeNil())

				fmt.Printf("%v, %v, %v\n", rmsgBody, rmsgStr, m)
				Expect(rmsgBody).To(ContainSubstring(msgBody))

			})
		})

	})

})
