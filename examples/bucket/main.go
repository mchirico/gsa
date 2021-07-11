package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/mchirico/gsa/gsess"
)

func main() {

	bucketToCreate := "cwxstat-3-test"
	gsa := gsess.NewAWS()

	result, err := gsa.ListBuckets()
	if err != nil {
		return
	}

	for _, bucket := range result.Buckets {
		if aws.StringValue(bucket.Name) == bucketToCreate {
			gsa.DeleteBucket(bucketToCreate)
			break
		}
	}

	err = gsa.CreateBucket(bucketToCreate)
	if err != nil {
		fmt.Println(err)
	}
	gsa.PutItem(bucketToCreate, "key/item1", "This is some text")
	_, data, _ := gsa.GetItem(bucketToCreate, "key/item1")
	fmt.Println(data)

}
