// Copyright 2017 Canonical Ltd.
// Licensed under the LGPLv3, see LICENCE file for details.

package ase

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/storage"
)

func TestServer(t *testing.T) {
	err := Start()
	if err != nil {
		t.Error(err)
	}
	sc, err := storage.NewBasicClient(storage.StorageEmulatorAccountName, "")
	if err != nil {
		t.Error(err)
	}
	bs := sc.GetBlobService()
	// PUT
	err = bs.CreateBlockBlobFromReader("testc", "test", 5, strings.NewReader("hello"), nil)
	if err != nil {
		t.Error(err)
	}
	b, err := bs.GetBlob("testc", "test")
	if err != nil {
		t.Error(err)
	}
	s, err := ioutil.ReadAll(b)
	if string(s) != "hello" {
		t.Error(`expected to read "hello" from test blob`)
	}
	err = bs.DeleteBlob("testc", "test", nil)
	if err != nil {
		t.Error(err)
	}
	_, err = bs.GetBlob("testc", "test")
	if err == nil {
		t.Error("expected error reading a deleted test blob")
	}
	err = bs.CreateBlockBlobFromReader("testc", "test2/t", 6, strings.NewReader(" world"), nil)
	if err != nil {
		t.Error(err)
	}
	b, err = bs.GetBlob("testc", "test2/t")
	if err != nil {
		t.Error(err)
	}
	s, err = ioutil.ReadAll(b)
	if string(s) != " world" {
		t.Error(`expected to read "world" from test blob`)
	}
}
