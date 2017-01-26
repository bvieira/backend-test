package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// JobRepository access and update jobs data
type JobRepository interface {
	Add(ctx context.Context, job Job) error
	Search(ctx context.Context, content string, city string, sortingAsc bool) ([]Job, error)
}

// ElasticSearchJobRepository JobRepository impl for elastic search
type ElasticSearchJobRepository struct {
	Repository  *ElasticSearch
	mapping     string
	initialized bool
	rmutex      sync.RWMutex
}

// newElasticSearchJobRepository ElasticSearchJobRepository constructor
func newElasticSearchJobRepository(elasticSearch *ElasticSearch, mapping string) JobRepository {
	return &ElasticSearchJobRepository{Repository: elasticSearch, mapping: mapping}
}

func (r *ElasticSearchJobRepository) init(ctx context.Context) error {
	r.rmutex.RLock()
	initialized := r.initialized
	r.rmutex.RUnlock()

	if !initialized {
		r.rmutex.Lock()
		defer r.rmutex.Unlock()
		if err := r.Repository.InitIndex(ctx, "jobs", r.mapping); err != nil {
			return err
		}
		r.initialized = true
	}
	return nil
}

// Add adds jobs on repository
func (r *ElasticSearchJobRepository) Add(ctx context.Context, job Job) error {
	if err := r.init(ctx); err != nil { //lazy index initialization
		return err
	}
	return r.Repository.Add(ctx, "jobs", job)
}

// Search find jobs on repository
func (r *ElasticSearchJobRepository) Search(ctx context.Context, content string, city string, sortingAsc bool) ([]Job, error) {
	var queries []Query
	if content != "" {
		queries = append(queries, Query{Value: content, Fields: []string{"title^3", "description"}, Operator: "and"})
	}
	if city != "" {
		queries = append(queries, Query{Value: city, Fields: []string{"cidade"}, Operator: "and"})
	}
	result, err := r.Repository.Search(ctx, "jobs", &Sort{Field: "salario", Ascending: sortingAsc}, queries...)
	if err != nil {
		return nil, err
	}
	return toJobs(result)
}

func toJobs(searchResult []json.RawMessage) ([]Job, error) {
	jobs := make([]Job, 0)
	for _, r := range searchResult {
		var job Job
		perr := json.Unmarshal(r, &job)
		if perr != nil {
			return nil, NewParserError(fmt.Sprintf("error mapping search result to job, message: %s", perr.Error()))
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}
