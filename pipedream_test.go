package pipedream

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

var testCSSFile = []byte(`
#id {
	color: #333333;
}

div {
	margin: 8549%;
}
`)

var (
	// temp directory for test files to be put
	testTmp string
)

func TestMain(m *testing.M) {
	var err error
	testTmp, err = ioutil.TempDir("", "pipedream")
	if err != nil {
		fmt.Println("failed to create temp directory:", err)
		os.Exit(1)
	}

	code := m.Run()

	os.RemoveAll(testTmp)

	os.Exit(code)
}
