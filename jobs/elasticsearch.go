package jobs

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"encoding/json"

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
	client *elastic.Client
}

// newElasticSearch ElasticSearch constructor
func newElasticSearch(server string, retry int, sniff bool) *ElasticSearch {
	c, err := elastic.NewClient(
		elastic.SetURL(server),
		elastic.SetMaxRetries(retry),
		elastic.SetSniff(sniff))
	if err != nil {
		panic(err) //invalid config
	}
	return &ElasticSearch{client: c}
}

// InitIndex create index if not exists
func (e ElasticSearch) InitIndex(ctx context.Context, name, mapping string) error {
	if exists, err := e.client.IndexExists(name).Do(ctx); err != nil || exists {
		return err
	}
	_, err := e.client.CreateIndex(name).Body(mapping).Do(ctx)
	return err
}

// Add add content do index
func (e ElasticSearch) Add(ctx context.Context, index string, content Indexable) error {
	_, err := e.client.Index().Index(index).Type(strings.ToLower(reflect.TypeOf(content).Name())).Id(content.ID()).BodyJson(content).Do(ctx)
	return err
}

// Search search content on index
func (e ElasticSearch) Search(ctx context.Context, index string, sort *Sort, queries ...Query) ([]json.RawMessage, error) {
	if len(queries) < 1 {
		panic(errors.New("should never call search without a query"))
	}

	s := e.client.Search(index).Query(createElasticCompoundQuery(queries...))
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
