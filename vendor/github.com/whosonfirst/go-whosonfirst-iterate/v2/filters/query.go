package filters

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/aaronland/go-json-query"
)

// type QueryFilters implements the `Filters` interface for filtering documents using `aaronland/go-json-query` query strings.
type QueryFilters struct {
	Filters
	// Include is a `aaronland/go-json-query.QuerySet` instance containing rules that must match for a document to be considered for further processing.
	Include *query.QuerySet
	// Exclude is a `aaronland/go-json-query.QuerySet` instance containing rules that if matched will prevent a document from being considered for further processing.
	Exclude *query.QuerySet
}

// NewQueryFiltersFromURI() will return a `QueryFilters` instance derived from 'uri' which is expected
// to take the form of:
//
//	{SCHEME}://{HOST}?{PARAMETERS}
//
// Where {SCHEME} and {HOST} can be any value so long as they conform to the URI standard. {PARAMETERS} is expected to be
// zero or more of the following parameters:
// * `?include=` Zero or more `aaronland/go-json-query` query strings containing rules that must match for a document to be considered for further processing.
// * `?exclude=` Zero or more `aaronland/go-json-query`	query strings containing rules that if matched will prevent a document from being considered for further processing.
// * `?include_mode=` A valid `aaronland/go-json-query` query mode string for testing inclusion rules.
// * `?exclude_mode=` A valid `aaronland/go-json-query` query mode string for testing exclusion rules.
func NewQueryFiltersFromURI(ctx context.Context, uri string) (Filters, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()
	return NewQueryFiltersFromQuery(ctx, q)
}

// NewQueryFiltersFromQuery() will return a `QueryFilters` instance derived from 'q' which is expected
// to contain zero or more of the following parameters:
// * `?include=` Zero or more `aaronland/go-json-query` query strings containing rules that must match for a document to be considered for further processing.
// * `?exclude=` Zero or more `aaronland/go-json-query`	query strings containing rules that if matched will prevent a document from being considered for further processing.
// * `?include_mode=` A valid `aaronland/go-json-query` query mode string for testing inclusion rules.
// * `?exclude_mode=` A valid `aaronland/go-json-query` query mode string for testing exclusion rules.
func NewQueryFiltersFromQuery(ctx context.Context, q url.Values) (Filters, error) {

	f := &QueryFilters{}

	q_includes := q["include"]
	q_excludes := q["exclude"]

	includes := make([]string, len(q_includes))
	excludes := make([]string, len(q_excludes))

	for idx, q_enc := range q_includes {

		q, err := url.QueryUnescape(q_enc)

		if err != nil {
			return nil, fmt.Errorf("Failed to unescape '%s' include parameter, %w", q_enc, err)
		}

		includes[idx] = q
	}

	for idx, q_enc := range q_excludes {

		q, err := url.QueryUnescape(q_enc)

		if err != nil {
			return nil, fmt.Errorf("Failed to unescape '%s' exclude parameter, %w", q_enc, err)
		}

		excludes[idx] = q
	}

	if len(includes) > 0 {

		includes_mode := q.Get("include_mode")
		includes_qs, err := querySetFromStrings(ctx, includes_mode, includes...)

		if err != nil {
			return nil, fmt.Errorf("Failed to construct include query set, %w", err)
		}

		f.Include = includes_qs
	}

	if len(excludes) > 0 {

		excludes_mode := q.Get("exclude_mode")
		excludes_qs, err := querySetFromStrings(ctx, excludes_mode, excludes...)

		if err != nil {
			return nil, fmt.Errorf("Failed to construct exclude query set, %w", err)
		}

		f.Exclude = excludes_qs
	}

	return f, nil
}

// querySetFromStrings() will return a `query.QuerySet` instance derive from 'mode' and 'flags'.
func querySetFromStrings(ctx context.Context, mode string, flags ...string) (*query.QuerySet, error) {

	var query_flags query.QueryFlags
	query_mode := query.QUERYSET_MODE_ALL

	for _, fl := range flags {

		err := query_flags.Set(fl)

		if err != nil {
			return nil, fmt.Errorf("Failed to assign '%s' to query flags, %w", fl, err)
		}
	}

	switch mode {
	case query.QUERYSET_MODE_ALL, query.QUERYSET_MODE_ANY:
		query_mode = mode
	default:
		// pass
	}

	qs := &query.QuerySet{
		Queries: query_flags,
		Mode:    query_mode,
	}

	return qs, nil
}

// Apply() performs filtering operations against 'fh' and returns a boolean value indicating whether the record should be considered for further processing.
func (f *QueryFilters) Apply(ctx context.Context, fh io.ReadSeeker) (bool, error) {

	body, err := io.ReadAll(fh)

	if err != nil {
		return false, fmt.Errorf("Failed to read document, %w", err)
	}

	includes_qs := f.Include
	excludes_qs := f.Exclude

	if includes_qs != nil {

		matches, err := query.Matches(ctx, includes_qs, body)

		if err != nil {
			return false, fmt.Errorf("Failed to perform includes matching, %w", err)
		}

		if !matches {
			return false, nil
		}

	}

	if excludes_qs != nil {

		matches, err := query.Matches(ctx, excludes_qs, body)

		if err != nil {
			return false, fmt.Errorf("Failed to perform excludes matching, %w", err)
		}

		if matches {
			return false, nil
		}
	}

	return true, nil
}
