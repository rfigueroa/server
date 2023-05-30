// Copyright (c) 2023 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package build

import (
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestBuild_Engine_CleanBuilds(t *testing.T) {
	// setup types
	_buildOne := testBuild()
	_buildOne.SetID(1)
	_buildOne.SetRepoID(1)
	_buildOne.SetNumber(1)
	_buildOne.SetCreated(1)
	_buildOne.SetStatus("pending")

	_buildTwo := testBuild()
	_buildTwo.SetID(2)
	_buildTwo.SetRepoID(1)
	_buildTwo.SetNumber(2)
	_buildTwo.SetCreated(2)
	_buildTwo.SetStatus("running")

	// setup types
	_buildThree := testBuild()
	_buildThree.SetID(3)
	_buildThree.SetRepoID(1)
	_buildThree.SetNumber(3)
	_buildThree.SetCreated(1)
	_buildThree.SetStatus("success")

	_buildFour := testBuild()
	_buildFour.SetID(4)
	_buildFour.SetRepoID(1)
	_buildFour.SetNumber(4)
	_buildFour.SetCreated(5)
	_buildFour.SetStatus("running")

	_postgres, _mock := testPostgres(t)
	defer func() { _sql, _ := _postgres.client.DB(); _sql.Close() }()

	// ensure the mock expects the name query
	_mock.ExpectExec(`UPDATE "builds" SET "status"=$1,"error"=$2,"finished"=$3,"deploy_payload"=$4 WHERE created < $5 AND (status = 'running' OR status = 'pending')`).
		WithArgs("error", "msg", time.Now().UTC().Unix(), AnyArgument{}, 3).
		WillReturnResult(sqlmock.NewResult(1, 2))

	_sqlite := testSqlite(t)
	defer func() { _sql, _ := _sqlite.client.DB(); _sql.Close() }()

	err := _sqlite.CreateBuild(_buildOne)
	if err != nil {
		t.Errorf("unable to create test build for sqlite: %v", err)
	}

	err = _sqlite.CreateBuild(_buildTwo)
	if err != nil {
		t.Errorf("unable to create test build for sqlite: %v", err)
	}

	err = _sqlite.CreateBuild(_buildThree)
	if err != nil {
		t.Errorf("unable to create test build for sqlite: %v", err)
	}

	err = _sqlite.CreateBuild(_buildFour)
	if err != nil {
		t.Errorf("unable to create test build for sqlite: %v", err)
	}

	// setup tests
	tests := []struct {
		failure  bool
		name     string
		database *engine
		want     int64
	}{
		{
			failure:  false,
			name:     "postgres",
			database: _postgres,
			want:     2,
		},
		{
			failure:  false,
			name:     "sqlite3",
			database: _sqlite,
			want:     2,
		},
	}

	// run tests
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.database.CleanBuilds("msg", 3)

			if test.failure {
				if err == nil {
					t.Errorf("CleanBuilds for %s should have returned err", test.name)
				}

				return
			}

			if err != nil {
				t.Errorf("CleanBuilds for %s returned err: %v", test.name, err)
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("CleanBuilds for %s is %v, want %v", test.name, got, test.want)
			}
		})
	}
}