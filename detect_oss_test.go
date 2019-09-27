package getter

import (
	"testing"
)

func TestOSSDetector(t *testing.T) {
	cases := []struct {
		Input  string
		Output string
	}{
		{
			"oss-cn-hangzhou.aliyuncs.com/test/test.txt",
			"osss::oss://oss-cn-hangzhou.aliyuncs.com/test/test.txt",
		},
	}

	pwd := "/pwd"
	f := new(OSSDetector)
	for i, tc := range cases {
		output, ok, err := f.Detect(tc.Input, pwd)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if !ok {
			t.Fatal("not ok")
		}

		if output != tc.Output {
			t.Fatalf("%d: bad: %#v", i, output)
		}
	}
}
