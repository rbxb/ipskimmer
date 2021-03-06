package ipskimmer

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func WriteToVisitorLog(path string, addr string, time int64) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprint(f, addr, " ", time, "\n")
	return err
}

func ReadLink(path string) (string, string, error) {
	split, err := readLinkSplit(path)
	if err != nil {
		return "", "", err
	}
	return split[0], split[1], nil
}

func ReadLinkExpires(path string) (int64, error) {
	split, err := readLinkSplit(path)
	if err != nil {
		return -1, err
	}
	n, err := strconv.ParseInt(split[2], 10, 64)
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

func WriteLink(path string, resource string, key string, expires int64) error {
	buf := bytes.NewBuffer(nil)
	fmt.Fprint(buf, resource, " ", key, " ", expires)
	return os.WriteFile(path, buf.Bytes(), 0666)
}
