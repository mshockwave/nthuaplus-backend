package storage

import (
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2"
	"golang.org/x/net/context"

	"../public"
)

const(
	THUMBNAILS_FOLDER_NAME = "thumbnails"
	APPLICATIONS_FOLDER_NAME = "applications"
)

var defaultTokenSource oauth2.TokenSource = nil
var ctx context.Context

func init(){

	ctx = context.Background()

	var err error
	defaultTokenSource, err = google.DefaultTokenSource( ctx,
		storage.ScopeFullControl,
	)
	if err != nil || defaultTokenSource == nil{
		public.LogE.Fatalf("Error getting storage token source: %s\n", err.Error())
	}
}

type StorageClient struct {
	Client *storage.Client
	Ctx	context.Context
}

func PathJoin(segs ...string) string {return public.StringJoin("/", segs...)}

func GetNewStorageClient() (*StorageClient, error) {
	client, err := storage.NewClient(ctx, cloud.WithTokenSource(defaultTokenSource))
	return &StorageClient{
		Client: client,
		Ctx: ctx,
	}, err
}

func (this *StorageClient) Close() { this.Client.Close() }

func (this *StorageClient) GetDefaultBucket() *storage.BucketHandle{ return this.Client.Bucket(public.MAIN_STORAGE_BUCKET) }

/*
func (this *StorageClient) ListDir(dir string) (*storage.ObjectList, error){
	q := &storage.Query{
		Prefix: dir,
		Delimiter: "/",
	}
	if q == nil { return nil, errors.New("Query nil") }

	bucket := this.Client.Bucket(public.MAIN_STORAGE_BUCKET)
	if bucket == nil { return nil, errors.New("Bucket nil") }

	return bucket.List(this.Ctx, q)
}
*/
