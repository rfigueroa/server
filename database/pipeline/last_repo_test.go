// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package pipeline

import (
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-vela/types/library"
)

func TestPipeline_Engine_LastPipelineForRepo(t *testing.T) {
	// setup types
	_pipelineOne := testPipeline()
	_pipelineOne.SetID(1)
	_pipelineOne.SetRepoID(1)
	_pipelineOne.SetNumber(1)
	_pipelineOne.SetRef("refs/heads/master")
	_pipelineOne.SetType("yaml")
	_pipelineOne.SetVersion("1")
	_pipelineOne.SetData([]byte("foo"))

	_pipelineTwo := testPipeline()
	_pipelineTwo.SetID(2)
	_pipelineTwo.SetRepoID(1)
	_pipelineTwo.SetNumber(2)
	_pipelineTwo.SetRef("refs/heads/main")
	_pipelineTwo.SetType("yaml")
	_pipelineTwo.SetVersion("1")
	_pipelineTwo.SetData([]byte("foo"))

	_postgres, _mock := testPostgres(t)
	defer func() { _sql, _ := _postgres.client.DB(); _sql.Close() }()

	// create expected result in mock
	_rows := sqlmock.NewRows(
		[]string{"id", "repo_id", "number", "commit", "flavor", "platform", "ref", "type", "version", "services", "stages", "steps", "templates", "data"}).
		AddRow(2, 1, 2, "", "", "", "refs/heads/main", "yaml", "1", false, false, false, false, []byte{120, 94, 74, 203, 207, 7, 4, 0, 0, 255, 255, 2, 130, 1, 69})

	// ensure the mock expects the query
	_mock.ExpectQuery(`SELECT * FROM "pipelines" WHERE repo_id = $1 ORDER BY number DESC LIMIT 1`).WithArgs(1).WillReturnRows(_rows)

	_sqlite := testSqlite(t)
	defer func() { _sql, _ := _sqlite.client.DB(); _sql.Close() }()

	err := _sqlite.CreatePipeline(_pipelineOne)
	if err != nil {
		t.Errorf("unable to create test pipeline for sqlite: %v", err)
	}

	err = _sqlite.CreatePipeline(_pipelineTwo)
	if err != nil {
		t.Errorf("unable to create test pipeline for sqlite: %v", err)
	}

	// setup tests
	tests := []struct {
		failure  bool
		name     string
		database *engine
		want     *library.Pipeline
	}{
		{
			failure:  false,
			name:     "postgres",
			database: _postgres,
			want:     _pipelineTwo,
		},
		{
			failure:  false,
			name:     "sqlite3",
			database: _sqlite,
			want:     _pipelineTwo,
		},
	}

	// run tests
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.database.LastPipelineForRepo(&library.Repo{ID: _pipelineOne.RepoID})

			if test.failure {
				if err == nil {
					t.Errorf("LastPipelineForRepo for %s should have returned err", test.name)
				}

				return
			}

			if err != nil {
				t.Errorf("LastPipelineForRepo for %s returned err: %v", test.name, err)
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("LastPipelineForRepo for %s is %v, want %v", test.name, got, test.want)
			}
		})
	}
}