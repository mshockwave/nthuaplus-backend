package storage

import(
	"testing"
	"github.com/mshockwave/nthuaplus-backend/public"
	"google.golang.org/cloud/storage"
)

func TestStorageClient(t *testing.T) {
	client, err := GetNewStorageClient()
	if err != nil || client == nil{
		t.Errorf("Error getting client: %s\n", err.Error())
		t.FailNow()
	}
	client.Close()

	client, err = GetNewStorageClient()
	if err != nil || client == nil{
		t.Errorf("Error getting client: %s\n", err.Error())
		t.FailNow()
	}
	defer client.Close()

	bucket := client.Client.Bucket(public.MAIN_STORAGE_BUCKET)
	if bucket == nil {
		t.Error("Bucket nil")
		t.FailNow()
	}

	listBucket(t, bucket)
}

func listBucket(t *testing.T, bucket *storage.BucketHandle) {
	if list, e := bucket.List(ctx, nil); e == nil {
		t.Logf("Object count: %d\n", len(list.Results))
		for _, obj := range list.Results {
			t.Logf("Object Name: %s\n", obj.Name)
		}

		t.Logf("Prefix count: %d\n", len(list.Prefixes))
		for _, prefix := range list.Prefixes {
			t.Logf("Prefix: %s\n", prefix)
		}
	}else{
		t.Error("Error listing bucket: %s\n", e.Error())
		t.FailNow()
	}
}
