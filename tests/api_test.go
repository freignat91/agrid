package tests

import (
	"os"
	"testing"

	"github.com/freignat91/agrid/agridapi"
)

const (
	server    = "127.0.0.1:30103"
	dataCheck = "To be, or not to be: that is the question:"
)

var api *agridapi.AgridAPI

func TestMain(m *testing.M) {
	api = agridapi.New(server)
	//api.SetLogLevel("info")
	os.Exit(m.Run())
}

func TestFileCreate(t *testing.T) {
	file, err := api.CreateFile("/test/ee.txt", make(map[string]string), "")
	if err != nil {
		t.Fatalf("CreateFile error: %v\n", err)
	}
	if _, err := file.WriteString("essai de text1\n"); err != nil {
		t.Fatalf("file.Write error: %v\n", err)
	}
	if _, err := file.WriteString("essai de text2\n"); err != nil {
		t.Fatalf("file.Write error: %v\n", err)
	}
	if _, err := file.WriteString("essai de text3\n"); err != nil {
		t.Fatalf("file.Write error: %v\n", err)
	}
	if _, err := file.Seek(9, 0); err != nil {
		t.Fatalf("file.Write error: %v\n", err)
	}
	if _, err := file.WriteString("xxxxx"); err != nil {
		t.Fatalf("file.Write error: %v\n", err)
	}
	if err := file.Sync(); err != nil {
		t.Fatalf("file.Sync error: %v\n", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("file.Close error: %v\n", err)
	}
	file2, err := api.OpenFile("/test/ee.txt", 0, "")
	if err != nil {
		t.Fatalf("OpenFile error: %v\n", err)
	}
	if _, err := file2.WriteString("essai de text5\n"); err != nil {
		t.Fatalf("file.Write error: %v\n", err)
	}
	if err := file2.Close(); err != nil {
		t.Fatalf("file.Close error: %v\n", err)
	}
	if err := api.FileRm("/test", 0, true); err != nil {
		t.Fatalf("FileRm error: %v\n", err)
	}
}

/*
//user=common, one thread, not encrypted
func TestFileCommonThread1(t *testing.T) {
	executeTest(t, 1, "", "test1KB.file")
}

//user=common, one thread, encrypted
func TestFileCommonThread1Encrypted(t *testing.T) {
	executeTest(t, 1, "test", "test1KB.file")
}

//user=common, two threads, not encrypted
func TestFileCommonThread2(t *testing.T) {
	executeTest(t, 2, "", "test1MB.file")
}

//user=common, two threads, encrypted
func TestFileCommonThread2Encrypted(t *testing.T) {
	executeTest(t, 2, "test", "test1MB.file")
}

func TestUserCreate(t *testing.T) {
	_, err := api.UserCreate("test", "tokenTest")
	if err != nil {
		t.Fatalf("UserCreate error: %v\n", err)
	}
	api.SetUser("test:tokenTest")
}

//user=test, one thread, not encrypted
func TestFileUserThread1(t *testing.T) {
	executeTest(t, 1, "", "test1KB.file")
}

//user=test, one thread, encrypted
func TestFileUserThread1Encrypted(t *testing.T) {
	executeTest(t, 1, "test", "test1KB.file")
}

//user=test, two threads, not encrypted
func TestFileUserThread2(t *testing.T) {
	executeTest(t, 2, "", "test1MB.file")
}

//user=test, two threads, encrypted
func TestFileUserThread2Encrypted(t *testing.T) {
	executeTest(t, 2, "test", "test1MB.file")
}

func TestUserRemove(t *testing.T) {
	if err := api.UserRemove("test:tokenTest", true); err != nil {
		t.Fatalf("UserRemove error: %v\n", err)
	}
}

//Generic test
func executeTest(t *testing.T, nbThread int, key string, fileName string) {
	if err := api.FileRm("/test", 0, true); err != nil {
		t.Fatalf("FileRm error: %v\n", err)
	}
	//Store file
	if _, err := api.FileStore(fileName, "/test/ws.txt", nil, nbThread, key); err != nil {
		t.Fatalf("FileStore error: %v\n", err)
	}
	//retrieve file
	time.Sleep(1000 * time.Millisecond)
	if _, _, err := api.FileRetrieve("/test/ws.txt", "/tmp/test.txt", 0, nbThread, key); err != nil {
		t.Fatalf("FileRetrieve error: %v\n", err)
	}
	//Read and verify data file
	data, err := ioutil.ReadFile("/tmp/test.txt")
	if err != nil {
		t.Fatalf("FileRetrieve file read error : %v\n", err)
	}
	if string(data[:len(dataCheck)]) != dataCheck {
		t.Fatalf("FileRetrieve file data check failed\n")
	}
	//verify list file
	list1, err := api.FileLs("/test", false)
	if err != nil {
		t.Fatalf("FileLs error: %v\n", err)
	}
	if len(list1) != 1 {
		t.Fatalf("FileLs number of file should be 1: %v\n", list1)
	}
	//Remove /test
	if err := api.FileRm("/test/ws.txt", 0, false); err != nil {
		t.Fatalf("FileRm error: %v\n", err)
	}
	//verify no file is anymore /test folder
	list2, err := api.FileLs("/test", false)
	if err != nil {
		t.Fatalf("FileLs error: %v\n", err)
	}
	if len(list2) != 0 {
		t.Fatalf("FileLs number of file should be 0: %v\n", list2)
	}
}
*/
