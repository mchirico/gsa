package gsess

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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
