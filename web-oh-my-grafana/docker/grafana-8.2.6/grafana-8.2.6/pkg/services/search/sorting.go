package search

import (
	"sort"

	"github.com/grafana/grafana/pkg/services/sqlstore/searchstore"
)

var (
	SortAlphaAsc = SortOption{
		Name:        "alpha-asc",
		DisplayName: "Alphabetically (A–Z)",
		Description: "Sort results in an alphabetically ascending order",
		Index:       0,
		Filter: []SortOptionFilter{
			searchstore.TitleSorter{},
		},
	}
	SortAlphaDesc = SortOption{
		Name:        "alpha-desc",
		DisplayName: "Alphabetically (Z–A)",
		Description: "Sort results in an alphabetically descending order",
		Index:       0,
		Filter: []SortOptionFilter{
			searchstore.TitleSorter{Descending: true},
		},
	}
)

type SortOption struct {
	Name        string
	DisplayName string
	Description string
	Index       int
	MetaName    string
	Filter      []SortOptionFilter
}

type SortOptionFilter interface {
	searchstore.FilterOrderBy
}

// RegisterSortOption allows for hooking in more search options from
// other services.
func (s *SearchService) RegisterSortOption(option SortOption) {
	s.sortOptions[option.Name] = option
}

func (s *SearchService) SortOptions() []SortOption {
	opts := make([]SortOption, 0, len(s.sortOptions))
	for _, o := range s.sortOptions {
		opts = append(opts, o)
	}
	sort.Slice(opts, func(i, j int) bool {
		return opts[i].Index < opts[j].Index || (opts[i].Index == opts[j].Index && opts[i].Name < opts[j].Name)
	})
	return opts
}
