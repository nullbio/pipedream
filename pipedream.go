// Package pipedream is for serving assets in production http workloads
package pipedream

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

/*
4. On the fly compilation middlewares
5. Compiler command line
*/

func fingerprintFile(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	return fingerprintReader(f)
}

func fingerprintReader(reader io.Reader) (string, error) {
	hash := md5.New()

	_, err := io.Copy(hash, reader)
	if err != nil {
		return "", err
	}

	hex := fmt.Sprintf("%x", hash.Sum(nil))

	return hex, nil
}
