package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestElasticSearchJobRepository_Add(t *testing.T) {
	type args struct {
		ctx context.Context
		job Job
	}
	tests := []struct {
		name    string
		r       JobRepository
		args    args
		wantErr bool
	}{
		{"index not initialized", newElasticSearchJobRepository(&mockRepository{initFn: func() error { return errors.New("index not initialized") }}, ""), args{context.TODO(), Job{}}, true},
		{"error on add", newElasticSearchJobRepository(&mockRepository{initFn: func() error { return nil }, addFn: func() error { return errors.New("error on add") }}, ""), args{context.TODO(), Job{}}, true},
		{"success", newElasticSearchJobRepository(&mockRepository{initFn: func() error { return nil }, addFn: func() error { return nil }}, ""), args{context.TODO(), Job{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.Add(tt.args.ctx, tt.args.job); (err != nil) != tt.wantErr {
				t.Errorf("ElasticSearchJobRepository.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestElasticSearchJobRepository_Search(t *testing.T) {
	type args struct {
		ctx        context.Context
		content    string
		city       string
		sortingAsc bool
	}
	tests := []struct {
		name    string
		r       JobRepository
		args    args
		want    []Job
		wantErr bool
	}{
		{"success", newElasticSearchJobRepository(&mockRepository{searchFn: func() ([]json.RawMessage, error) { return []json.RawMessage{}, nil }}, ""), args{context.TODO(), "aaa", "bbb", false}, []Job{}, false},
		{"error", newElasticSearchJobRepository(&mockRepository{searchFn: func() ([]json.RawMessage, error) { return nil, errors.New("error") }}, ""), args{context.TODO(), "aaa", "", false}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.Search(tt.args.ctx, tt.args.content, tt.args.city, tt.args.sortingAsc)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElasticSearchJobRepository.Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ElasticSearchJobRepository.Search() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toJobs(t *testing.T) {
	tests := []struct {
		name    string
		args    []json.RawMessage
		want    []Job
		wantErr bool
	}{
		{"success", []json.RawMessage{jobRawJSONExample()}, []Job{jobExample()}, false},
		{"invalid json", []json.RawMessage{json.RawMessage([]byte("{"))}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toJobs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("toJobs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toJobs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func jobRawJSONExample() json.RawMessage {
	body := `{"title":"Assistente de Contabilidade","description":"<li> Realizar classificação, conciliação e lançamento contábil e participar na apuração de impostos e preenchimento de guias de recolhimento junto aos órgãos do governo. Controlar escrituração de livros fiscais e auxiliar na elaboração de balancetes e demonstrativos de contabilidade.</li>","salario":1500,"cidade":["Canoas"],"cidadeFormated":["Canoas - RS (1)"]}`
	return json.RawMessage([]byte(body))
}

func jobExample() Job {
	return Job{Title: "Assistente de Contabilidade",
		Description: "<li> Realizar classificação, conciliação e lançamento contábil e participar na apuração de impostos e preenchimento de guias de recolhimento junto aos órgãos do governo. Controlar escrituração de livros fiscais e auxiliar na elaboração de balancetes e demonstrativos de contabilidade.</li>", Salary: 1500, City: []string{"Canoas"}, CityFormatted: []string{"Canoas - RS (1)"}}
}

type mockRepository struct {
	initFn   func() error
	addFn    func() error
	searchFn func() ([]json.RawMessage, error)
}

func (r mockRepository) InitIndex(ctx context.Context, name, mapping string) error {
	return r.initFn()
}
func (r mockRepository) Add(ctx context.Context, index string, content Indexable) error {
	return r.addFn()
}
func (r mockRepository) Search(ctx context.Context, index string, sort *Sort, queries ...Query) ([]json.RawMessage, error) {
	return r.searchFn()
}
