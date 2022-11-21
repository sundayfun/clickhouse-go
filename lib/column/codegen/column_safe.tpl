// Licensed to ClickHouse, Inc. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. ClickHouse, Inc. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// +build !amd64,!arm64

// Code generated by make codegen DO NOT EDIT.
// source: lib/column/codegen/column_safe.tpl

package column

import (
	"github.com/sundayfun/clickhouse-go/v2/lib/binary"
)


{{- range . }}

func (col *{{ .ChType }}) Decode(decoder *binary.Decoder, rows int) error {
	for i := 0; i < rows; i++ {
		v, err := decoder.{{ .ChType }}()
		if err != nil {
			return err
		}
		col.data = append(col.data, v)
	}
	return nil
}

func (col *{{ .ChType }}) Encode(encoder *binary.Encoder) error {
	for _, v := range col.data {
		if err := encoder.{{ .ChType }}(v); err != nil {
			return err
		}
	}
	return nil
}

{{- end }}