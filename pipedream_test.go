package pipedream

import (
	"bytes"
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

func TestFingerprintFile(t *testing.T) {
	t.Parallel()

	f, err := ioutil.TempFile(testTmp, "fingerprint")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	f.Write(testCSSFile)

	if err := f.Close(); err != nil {
		t.Fatal("failed to close css file:", err)
	}

	print, err := fingerprintFile(f.Name())
	if err != nil {
		t.Error(err)
	}

	if print != "78a359fd775d5e999ee0dc43a72dc862" {
		t.Error("print was wrong:", print)
	}
}

func TestFingerprintReader(t *testing.T) {
	t.Parallel()

	reader := bytes.NewReader(testCSSFile)

	print, err := fingerprintReader(reader)
	if err != nil {
		t.Error(err)
	}

	if print != "78a359fd775d5e999ee0dc43a72dc862" {
		t.Error("print was wrong:", print)
	}
}
