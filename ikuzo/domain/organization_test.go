// Copyright 2020 Delving B.V.
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

package domain

import (
	"errors"
	"testing"
)

func TestID_Valid(t *testing.T) {
	tests := []struct {
		name string
		id   OrganizationID
		want error
	}{
		{
			"valid identifier",
			OrganizationID("demo"),
			nil,
		},
		{
			"identifier too long",
			OrganizationID("demodemodemo"),
			ErrIDTooLong,
		},
		{
			"identifier must be all lowercase",
			OrganizationID("DemoOrg"),
			ErrIDNotLowercase,
		},
		{
			"identifier must not contain special characters",
			OrganizationID("demo-org"),
			ErrIDInvalidCharacter,
		},
		{
			"identifier must not be empty",
			OrganizationID(""),
			ErrIDCannotBeEmpty,
		},
		{
			"identifier cannot be reserved identifier: public",
			OrganizationID("public"),
			ErrIDExists,
		},
		{
			"identifier cannot be reserved identifier: all",
			OrganizationID("all"),
			ErrIDExists,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := tt.id.Valid()
			if !errors.Is(got, tt.want) {
				t.Errorf("ID.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewOrganizationID(t *testing.T) {
	type args struct {
		input string
	}

	tests := []struct {
		name    string
		args    args
		want    OrganizationID
		wantErr bool
	}{
		{
			"valid input",
			args{input: "demo"},
			OrganizationID("demo"),
			false,
		},
		{
			"invalid input",
			args{input: "Demo"},
			OrganizationID(""),
			true,
		},
		{
			"id cannot be empty",
			args{input: ""},
			OrganizationID(""),
			true,
		},
		{
			"id cannot be a protected name",
			args{input: "public"},
			OrganizationID(""),
			true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewOrganizationID(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOrganizationID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NewOrganizationID() = %v, want %v", got, tt.want)
			}
		})
	}
}
