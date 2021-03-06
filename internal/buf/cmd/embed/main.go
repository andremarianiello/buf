// Copyright 2020 Buf Technologies Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/format"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"

	"github.com/bufbuild/buf/internal/pkg/app"
	"github.com/bufbuild/buf/internal/pkg/storage/storagepath"
	"github.com/bufbuild/buf/internal/pkg/util/utilstring"
)

const sliceLength = math.MaxInt64

var errUsage = errors.New("usage: embed path/to/dir package [ext]")

func main() {
	app.Main(context.Background(), run)
}

func run(ctx context.Context, container app.Container) error {
	numArgs := container.NumArgs()
	if numArgs != 3 && numArgs != 4 {
		return errUsage
	}
	absDirPath, err := getAbsDirPath(container.Arg(1))
	if err != nil {
		return err
	}
	ext := ""
	if numArgs == 4 {
		ext = container.Arg(3)
	}
	storageFilePathMap, err := getStorageFilePathMap(absDirPath, ext)
	if err != nil {
		return err
	}
	storageFilePathToData, err := getStorageFilePathToData(absDirPath, storageFilePathMap)
	if err != nil {
		return err
	}
	golangFileData, err := getGolangFileData(storageFilePathMap, storageFilePathToData, container.Arg(2))
	if err != nil {
		return err
	}
	_, err = container.Stdout().Write(golangFileData)
	return err
}

func getAbsDirPath(dirPath string) (string, error) {
	absDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		return "", err
	}
	fileInfo, err := os.Stat(absDirPath)
	if err != nil {
		return "", err
	}
	if !fileInfo.IsDir() {
		return "", fmt.Errorf("expected %q to be a directory", absDirPath)
	}
	return absDirPath, nil
}

func getStorageFilePathMap(absDirPath string, ext string) (map[string]struct{}, error) {
	storageFilePathMap := make(map[string]struct{})
	if walkErr := filepath.Walk(
		absDirPath,
		func(absFilePath string, fileInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if fileInfo.Mode().IsRegular() {
				if ext != "" && filepath.Ext(absFilePath) == ext {
					relFilePath, err := filepath.Rel(absDirPath, absFilePath)
					if err != nil {
						return err
					}
					storageFilePath, err := storagepath.NormalizeAndValidate(relFilePath)
					if err != nil {
						return err
					}
					if _, ok := storageFilePathMap[storageFilePath]; ok {
						return fmt.Errorf("duplicate file: %v", storageFilePath)
					}
					storageFilePathMap[storageFilePath] = struct{}{}
				}
			}
			return nil
		},
	); walkErr != nil {
		return nil, walkErr
	}
	return storageFilePathMap, nil
}

func getStorageFilePathToData(absDirPath string, storageFilePathMap map[string]struct{}) (map[string][]byte, error) {
	storageFilePathToData := make(map[string][]byte, len(storageFilePathMap))
	for storageFilePath := range storageFilePathMap {
		data, err := ioutil.ReadFile(filepath.Join(absDirPath, storagepath.Unnormalize(storageFilePath)))
		if err != nil {
			return nil, err
		}
		storageFilePathToData[storageFilePath] = data
	}
	return storageFilePathToData, nil
}

func getGolangFileData(
	storageFilePathMap map[string]struct{},
	storageFilePathToData map[string][]byte,
	packageName string,
) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	_, _ = buffer.WriteString(`// Code generated by embed. DO NOT EDIT.

package `)
	_, _ = buffer.WriteString(packageName)
	_, _ = buffer.WriteString(`

import (
	"github.com/bufbuild/buf/internal/pkg/storage"
	"github.com/bufbuild/buf/internal/pkg/storage/storagemem"
)

var (
	// ReadBucket is the storage.ReadBucket with the static data generated for this package.
	ReadBucket storage.ReadBucket

	pathToData = map[string][]byte{
`)

	for _, storageFilePath := range utilstring.MapToSortedSlice(storageFilePathMap) {
		_, _ = buffer.WriteString(`"`)
		_, _ = buffer.WriteString(storageFilePath)
		_, _ = buffer.WriteString(`": {
`)
		data := storageFilePathToData[storageFilePath]
		for len(data) > 0 {
			n := sliceLength
			if n > len(data) {
				n = len(data)
			}
			accum := ""
			for _, elem := range data[:n] {
				accum += fmt.Sprintf("0x%02x,", elem)
			}
			_, _ = buffer.WriteString(accum)
			_, _ = buffer.WriteString("\n")
			data = data[n:]
		}
		_, _ = buffer.WriteString(`},
`)
	}
	_, _ = buffer.WriteString(`}
)

func init() {
	readBucket, err := storagemem.NewImmutableReadBucket(pathToData)
	if err != nil {
		panic(err.Error())
	}
	ReadBucket = readBucket
}`)

	return format.Source(buffer.Bytes())
}
