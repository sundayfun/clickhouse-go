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

package std

import (
	"encoding/json"
	"github.com/sundayfun/clickhouse-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type Releases struct {
	Version string
}

type Repository struct {
	URL      string `json:"url"`
	Releases []Releases
}

type Achievement struct {
	Name string
}
type Account struct {
	Id            uint32
	Name          string
	Organizations []string `json:"orgs"`
	Repositories  []Repository
	Achievement   Achievement
}

type GithubEvent struct {
	Title        string
	Type         string
	Assignee     Account  `json:"assignee"`
	Labels       []string `json:"labels"`
	Contributors []Account
	// should not be exported
	createdAt string
}

var testDate, _ = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2022-05-25 17:20:57 +0100 WEST")

func toJson(obj interface{}) string {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return "unable to marshal"
	}
	return string(bytes)
}

func TestStdJson(t *testing.T) {
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{"127.0.0.1:9000"},
	})
	if err := checkMinServerVersion(conn, 22, 6, 1); err != nil {
		t.Skip(err.Error())
		return
	}
	conn.Close()
	conn = clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{"127.0.0.1:9000"},
		Settings: clickhouse.Settings{
			"allow_experimental_object_type": 1,
		},
	})
	conn.Exec("DROP TABLE json_std_test")
	const ddl = `
		CREATE TABLE json_std_test (
			  event JSON
		) Engine Memory
		`
	defer func() {
		conn.Exec("DROP TABLE json_std_test")
	}()
	_, err := conn.Exec(ddl)
	require.NoError(t, err)
	scope, err := conn.Begin()
	require.NoError(t, err)
	batch, err := scope.Prepare("INSERT INTO json_std_test")
	require.NoError(t, err)
	col1Data := GithubEvent{
		Title: "Document JSON support",
		Type:  "Issue",
		Assignee: Account{
			Id:            1244,
			Name:          "Geoff",
			Achievement:   Achievement{Name: "Mars Star"},
			Repositories:  []Repository{{URL: "https://github.com/ClickHouse/clickhouse-python", Releases: []Releases{{Version: "1.0.0"}, {Version: "1.1.0"}}}, {URL: "https://github.com/ClickHouse/clickhouse-go", Releases: []Releases{{Version: "2.0.0"}, {Version: "2.1.0"}}}},
			Organizations: []string{"Support Engineer", "Integrations"},
		},
		Labels: []string{"Help wanted"},
		Contributors: []Account{
			{Id: 2244, Achievement: Achievement{Name: "Adding JSON to go driver"}, Organizations: []string{"Support Engineer", "Consulting", "PM", "Integrations"}, Name: "Dale", Repositories: []Repository{{URL: "https://github.com/ClickHouse/clickhouse-go", Releases: []Releases{{Version: "2.0.0"}, {Version: "2.1.0"}}}, {URL: "https://github.com/grafana/clickhouse", Releases: []Releases{{Version: "1.2.0"}, {Version: "1.3.0"}}}}},
			{Id: 2344, Achievement: Achievement{Name: "Managing S3 buckets"}, Organizations: []string{"Support Engineer", "Consulting"}, Name: "Melyvn", Repositories: []Repository{{URL: "https://github.com/ClickHouse/support", Releases: []Releases{{Version: "1.0.0"}, {Version: "2.3.0"}, {Version: "2.4.0"}}}}},
		},
	}
	_, err = batch.Exec(col1Data)
	require.NoError(t, err)
	require.NoError(t, scope.Commit())
	// must pass interface{} - maps must be strongly typed so map[string]interface{} wont work - it wont convert
	var event interface{}
	rows := conn.QueryRow("SELECT * FROM json_std_test")
	require.NoError(t, rows.Scan(&event))
	assert.JSONEq(t, toJson(col1Data), toJson(event))
	// again pass interface{} for anthing other than primitives
	rows = conn.QueryRow("SELECT event.assignee.Achievement FROM json_std_test")
	var achievement interface{}
	require.NoError(t, rows.Scan(&achievement))
	assert.JSONEq(t, toJson(col1Data.Assignee.Achievement), toJson(achievement))
	rows = conn.QueryRow("SELECT event.assignee.Repositories FROM json_std_test")
	var repositories interface{}
	require.NoError(t, rows.Scan(&repositories))
	assert.JSONEq(t, toJson(col1Data.Assignee.Repositories), toJson(repositories))
}
