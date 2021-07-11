package gsess

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type GSA struct {
	Sess *session.Session
}

func NewAWS() *GSA {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	gsa := &GSA{Sess: sess}
	return gsa
}

func (gsa *GSA) GetItem(bucket string, item string) error {

	buf := &aws.WriteAtBuffer{}

	sess := gsa.Sess

	downloader := s3manager.NewDownloader(sess)

	numBytes, err := downloader.Download(buf,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(item),
		})
	if err != nil {
		return err
	}

	fmt.Println("Downloaded data:\n\n", string(buf.Bytes()), "\n", numBytes, "bytes")
	return nil
}
