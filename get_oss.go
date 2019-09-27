package getter

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

// GCSGetter is a Getter implementation that will download a module from
// a GCS bucket.
type OSSGetter struct {
	getter
}

func (g *OSSGetter) ClientMode(u *url.URL) (ClientMode, error) {
	// Parse URL
	bucketName, objectPath,accessKeyId, accessKeySecret,err := g.parseURL(u)
	if err != nil {
		return 0, err
	}

	client, err := oss.New(u.Host, accessKeyId, accessKeySecret)
	if err != nil {
		return 0, err
	}

	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return 0, err
	}

	if strings.HasSuffix(u.Path, "/") {
		return ClientModeDir, nil
	}

	lor, err := bucket.ListObjects(oss.MaxKeys(5), oss.Prefix(objectPath))
	for _, obj := range lor.Objects {
		if strings.HasSuffix(obj.Key, "/") {
			// A directory matched the prefix search, so this must be a directory
			return ClientModeDir, nil
		} else if obj.Key != objectPath {
			// A file matched the prefix search and doesn't have the same name
			// as the query, so this must be a directory
			return ClientModeDir, nil
		}
	}

	// There are no directories or subdirectories, and if a match was returned,
	// it was exactly equal to the prefix search. So return File mode
	return ClientModeFile, nil
}

func (g *OSSGetter) Get(dst string, u *url.URL) error {
	bucketName, objectPath,accessKeyId, accessKeySecret,err := g.parseURL(u)
	if err != nil {
		return err
	}

	// Remove destination if it already exists
	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		// Remove the destination
		if err := os.RemoveAll(dst); err != nil {
			return err
		}
	}

	// Create all the parent directories
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	client, err := oss.New(u.Host, accessKeyId, accessKeySecret)
	if err != nil {
		return  err
	}

	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return err
	}

	pre := oss.Prefix(objectPath)
	marker := oss.Marker("")
	var lor oss.ListObjectsResult
	for {
		lor, err = bucket.ListObjects(oss.MaxKeys(8), marker, pre)
		if err != nil {
			return err
		}
		pre = oss.Prefix(lor.Prefix)
		marker = oss.Marker(lor.NextMarker)

		wg := &sync.WaitGroup{}
		errChan := make(chan error, len(lor.Objects))
		for _, object := range lor.Objects {
			wg.Add(1)
			go func(objectKey string) {
				defer wg.Done()
				if strings.HasSuffix(objectKey, "/") {
					return
				}
				subFile := strings.Replace(objectKey, objectPath, "", 1)
				itemDir := path.Join(dst, path.Dir(subFile))
				itemFile := path.Join(dst, subFile)

				if err := os.MkdirAll(itemDir, 0755); err != nil {
					errChan <- err
				}

				for i:=0; i < 3; i++ {
					err = bucket.GetObjectToFile(objectKey, itemFile)
					if err == nil {
						continue
					}
				}
				if err != nil {
					errChan <- err
				}

			}(object.Key)
		}

		wg.Wait()

		select {
		case err := <-errChan:
			return err
		default:
			// nothing
		}

		if !lor.IsTruncated {
			break
		}
	}

	return nil
}

func (g *OSSGetter) GetFile(dst string, u *url.URL) error {
	// Parse URL
	bucketName, objectPath,accessKeyId, accessKeySecret,err := g.parseURL(u)
	if err != nil {
		return  err
	}

	client, err := oss.New(u.Host, accessKeyId, accessKeySecret)
	if err != nil {
		return  err
	}

	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return err
	}

	return bucket.GetObjectToFile(objectPath, dst)
}

func (g *OSSGetter) parseURL(u *url.URL) (bucket, path, accessKeyId, accessKeySecret string, err error) {
    // oss://oss-cn-hangzhou.aliyuncs.com/test-bucket/hello.txt?access_key_id=KEYID&access_key_secret=SECRETKEY
	if strings.Contains(u.Host, "oss") {
		hostParts := strings.Split(u.Host, ".")
		if len(hostParts) != 3 {
			err = fmt.Errorf("URL is not a valid OSS URL")
			return
		}

		pathParts := strings.SplitN(u.Path, "/", 3)
		if len(pathParts) != 3 {
			err = fmt.Errorf("URL is not a valid OSS URL")
			return
		}
		bucket = pathParts[1]
		path = pathParts[2]

		accessKeyId = u.Query().Get("access_key_id")
		accessKeySecret = u.Query().Get("access_key_secret")
	}
	return
}

