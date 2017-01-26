package jobs

import (
	"context"
	"errors"
	"log"
	"reflect"
	"strings"

	"encoding/json"

	"sync"

	"time"

	elastic "gopkg.in/olivere/elastic.v5"
)

// Indexable contant that can be indexable
type Indexable interface {
	ID() string
}

// Query represents search query
type Query struct {
	Value    string
	Fields   []string
	Operator string
}

// Sort represents search sort
type Sort struct {
	Field     string
	Ascending bool
}

// ElasticSearch impl of elastic search connection
type ElasticSearch struct {
	elasticClient *elastic.Client
	rmutex        sync.RWMutex
}

// newElasticSearch ElasticSearch constructor
func newElasticSearch(server string, retry int, sniff bool, reconnectWait int) *ElasticSearch {
	load := func() (*elastic.Client, error) {
		return elastic.NewClient(
			elastic.SetURL(server),
			elastic.SetMaxRetries(retry),
			elastic.SetSniff(sniff))
	}
	e := &ElasticSearch{}
	go e.loadClient(reconnectWait, load)
	return e
}

func (e *ElasticSearch) loadClient(retryWait int, load func() (*elastic.Client, error)) {
	log.Print("message=\"starting elasticsearch connect\" kind=elasticsearch")
	c, err := load()
	for ; err != nil; c, err = load() {
		log.Printf("message=\"error connecting on elasticsearch, retry after %d seconds\" kind=elasticsearch error=\"%s\"", retryWait, err.Error())
		time.Sleep(time.Duration(retryWait) * time.Second)
	}
	e.rmutex.Lock()
	e.elasticClient = c
	e.rmutex.Unlock()
	log.Print("message=\"connected on elasticsearch with success\" kind=elasticsearch")
}

func (e *ElasticSearch) client() *elastic.Client {
	e.rmutex.RLock()
	defer e.rmutex.RUnlock()
	return e.elasticClient
}

// InitIndex create index if not exists
func (e *ElasticSearch) InitIndex(ctx context.Context, name, mapping string) error {
	if e.client() == nil {
		return NewElasticsearchConnectError("could not connect on elastic search")
	}
	if exists, err := e.client().IndexExists(name).Do(ctx); err != nil || exists {
		return err
	}
	_, err := e.client().CreateIndex(name).Body(mapping).Do(ctx)
	return err
}

// Add add content do index
func (e *ElasticSearch) Add(ctx context.Context, index string, content Indexable) error {
	if e.client() == nil {
		return NewElasticsearchConnectError("could not connect on elastic search")
	}

	_, err := e.client().Index().Index(index).Type(strings.ToLower(reflect.TypeOf(content).Name())).Id(content.ID()).BodyJson(content).Do(ctx)
	return err
}

// Search search content on index
func (e *ElasticSearch) Search(ctx context.Context, index string, sort *Sort, queries ...Query) ([]json.RawMessage, error) {
	if len(queries) < 1 {
		panic(errors.New("should never call search without a query"))
	}
	if e.client() == nil {
		return nil, NewElasticsearchConnectError("could not connect on elastic search")
	}

	s := e.client().Search(index).Query(createElasticCompoundQuery(queries...))
	if sort != nil {
		s.Sort(sort.Field, sort.Ascending)
	}
	searchResult, err := s.Do(ctx)
	if err != nil {
		return nil, err
	}

	var result []json.RawMessage
	for _, hit := range searchResult.Hits.Hits {
		result = append(result, *hit.Source)
	}
	return result, nil
}

func createElasticCompoundQuery(queries ...Query) elastic.Query {
	if len(queries) == 1 {
		return createElasticQuery(queries[0])
	}

	query := elastic.NewBoolQuery()
	for _, q := range queries {
		query.Must(createElasticQuery(q))
	}
	return query
}

func createElasticQuery(q Query) elastic.Query {
	query := elastic.NewSimpleQueryStringQuery(q.Value)
	query.DefaultOperator(q.Operator)
	query.AnalyzeWildcard(true)
	for _, f := range q.Fields {
		query.Field(f)
	}
	return query
}
