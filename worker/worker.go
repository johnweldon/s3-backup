package worker

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"gopkg.in/amz.v3/s3"

	"gopkg.in/johnweldon/s3backup.v0/config"
)

const (
	// FolderLayout describes the format of the backup folder name.
	FolderLayout string = "2006-01-02-150405"

	// BinaryData is the default MIME type for backups.
	BinaryData string = "application/octet-stream"
)

// Plan describes a basic backup plan.
type Plan struct {
	// Files are the list of file names to be backed up.
	Files []string
	// Settings hold basic s3 configuration.
	Settings *config.Settings

	server      *s3.S3
	bucket      *s3.Bucket
	err         error
	initialized bool
	ensured     bool
}

// NewPlan builds a Plan from a config.Settings and a list of files to
// be backed up.
func NewPlan(settings *config.Settings, files []string) *Plan {
	return &Plan{
		Files:    files,
		Settings: settings,
	}
}

// Execute performs the backup based on the constructed Plan.
func (p *Plan) Execute() error {
	if p.err != nil {
		return p.err
	}

	p.initialize()
	p.uploadFiles()
	return p.err
}

// Err returns an error if the Plan is in a failed state, or nil otherwise.
func (p *Plan) Err() error {
	return p.err
}

// Reset clears any errors and connection details, and puts the Plan
// into a clean state to attempt another Execute.
func (p *Plan) Reset() {
	p.server = nil
	p.bucket = nil
	p.err = nil
	p.initialized = false
	p.ensured = false
}

func (p *Plan) uploadFiles() {
	if p.err != nil {
		return
	}

	var err error
	defer func() {
		if err != nil {
			p.err = err
		}
	}()

	uploads := map[string]string{}
	folder := time.Now().Format(FolderLayout)
	for _, name := range p.Files {
		uploads[name] = folder + "/" + path.Base(name)
	}

	err = doUpload(p.bucket, uploads)
}

func (p *Plan) initialize() {
	if p.initialized && p.err == nil {
		return
	}

	var err error
	defer func() {
		if err != nil {
			p.initialized = false
			p.err = err
		} else {
			p.initialized = true
			p.err = nil
		}
	}()

	p.server = s3.New(p.Settings.Auth, p.Settings.Region)
	p.ensureBucket()
}

func (p *Plan) ensureBucket() {
	if p.ensured {
		return
	}

	var err error
	defer func() {
		p.ensured = err == nil
	}()

	p.bucket, err = p.server.Bucket(p.Settings.Bucket)
	if err != nil {
		p.err = err
		return
	}

	for i := 0; i < 3; i++ {
		var exists bool
		exists, err = bucketExists(p.bucket)
		if exists {
			return
		}
		err = p.bucket.PutBucket(s3.BucketOwnerFull)
	}
}

type planErrors struct {
	errors []error
}

func (e *planErrors) Err() error {
	if len(e.errors) > 0 {
		return e
	}
	return nil
}

func (e *planErrors) Error() string {
	if len(e.errors) > 0 {
		var msgs []string
		for _, err := range e.errors {
			msgs = append(msgs, err.Error())
		}
		return strings.Join(msgs, "\n")
	}
	return ""
}

func (e *planErrors) Check(name string, steps ...func() error) bool {
	success := true
	for ix, step := range steps {
		if success {
			if err := step(); err != nil {
				e.errors = append(e.errors, err)
				success = false
			}
		} else {
			e.errors = append(e.errors, fmt.Errorf("skipped step %d on upload of %q due to previous errors", ix, name))
		}
	}
	return success
}

func doUpload(bucket *s3.Bucket, uploads map[string]string) error {
	var err planErrors
	for src, dst := range uploads {
		var e error
		var rdr io.Reader
		var length int64
		err.Check(
			src,
			func() error { rdr, length, e = openSourceFile(src); return e },
			func() error { return bucket.PutReader(dst, rdr, length, BinaryData, s3.BucketOwnerRead) },
		)
	}
	return err.Err()
}

func openSourceFile(path string) (io.Reader, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	fi, err := file.Stat()
	if err != nil {
		return nil, 0, err
	}
	return file, fi.Size(), nil
}

func bucketExists(b *s3.Bucket) (bool, error) {
	_, err := b.List("", "", "", 1)
	if err != nil {
		if se, ok := err.(*s3.Error); ok {
			switch se.Code {
			case "NoSuchBucket":
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}
