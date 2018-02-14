// Package s3 stores backups on AWS S3
package s3

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
	"github.com/stairlin/kargo/context"
	"github.com/stairlin/kargo/log"
	"github.com/stairlin/kargo/pkg/unit"
	"github.com/stairlin/kargo/plugin/storage"
)

const (
	name    = "s3"
	slash   = "/"
	maxKeys = 1000 // S3 max items per listing

	limit     = unit.GB * 5 // Max allowed chunk size
	chunk     = unit.MB * 250
	separator = "/"
)

func init() {
	storage.Add(name, func() storage.Storage {
		return &Store{}
	})
}

// Store is an S3 store
type Store struct {
	ID     string `toml:"id"`
	Secret string `toml:"secret"`
	Token  string `toml:"token"`
	Folder string `toml:"folder"`
	Region string `toml:"region"`
	Bucket string `toml:"bucket"`
	Debug  bool   `toml:"debug"`

	Sesh *session.Session
	S3   *s3.S3
}

func (s *Store) Name() string {
	return name
}

func (s *Store) Init() error {
	if s.Folder != "" {
		s.Folder = path.Clean(s.Folder)
	}

	// Sessions should be cached when possible, because creating a new Session
	// will load all configuration values from the environment, and config files
	// each time the Session is created.
	sesh, err := session.NewSession(&aws.Config{
		Region: aws.String(s.Region),
		Credentials: credentials.NewStaticCredentials(
			s.ID,
			s.Secret,
			s.Token,
		),
	})
	if err != nil {
		return errors.Wrap(err, "cannot create AWS session")
	}
	s.Sesh = sesh
	s.S3 = s3.New(sesh)

	// DEBUG
	if s.Debug {
		// Log every request made and its payload
		sesh.Handlers.Send.PushFront(func(r *request.Request) {
			fmt.Printf("Request: %s/%s, Payload: %s\n",
				r.ClientInfo.ServiceName, r.Operation.Name, r.Params)
		})
	}
	return nil
}

func (s *Store) Info(ctx *context.Context, key string) (os.FileInfo, error) {
	input := &s3.HeadObjectInput{
		Key:    aws.String(key),
		Bucket: aws.String(s.Bucket),
	}

	out, err := s.S3.HeadObjectWithContext(ctx, input)
	switch err := err.(type) {
	case nil:
		return s.headObjectOutputInfo(out), nil
	case awserr.Error:
		if err.Code() == s3.ErrCodeNoSuchKey {
			return nil, storage.ErrKeyNotFound
		}
	}
	return nil, errors.Wrap(err, "cannot get data from S3")
}

func (s *Store) Push(ctx *context.Context, key string, r io.Reader) error {
	// TODO: Buffer it has the minimum chunk size and start a multipart
	// upload
	f, err := ctx.CreateTempFile(r)
	if err != nil {
		return err
	}
	stat, err := f.Stat()
	if err != nil {
		return errors.Wrap(err, "cannot get temporary file stat")
	}

	if stat.Size() > int64(limit) {
		// Multipart upload
		uploader := s3manager.NewUploader(s.Sesh, func(u *s3manager.Uploader) {
			u.Concurrency = 2
			u.LeavePartsOnError = false
			u.PartSize = s3manager.MinUploadPartSize

			// Adjust PartSize until the number of parts is small enough.
			size := stat.Size()
			if size/u.PartSize >= s3manager.MaxUploadParts {
				// Calculate partition size rounded up to the nearest MB
				u.PartSize = (((size / s3manager.MaxUploadParts) >> 20) + 1) << 20
			}
		})

		input := s3manager.UploadInput{
			Key:    aws.String(key),
			Bucket: aws.String(s.Bucket),
			Body:   f,
		}
		if _, err = uploader.UploadWithContext(ctx, &input); err != nil {
			return errors.Wrap(err, "cannot upload multipart backup to S3")
		}
	} else {
		// Simple upload
		input := &s3.PutObjectInput{
			Key:    aws.String(key),
			Bucket: aws.String(s.Bucket),
			Body:   f,
		}

		_, err := s.S3.PutObjectWithContext(ctx, input)
		if err != nil {
			return errors.Wrap(err, "cannot upload data to S3")
		}
	}
	return nil
}

func (s *Store) Pull(
	ctx *context.Context, key string,
) (io.ReadCloser, os.FileInfo, error) {
	input := &s3.GetObjectInput{
		Key:    aws.String(key),
		Bucket: aws.String(s.Bucket),
	}

	out, err := s.S3.GetObjectWithContext(ctx, input)
	switch err := err.(type) {
	case nil:
		return out.Body, s.getObjectOutputInfo(out), nil
	case awserr.Error:
		if err.Code() == s3.ErrCodeNoSuchKey {
			return nil, nil, storage.ErrKeyNotFound
		}
	}
	return nil, nil, errors.Wrap(err, "cannot get data from S3")
}

func (s *Store) Walk(
	ctx *context.Context,
	filter *storage.WalkFilter,
	walkFn func(key string, f os.FileInfo, err error) error,
) {
	input := &s3.ListObjectsInput{
		Bucket:  aws.String(s.Bucket),
		MaxKeys: aws.Int64(maxKeys),
	}
	if s.Folder != "" && s.Folder != slash {
		input.Prefix = aws.String(path.Clean(s.Folder) + slash)
	}
	if filter.Prefix != "" {
		input.Prefix = aws.String(aws.StringValue(input.Prefix) + filter.Prefix)
	}

	// Fetch list
	out, err := s.S3.ListObjectsWithContext(ctx, input)
	if err != nil {
		walkFn("", nil, errors.Wrap(err, "cannot list backups from S3"))
		return
	}
	if len(out.Contents) == maxKeys {
		ctx.Warn(
			"S3 list objects max key reached. Some keys will be missing.",
			log.Int64("keys", maxKeys),
		)
	}

	var i int
	var objects []*s3.Object
	for _, o := range out.Contents {
		i++
		if (filter.Limit == 0 || i <= int(filter.Limit)) &&
			isBetween(filter, aws.TimeValue(o.LastModified).UnixNano()) &&
			matches(filter, aws.StringValue(o.Key)) {
			objects = append(objects, o)
		}
	}

	// Sort items
	sort.Sort(byModTimeDesc(objects))

	for _, o := range objects {
		info := s.objectInfo(o)
		if err := walkFn(info.Name(), info, nil); err != nil {
			return
		}
	}
}

func (s *Store) objectInfo(o *s3.Object) *info {
	path, err := filepath.Rel(s.Folder, aws.StringValue(o.Key))
	if err != nil {
		path = aws.StringValue(o.Key)
	}
	return &info{
		name:    path,
		size:    aws.Int64Value(o.Size),
		modTime: aws.TimeValue(o.LastModified),
	}
}

func (s *Store) getObjectOutputInfo(o *s3.GetObjectOutput) *info {
	path, err := filepath.Rel(s.Folder, o.String())
	if err != nil {
		path = o.String()
	}
	return &info{
		name:    path,
		size:    aws.Int64Value(o.ContentLength),
		modTime: aws.TimeValue(o.LastModified),
	}
}

func (s *Store) headObjectOutputInfo(o *s3.HeadObjectOutput) *info {
	path, err := filepath.Rel(s.Folder, o.String())
	if err != nil {
		path = o.String()
	}
	return &info{
		name:    path,
		size:    aws.Int64Value(o.ContentLength),
		modTime: aws.TimeValue(o.LastModified),
	}
}

// object wraps an S3 object to a struct that implements os.FileInfo
type info struct {
	name    string
	size    int64
	modTime time.Time
}

// base name of the file
func (i *info) Name() string {
	return i.name
}

// length in bytes for regular files; system-dependent for others
func (i *info) Size() int64 {
	return i.size
}

// file mode bits
func (i *info) Mode() os.FileMode {
	return os.ModePerm
}

// modification time
func (i *info) ModTime() time.Time {
	return i.modTime
}

// abbreviation for Mode().IsDir()
func (i *info) IsDir() bool {
	return false
}

// underlying data source (can return nil)
func (i *info) Sys() interface{} {
	return nil
}

func isBetween(f *storage.WalkFilter, t int64) bool {
	return t >= f.From && t <= f.To
}

func matches(f *storage.WalkFilter, name string) bool {
	return strings.HasPrefix(name, f.Prefix) &&
		(f.Pattern == nil || f.Pattern.MatchString(name))
}

type byModTimeDesc []*s3.Object

func (l byModTimeDesc) Len() int      { return len(l) }
func (l byModTimeDesc) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l byModTimeDesc) Less(i, j int) bool {
	a := aws.TimeValue(l[i].LastModified).UnixNano()
	b := aws.TimeValue(l[j].LastModified).UnixNano()
	return a > b
}
