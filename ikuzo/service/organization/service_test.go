// nolint:gocritic
package organization_test

import (
	"context"
	"errors"
	"testing"

	"github.com/delving/hub3/ikuzo/domain"
	"github.com/delving/hub3/ikuzo/service/organization"
	"github.com/delving/hub3/ikuzo/storage/memory"
	"github.com/matryer/is"
)

func TestService_Shutdown(t *testing.T) {
	ts := memory.NewOrganizationStore()

	type fields struct {
		store organization.Store
	}

	type args struct {
		ctx context.Context
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"shutting down without error",
			fields{store: ts},
			args{ctx: context.TODO()},
			false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s, err := organization.NewService(tt.fields.store)
			if err != nil {
				t.Errorf("%s = unexpected error; %s", tt.name, err)
			}

			if err := s.Shutdown(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Service.Shutdown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Put(t *testing.T) {
	type fields struct {
		store organization.Store
	}

	type args struct {
		ctx context.Context
		org domain.Organization
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"store valid org",
			fields{store: memory.NewOrganizationStore()},
			args{
				ctx: context.TODO(),
				org: domain.Organization{ID: domain.OrganizationID("demo")},
			},
			false,
		},
		{
			"store invalid org",
			fields{store: memory.NewOrganizationStore()},
			args{
				ctx: context.TODO(),
				org: domain.Organization{ID: domain.OrganizationID("")},
			},
			true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			s, err := organization.NewService(tt.fields.store)
			if err != nil {
				t.Errorf("%s = unexpected error; %s", tt.name, err)
			}

			if err := s.Put(tt.args.ctx, tt.args.org); (err != nil) != tt.wantErr {
				t.Errorf("Service.Put() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService(t *testing.T) {
	is := is.New(t)
	ctx := context.TODO()

	svc, err := organization.NewService(memory.NewOrganizationStore())
	is.NoErr(err)

	orgs, err := svc.Filter(ctx)
	is.NoErr(err)
	is.Equal(len(orgs), 0)

	// test put
	orgID, err := domain.NewOrganizationID("demo")
	is.NoErr(err)

	err = svc.Put(ctx, domain.Organization{ID: orgID})
	is.NoErr(err)

	// should have one org
	orgs, err = svc.Filter(ctx)
	is.NoErr(err)
	is.Equal(len(orgs), 1)

	// get an org
	getOrgID, err := svc.Get(ctx, orgID)
	is.NoErr(err)
	is.Equal(orgID, getOrgID.ID)

	// delete an org
	err = svc.Delete(ctx, orgID)
	is.NoErr(err)
	orgs, err = svc.Filter(ctx)
	is.NoErr(err)
	is.Equal(len(orgs), 0)

	// org not found
	getOrgID, err = svc.Get(ctx, orgID)
	is.True(errors.Is(err, domain.ErrOrgNotFound))
}
