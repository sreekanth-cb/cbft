//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the
//  License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing,
//  software distributed under the License is distributed on an "AS
//  IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
//  express or implied. See the License for the specific language
//  governing permissions and limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/blevesearch/bleve"
)

// A PIndex represents a "physical" index or a index "partition".

const pindexPathSuffix string = ".pindex"

type PIndex struct {
	name   string
	path   string
	bindex bleve.Index
	stream Stream
}

func (pindex *PIndex) Name() string {
	return pindex.name
}

func (pindex *PIndex) Path() string {
	return pindex.path
}

func (pindex *PIndex) BIndex() bleve.Index {
	return pindex.bindex
}

func (pindex *PIndex) Stream() Stream {
	return pindex.stream
}

func NewPIndex(name, path string, indexMappingBytes []byte) (*PIndex, error) {
	indexMapping := bleve.NewIndexMapping()

	if len(indexMappingBytes) > 0 {
		if err := json.Unmarshal(indexMappingBytes, &indexMapping); err != nil {
			return nil, fmt.Errorf("error: could not parse index mapping: %v", err)
		}
	}

	bindex, err := bleve.New(path, indexMapping)
	if err != nil {
		return nil, fmt.Errorf("error: new bleve index, path: %s, err: %s",
			path, err)
	}

	return StartPIndex(name, path, bindex)
}

func OpenPIndex(name, path string) (*PIndex, error) {
	bindex, err := bleve.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error: could not open bleve index, path: %v, err: %v",
			path, err)
	}
	// TODO: Check if name matches what was stored in bleve file.
	return StartPIndex(name, path, bindex)
}

func StartPIndex(name, path string, bindex bleve.Index) (*PIndex, error) {
	pindex := &PIndex{
		name:   name,
		path:   path,
		bindex: bindex,
		stream: make(Stream),
	}
	go pindex.Run()
	return pindex, nil
}

func PIndexPath(dataDir, pindexName string) string {
	// TODO: path security checks / mapping here; ex: "../etc/pswd"
	return dataDir + string(os.PathSeparator) + pindexName + pindexPathSuffix
}

func ParsePIndexPath(dataDir, pindexPath string) (string, bool) {
	if !strings.HasSuffix(pindexPath, pindexPathSuffix) {
		return "", false
	}
	prefix := dataDir + string(os.PathSeparator)
	if !strings.HasPrefix(pindexPath, prefix) {
		return "", false
	}
	pindexName := pindexPath[len(prefix):]
	pindexName = pindexName[0 : len(pindexName)-len(pindexPathSuffix)]
	return pindexName, true
}
