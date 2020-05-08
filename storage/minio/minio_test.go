package minio

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorage_Upload(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "YES" {
		t.Skip()
	}

	ctx := context.Background()
	client, err := New(Config{
		Endpoint:        "127.0.0.1:9000",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		SSL:             false,
		BucketName:      "test2",
		Location:        "us-east-1",
		RootPath:        "images",
	})
	assert.NoError(t, err)

	reader := strings.NewReader("Some content")
	path, err := client.UploadResized(ctx, reader, 100, 100)
	assert.NoError(t, err)

	res, err := client.Read(ctx, path)
	assert.NoError(t, err)

	readResult, err := ioutil.ReadAll(res)
	assert.NoError(t, err)
	assert.Equal(t, "Some content", string(readResult))
}
