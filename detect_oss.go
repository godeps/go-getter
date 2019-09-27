package getter

import (
	"fmt"
	"strings"
)

// GCSDetector implements Detector to detect GCS URLs and turn
// them into URLs that the GCSGetter can understand.
type OSSDetector struct{}

func (d *OSSDetector) Detect(src, _ string) (string, bool, error) {
	if len(src) == 0 {
		return "", false, nil
	}

	if strings.HasPrefix(src,"oss::") {
		return src, true, nil
	}

	if strings.HasPrefix(src, "oss://") || (strings.HasPrefix(src, "oss-") && strings.Contains(src, "aliyuncs.com")) {
		return d.detectHTTP(src)
	}

	return "", false, nil
}

func (d *OSSDetector) detectHTTP(src string) (string, bool, error) {
	// oss://oss-cn-hangzhou.aliyuncs.com/test-bucket/hello.txt?access_key_id=KEYID&access_key_secret=SECRETKEY"
	cases := []string {"http://", "https://"}
	for _, item := range cases {
		if strings.Contains(src, item) {
			src = strings.Replace(src, item, "oss://", 1)
			break
		}
	}
	if !strings.Contains(src, "oss://")  {
		src = "oss://" + src
	}
	parts := strings.Split(src, "/")
	if len(parts) < 5 {
		return "", false, fmt.Errorf(
			"URL is not a valid GCS URL")
	}
	//bucket := parts[3]
	//object := strings.Join(parts[4:], "/")

	return "oss::" + src, true, nil
}
