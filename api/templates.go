package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/block/iterable-go/logger"
	"github.com/block/iterable-go/rate"
	"github.com/block/iterable-go/types"
)

const (
	PathTemplates        = "templates"
	maxTemplatesPageSize = 1000
	defaultTemplatesSort = "id"
)

// Templates implements /api/templates API methods.
type Templates struct {
	api *apiClient
}

// NewTemplatesApi constructs a Templates API client.
func NewTemplatesApi(apiKey string, httpClient *http.Client, logger logger.Logger, limiter rate.Limiter) *Templates {
	return &Templates{
		api: newApiClient(apiKey, httpClient, logger, limiter),
	}
}

// Get retrieves one page of project template metadata sorted by template ID.
func (t *Templates) Get(page, pageSize int) (*types.TemplatesResponse, error) {
	if page < 1 {
		return nil, fmt.Errorf("page must be at least 1")
	}
	if pageSize < 1 || pageSize > maxTemplatesPageSize {
		return nil, fmt.Errorf("page size must be between 1 and %d", maxTemplatesPageSize)
	}

	query := make(url.Values)
	query.Set("page", strconv.Itoa(page))
	query.Set("pageSize", strconv.Itoa(pageSize))
	query.Set("sort", defaultTemplatesSort)

	var response types.TemplatesResponse
	return toNilErr(&response, t.api.getJson(PathTemplates+"?"+query.Encode(), &response))
}

// All retrieves all project template metadata across every available page.
func (t *Templates) All() ([]types.Template, error) {
	var templates []types.Template
	seenNextPages := make(map[string]struct{})

	for page := 1; ; page++ {
		response, err := t.Get(page, maxTemplatesPageSize)
		if err != nil {
			return nil, fmt.Errorf("get templates page %d: %w", page, err)
		}
		templates = append(templates, response.Templates...)
		if response.NextPageUrl == "" {
			return templates, nil
		}
		if _, exists := seenNextPages[response.NextPageUrl]; exists {
			return nil, fmt.Errorf("get templates page %d: repeated next page URL", page)
		}
		seenNextPages[response.NextPageUrl] = struct{}{}
	}
}
