package cofu

import (
	"io/ioutil"
	"testing"
)

func TestSendContentToTempfile(t *testing.T) {
	b := []byte(`hogehogehoge`)

	app := NewApp()
	defer app.Close()
	app.Init()

	path, err := app.SendContentToTempfile(b)
	if err != nil {
		t.Error(err)
	}

	t.Logf("tmpfile: %s", path)

	b2, err := ioutil.ReadFile(path)
	if err != nil {
		t.Error(err)
	}

	if string(b) != string(b2) {
		t.Errorf("invalid data: %s", string(b2))
	}
}

func TestSendFileToTempfile(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error(err)
	}
	defer tmpFile.Close()

	b := []byte(`hogehogehoge`)
	_, err = tmpFile.Write(b)
	if err != nil {
		t.Error(err)
	}

	app := NewApp()
	defer app.Close()
	app.Init()

	path, err := app.SendFileToTempfile(tmpFile.Name())
	if err != nil {
		t.Error(err)
	}

	t.Logf("tmpfile: %s", path)

	b2, err := ioutil.ReadFile(path)
	if err != nil {
		t.Error(err)
	}

	if string(b) != string(b2) {
		t.Errorf("invalid data: %s", string(b2))
	}
}
