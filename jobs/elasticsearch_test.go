package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"

	elastic "gopkg.in/olivere/elastic.v5"
)

func TestElasticSearch_InitIndex(t *testing.T) {
	type args struct {
		ctx     context.Context
		name    string
		mapping string
	}
	tests := []struct {
		name    string
		e       *ElasticSearch
		args    args
		wantErr bool
	}{
		{"no client error", &ElasticSearch{}, args{context.TODO(), "jobs", "{}"}, true},
		{"elastic index not found error", &ElasticSearch{elasticClient: mockElasticClient(map[string]responseMock{"HEAD  /jobs": {nil, errors.New("index not found")}})}, args{context.TODO(), "jobs", "{}"}, true},
		{"index exists", &ElasticSearch{elasticClient: mockElasticClient(map[string]responseMock{"HEAD /jobs": {newResponse(200, "{}"), nil}})}, args{context.TODO(), "jobs", "{}"}, false},
		{"error creating index", &ElasticSearch{elasticClient: mockElasticClient(map[string]responseMock{"HEAD /jobs": {newResponse(404, "{}"), nil}, "PUT /jobs": {nil, errors.New("error on elastic")}})}, args{context.TODO(), "jobs", "{}"}, true},
		{"success", &ElasticSearch{elasticClient: mockElasticClient(map[string]responseMock{"HEAD /jobs": {newResponse(404, "{}"), nil}, "PUT /jobs": {newResponse(200, "{}"), nil}})}, args{context.TODO(), "jobs", "{}"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.e.InitIndex(tt.args.ctx, tt.args.name, tt.args.mapping); (err != nil) != tt.wantErr {
				t.Errorf("ElasticSearch.InitIndex() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestElasticSearch_Add(t *testing.T) {
	type args struct {
		ctx     context.Context
		index   string
		content Indexable
	}
	tests := []struct {
		name    string
		e       *ElasticSearch
		args    args
		wantErr bool
	}{
		{"no client error", &ElasticSearch{}, args{context.TODO(), "jobs", Job{}}, true},
		{"elastic search error", &ElasticSearch{elasticClient: mockElasticClient(map[string]responseMock{})}, args{context.TODO(), "jobs", Job{}}, true},
		{"success", &ElasticSearch{elasticClient: mockElasticClient(map[string]responseMock{"PUT /jobs/job/": {newResponse(200, "{}"), nil}})}, args{context.TODO(), "jobs", Job{Title: "a", Description: "desc", Salary: 10}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.e.Add(tt.args.ctx, tt.args.index, tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("ElasticSearch.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestElasticSearch_Search(t *testing.T) {
	type args struct {
		ctx     context.Context
		index   string
		sort    *Sort
		queries []Query
	}
	tests := []struct {
		name    string
		e       *ElasticSearch
		args    args
		want    []json.RawMessage
		wantErr bool
	}{
		{"no query error", nil, args{context.TODO(), "jobs", nil, []Query{}}, nil, true},
		{"no client error", &ElasticSearch{}, args{context.TODO(), "jobs", nil, []Query{Query{"something", []string{"field"}, "and"}}}, nil, true},
		{"elastic connection error", &ElasticSearch{elasticClient: mockElasticClient(map[string]responseMock{"POST /jobs/_search": {nil, errors.New("error on elastic")}})}, args{context.TODO(), "jobs", nil, []Query{Query{"something", []string{"field"}, "and"}}}, nil, true},
		{"success", &ElasticSearch{elasticClient: mockElasticClient(map[string]responseMock{"POST /jobs/_search": {newResponse(200, successElasticResponseBody()), nil}})}, args{context.TODO(), "jobs", &Sort{"field1", false}, []Query{Query{"something", []string{"field"}, "and"}}}, []json.RawMessage{successRawJSON()}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.e.Search(tt.args.ctx, tt.args.index, tt.args.sort, tt.args.queries...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ElasticSearch.Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ElasticSearch.Search() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createElasticCompoundQuery(t *testing.T) {
	tests := []struct {
		name string
		args []Query
		want elastic.Query
	}{
		{"single query", []Query{Query{"something", []string{"field"}, "and"}}, elastic.NewSimpleQueryStringQuery("something").DefaultOperator("and").Field("field").AnalyzeWildcard(true)},
		{"multiple query", []Query{Query{"something", []string{"field"}, "and"}, Query{"another thing", []string{"field2"}, "and"}}, elastic.NewBoolQuery().Must(elastic.NewSimpleQueryStringQuery("something").DefaultOperator("and").Field("field").AnalyzeWildcard(true)).Must(elastic.NewSimpleQueryStringQuery("another thing").DefaultOperator("and").Field("field2").AnalyzeWildcard(true))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createElasticCompoundQuery(tt.args...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createElasticCompoundQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createElasticQuery(t *testing.T) {
	tests := []struct {
		name string
		arg  Query
		want elastic.Query
	}{
		{"simple query string", Query{"something", []string{"field"}, "and"}, elastic.NewSimpleQueryStringQuery("something").DefaultOperator("and").Field("field").AnalyzeWildcard(true)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createElasticQuery(tt.arg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createElasticQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

type responseMock struct {
	resp *http.Response
	err  error
}

type mockRoundTripper struct {
	fn func(r *http.Request) (*http.Response, error)
}

func (s *mockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return s.fn(r)
}

func mockElasticClient(resp map[string]responseMock) *elastic.Client {
	c, _ := elastic.NewSimpleClient(elastic.SetHttpClient(mockHTTPClient(resp)))
	return c
}

func mockHTTPClient(resp map[string]responseMock) *http.Client {
	fn := func(r *http.Request) (*http.Response, error) {
		path := r.Method + " " + r.URL.Path
		for k, v := range resp {
			if strings.HasPrefix(path, k) {
				return v.resp, v.err
			}
		}
		return nil, fmt.Errorf("unmapped request for uri: %s", path)
	}
	return &http.Client{Transport: &mockRoundTripper{fn: fn}, CheckRedirect: nil, Jar: nil, Timeout: 0}
}

func newResponse(status int, body string) *http.Response {
	return &http.Response{
		Status:     strconv.Itoa(status),
		StatusCode: status,
		Body:       &dummyReadCloser{strings.NewReader(body)},
		Header:     http.Header{},
	}
}

type dummyReadCloser struct {
	body *strings.Reader
}

func (d *dummyReadCloser) Read(p []byte) (n int, err error) {
	return d.body.Read(p)
}

func (d *dummyReadCloser) Close() error {
	return nil
}

func successElasticResponseBody() string {
	return `{"took":199,"timed_out":false,"_shards":{"total":5,"successful":5,"failed":0},"hits":{"total":1,"max_score":6.9871044,"hits":[{"_index":"jobs","_type":"job","_id":"5446c3eae70df005eb555870d7e7c7a9138b3d80","_score":6.9871044,"_source":{"title":"Assistente de Contabilidade","description":"<li> Realizar classificação, conciliação e lançamento contábil e participar na apuração de impostos e preenchimento de guias de recolhimento junto aos órgãos do governo. Controlar escrituração de livros fiscais e auxiliar na elaboração de balancetes e demonstrativos de contabilidade.</li>","salario":1500,"cidade":["Canoas"],"cidadeFormated":["Canoas - RS (1)"]}}]}}`
}

func successRawJSON() json.RawMessage {
	body := `{"title":"Assistente de Contabilidade","description":"<li> Realizar classificação, conciliação e lançamento contábil e participar na apuração de impostos e preenchimento de guias de recolhimento junto aos órgãos do governo. Controlar escrituração de livros fiscais e auxiliar na elaboração de balancetes e demonstrativos de contabilidade.</li>","salario":1500,"cidade":["Canoas"],"cidadeFormated":["Canoas - RS (1)"]}`
	return json.RawMessage([]byte(body))
}
