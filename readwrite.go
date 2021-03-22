package ipskimmer

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ReadVisitorLog(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func WriteToVisitorLog(path string, visitors []visitor) error {
	buf := bytes.NewBuffer(nil)
	for _, v := range visitors {
		fmt.Fprintf(buf, v.addr, " ", v.time)
	}
	return os.WriteFile(path, buf.Bytes(), 0666)
}

func ReadLink(path string) (string, string, bool, error) {
	split, err := readLinkSplit(path)
	if err != nil {
		return "", "", false, err
	}
	return split[0], split[1], split[2] == "true", nil
}

func ReadLinkExpires(path string) (int64, error) {
	split, err := readLinkSplit(path)
	if err != nil {
		return -1, err
	}
	n, err := strconv.ParseInt(split[3], 10, 64)
	if err != nil {
		return -1, err
	}
	return n, nil
}

func readLinkSplit(path string) ([]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(b), " "), nil
}

func WriteLink(path string, resource string, key string, proxy bool, expires int64) error {
	buf := bytes.NewBuffer(nil)
	fmt.Fprint(buf, resource, " ", key, " ", proxy, " ", expires)
	return os.WriteFile(path, buf.Bytes(), 0666)
}
