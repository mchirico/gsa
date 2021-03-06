package gsess

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sqs"

	"log"
	"net/http"
	"strings"
	"time"
)

type GSA struct {
	Sess   *session.Session
	expire time.Duration
}

func NewAWS() *GSA {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	gsa := &GSA{Sess: sess, expire: 15 * time.Minute}
	return gsa
}

func (gsa *GSA) ListBuckets() (*s3.ListBucketsOutput, error) {
	sess := gsa.Sess
	svc := s3.New(sess)

	result, err := svc.ListBuckets(&s3.ListBucketsInput{})
	return result, err

}

func (gsa *GSA) CreateBucket(bucket string) error {
	sess := gsa.Sess
	svc := s3.New(sess)

	// Create the S3 Bucket
	_, err := svc.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return err
	}

	err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return err
	}

	return nil
}

func (gsa *GSA) DeleteBucket(bucket string) error {
	sess := gsa.Sess
	svc := s3.New(sess)

	_, err := svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return err
	}

	err = svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})

	if err != nil {
		return err
	}

	return nil
}

func (gsa *GSA) GetItem(bucket string, item string) (int64, string, error) {

	buf := &aws.WriteAtBuffer{}

	sess := gsa.Sess

	downloader := s3manager.NewDownloader(sess)

	numBytes, err := downloader.Download(buf,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(item),
		})
	if err != nil {
		return 0, "", err
	}

	return numBytes, string(buf.Bytes()), nil
}

func (gsa *GSA) PutItem(bucket string, item string, data string) (string, error) {

	h := md5.New()
	content := strings.NewReader(data)
	content.WriteTo(h)

	sess := gsa.Sess
	svc := s3.New(sess)

	resp, _ := svc.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(item),
	})

	md5s := base64.StdEncoding.EncodeToString(h.Sum(nil))
	resp.HTTPRequest.Header.Set("Content-MD5", md5s)

	url, err := resp.Presign(gsa.expire)
	if err != nil {
		fmt.Println("error presigning request", err)
		return "", err
	}

	req, err := http.NewRequest("PUT", url, strings.NewReader(data))
	req.Header.Set("Content-MD5", md5s)
	if err != nil {
		fmt.Println("error creating request", url)
		return "", err
	}

	defClient, err := http.DefaultClient.Do(req)
	fmt.Println(defClient, err)

	return url, nil
}

func (gsa *GSA) DeleteItem(bucket string, item string) error {

	sess := gsa.Sess
	svc := s3.New(sess)

	_, err := svc.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(bucket), Key: aws.String(item)})
	if err != nil {
		return err
	}

	err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(item),
	})
	if err != nil {
		return err
	}

	return nil

}

func (gsa *GSA) CreateSQS(qName string) (string, error) {

	sess := gsa.Sess
	svc := sqs.New(sess)

	result, err := svc.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String(qName),
		Attributes: aws.StringMap(map[string]string{
			"ReceiveMessageWaitTimeSeconds": "20",
		}),
	})
	if err != nil {
		return "", err
	}

	return aws.StringValue(result.QueueUrl), nil
}

func (gsa *GSA) SendSQS(qName string, delay int64, msgAttrib map[string]*sqs.MessageAttributeValue, msgBody string) (string, error) {

	sess := gsa.Sess
	svc := sqs.New(sess)

	resultURL, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &qName,
	})

	if err != nil {
		return "", err
	}

	result, err := svc.SendMessage(&sqs.SendMessageInput{
		DelaySeconds:      aws.Int64(delay),
		MessageAttributes: msgAttrib,
		MessageBody:       aws.String(msgBody),
		QueueUrl:          resultURL.QueueUrl,
	})

	return *result.MessageId, err

}

func (gsa *GSA) ReceiveSQS(qName string) (string, string, map[string]string, error) {

	msgBody := ""
	msgStr := ""
	m := map[string]string{}

	sess := gsa.Sess
	svc := sqs.New(sess)

	resultURL, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &qName,
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == sqs.ErrCodeQueueDoesNotExist {
			fmt.Errorf("Unable to find queue %q.", qName)
		}
		return "", "", nil, err
	}

	result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl: resultURL.QueueUrl,
		AttributeNames: aws.StringSlice([]string{
			"SentTimestamp",
		}),
		MaxNumberOfMessages: aws.Int64(1),
		MessageAttributeNames: aws.StringSlice([]string{
			"All",
		}),
		WaitTimeSeconds: aws.Int64(20),
	})

	if err != nil {
		return "", "", nil, err
	}

	fmt.Printf("Received %d messages.\n", len(result.Messages))
	if len(result.Messages) > 0 {

		msgBody, msgStr, m = MsgTakeApart(result.Messages)

		resultDelete, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      resultURL.QueueUrl,
			ReceiptHandle: result.Messages[0].ReceiptHandle,
		})

		if err != nil {
			fmt.Println("Delete Error", err)
			return "", "", nil, err
		}

		fmt.Println("Message Deleted", resultDelete)
	}

	return msgBody, msgStr, m, nil
}

func MsgTakeApart(messages []*sqs.Message) (string, string, map[string]string) {

	msgBody := ""
	msgStr := ""
	m := map[string]string{}

	for _, msg := range messages {
		msgBody = *msg.Body
		msgStr = msg.String()
		fmt.Printf("body: %s\n", *msg.Body)
		fmt.Printf("msg str: %v\n", msg.String())
		for k, v := range msg.MessageAttributes {
			fmt.Printf("key: %s  value: %v\n", k, *v.StringValue)
			m[k] = *v.StringValue
		}
	}
	return msgBody, msgStr, m
}

func (gsa *GSA) CreateInstance(keyName string) error {

	sess := gsa.Sess

	// Create EC2 service client
	svc := ec2.New(sess)

	// Specify the details of the instance that you want to create.
	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		// An Amazon Linux AMI ID for t2.micro instances in the us-west-2 region
		ImageId:      aws.String("ami-0233c2d874b811deb"),
		InstanceType: aws.String("t2.micro"),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		KeyName: &keyName,
	})

	if err != nil {
		fmt.Println("Could not create instance", err)
		return err
	}

	fmt.Println("Created instance", *runResult.Instances[0].InstanceId)

	// Add tags to the created instance
	_, errtag := svc.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{runResult.Instances[0].InstanceId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("GSAinstance"),
				Value: aws.String("gsa"),
			},
		},
	})
	if errtag != nil {
		log.Println("Could not create tags for instance", runResult.Instances[0].InstanceId, errtag)
		return errtag
	}

	fmt.Println("Successfully tagged instance")

	return nil
}
