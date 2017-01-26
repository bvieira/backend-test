package jobs

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestJobsService_Search(t *testing.T) {
	type args struct {
		ctx        context.Context
		content    string
		city       string
		sortingAsc bool
	}
	tests := []struct {
		name    string
		s       JobsService
		args    args
		want    []Job
		wantErr bool
	}{
		{"error", JobsService{&mockJobRepository{searchFn: func() ([]Job, error) { return nil, errors.New("error on search") }}}, args{context.TODO(), "a", "b", true}, nil, true},
		{"success", JobsService{&mockJobRepository{searchFn: func() ([]Job, error) { return []Job{Job{}}, nil }}}, args{context.TODO(), "a", "b", false}, []Job{Job{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.Search(tt.args.ctx, tt.args.content, tt.args.city, tt.args.sortingAsc)
			if (err != nil) != tt.wantErr {
				t.Errorf("JobsService.Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JobsService.Search() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJobsService_Add(t *testing.T) {
	type args struct {
		ctx  context.Context
		jobs []Job
	}
	tests := []struct {
		name    string
		s       JobsService
		args    args
		wantErr bool
	}{
		{"no jobs error", JobsService{&mockJobRepository{}}, args{context.TODO(), nil}, true},
		{"add error", JobsService{&mockJobRepository{addFn: func() error { return errors.New("error on add") }}}, args{context.TODO(), []Job{Job{}}}, true},
		{"success", JobsService{&mockJobRepository{addFn: func() error { return nil }}}, args{context.TODO(), []Job{Job{}}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.Add(tt.args.ctx, tt.args.jobs); (err != nil) != tt.wantErr {
				t.Errorf("JobsService.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type mockJobRepository struct {
	addFn    func() error
	searchFn func() ([]Job, error)
}

func (r mockJobRepository) Add(ctx context.Context, job Job) error {
	return r.addFn()
}
func (r mockJobRepository) Search(ctx context.Context, content string, city string, sortingAsc bool) ([]Job, error) {
	return r.searchFn()
}
