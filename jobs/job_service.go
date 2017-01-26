package jobs

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/bvieira/c-jobs/jobs/config"
)

// JobsService job services, process job info
type JobsService struct {
	repository JobRepository
}

// NewJobServices contructor for default configuration
func NewJobServices() *JobsService {
	mapping, err := ioutil.ReadFile(config.Get().ElasticSearchIndexMappingPath)
	if err != nil || len(mapping) == 0 {
		panic(fmt.Errorf("could not load config on path: %v, could be missing or empty, error: %v", config.Get().ElasticSearchIndexMappingPath, err))
	}

	return &JobsService{
		repository: newElasticSearchJobRepository(newElasticSearch(config.Get().ElasticSearchServer, config.Get().ElasticSearchMaxRetry, config.Get().ElasticSearchSniff, config.Get().ElasticSearchReconnectRetryTime), string(mapping)),
	}
}

// Search searches on repository for jobs with content, city sorted by salary
func (s JobsService) Search(ctx context.Context, content string, city string, sortingAsc bool) ([]Job, error) {
	return s.repository.Search(ctx, content, city, sortingAsc)
}

// Add index jobs on repository
func (s JobsService) Add(ctx context.Context, jobs []Job) error {
	if len(jobs) <= 0 {
		return NewInvalidRequestError("jobs is empty")
	}

	for _, job := range jobs {
		if err := s.repository.Add(ctx, job); err != nil {
			return err
		}
	}
	return nil
}
