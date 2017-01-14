package agridapi

import (
	"fmt"
	"os"
	"time"
)

type AFile struct {
	api       *AgridAPI
	client    *gnodeClient
	name      string
	fname     string
	key       string
	isCreated bool
	file      *os.File
	meta      map[string]string
}

func (api *AgridAPI) CreateFile(name string, key string) (*AFile, error) {
	af := AFile{}
	af.key = key
	af.api = api
	af.name = name
	af.meta = make(map[string]string)
	af.isCreated = true
	af.fname = fmt.Sprintf("/tmp/af%d", time.Now().UnixNano())
	if file, err := os.Create(af.fname); err != nil {
		return nil, err
	} else {
		af.file = file
	}
	api.info("Create file %s\n", af.name)
	return &af, nil
}

func (api *AgridAPI) OpenFile(name string, version int, key string) (*AFile, error) {
	af := AFile{}
	af.key = key
	af.api = api
	af.name = name
	af.fname = fmt.Sprintf("/tmp/af%d", time.Now().UnixNano())
	if meta, _, err := af.api.FileRetrieve(name, af.fname, version, 1, key); err != nil {
		return nil, err
	} else {
		af.meta = meta
	}
	if file, err := os.OpenFile(af.fname, os.O_RDWR, 0666); err != nil {
		return nil, err
	} else {
		af.file = file
	}
	api.info("Open file %s\n", af.name)
	return &af, nil
}

func (af *AFile) Close() error {
	af.file.Sync()
	args := []string{}
	for name, val := range af.meta {
		args = append(args, fmt.Sprintf("%s=%s", name, val))
	}
	if version, err := af.api.FileStore(af.fname, af.name, args, 1, af.key); err != nil {
		return err
	} else {
		af.api.info("file %s stored version %d\n", af.name, version)
		os.Remove(af.fname)
	}
	return nil
}

func (af *AFile) WriteString(data string) (int, error) {
	return af.Write([]byte(data))
}

func (af *AFile) Write(data []byte) (int, error) {
	return af.file.Write(data)
}

func (af *AFile) WriteStringAt(data string, at int64) (int, error) {
	return af.WriteStringAt(data, at)
}

func (af *AFile) WriteAt(data []byte, at int64) (int, error) {
	return af.WriteAt(data, at)
}

func (af *AFile) Read(data []byte) (int, error) {
	return af.file.Read(data)
}

func (af *AFile) ReadAt(data []byte, at int64) (int, error) {
	return af.file.ReadAt(data, at)
}

/*
func (af *AFile) ReadString() (string, error) {
	return "", nil
}
*/
func (af *AFile) Seek(offset int64, whence int) (int64, error) {
	return af.file.Seek(offset, whence)
}

func (af *AFile) getMetadata() map[string]string {
	return af.meta
}

func (af *AFile) Sync() error {
	return af.file.Sync()
}
