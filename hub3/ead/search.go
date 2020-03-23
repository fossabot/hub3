package ead

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/OneOfOne/xxhash"
	"github.com/allegro/bigcache"
	c "github.com/delving/hub3/config"
	"github.com/delving/hub3/hub3/fragments"
	"github.com/delving/hub3/hub3/index"
	"github.com/delving/hub3/hub3/models"
	"github.com/delving/hub3/ikuzo/service/x/search"
	"github.com/delving/hub3/ikuzo/storage/x/memory"
	"github.com/go-chi/chi"
	"github.com/olivere/elastic/v7"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

const (
	trueParamValue = "true"
)

var httpCache *bigcache.BigCache
var once sync.Once

// SearchResponse contains the EAD Search response.
type SearchResponse struct {

	// ArchiveCount returns the number of collapsed Archives that match the search  query
	ArchiveCount int `json:"archiveCount"`

	// Cursor the location of the first result in the ElasticSearch search response
	Cursor int `json:"cursor"`

	// TotalPages is the total number of pages in the search response
	TotalPages int `json:"totalPages"`

	// TotalClevelCount returns the total number of clevel that mathc the search query
	// this counts is per clevel, so multiple hits inside a clevel are counted as one
	TotalClevelCount int `json:"totalClevelCount"`

	// TotalDescriptionCount returns the total number of hits in the description.
	// This is an cardinatility aggregation so each hit inside the decription counts as a hit.
	TotalDescriptionCount int `json:"totalDescriptionCount"`

	// TotalHits is a combination of TotalClevelCount and TotalDescriptiontCount.
	TotalHits int `json:"totalHits"`

	// Archives contains the list of archives from the response constrained by the search pagination
	Archives []Archive `json:"archives"`

	// CLevels contains a paged result of the cLevels for a specific archive that match the search query
	// It is ordered by the ead orderKey.
	CLevels []CLevelEntry `json:"cLevels,omitempty"`

	// Facets holds the QueryFacets for filtering
	Facets []*fragments.QueryFacet `json:"facets,omitempty"`

	// Explain response from elasticsearch
	Explain *elastic.SearchResult `json:"explain,omitempty"`

	// Service is the elasticsearch query
	Service interface{} `json:"service,omitempty"`
}

// CLevel holds the search results per clevel entry in the an EAD Archive.
type CLevelEntry struct {

	// Path is the unique key to the path of the clevel in the archive tree
	Path string `json:"path"`

	// UnitID is the identifier of the clevel
	UnitID string `json:"unitID"`

	// Label is the title of the clevel
	Label string `json:"label"`

	// HubID is the unique identifier of the clevel as stored in the hub3 index
	HubID string `json:"hubID"`

	// ResultOrder is the place the search result has in the total list of results.
	// This can be used to aid the search pagination on the Archive result page.
	ResultOrder uint64 `json:"sortKey"`
}

// Archive holds all information for the EAD search results that are grouped
// by inventoryID. This is the EadID from the EAD header.
type Archive struct {
	InventoryID      string   `json:"inventoryID"`
	Title            string   `json:"title"`
	Period           []string `json:"period"`
	CLevelCount      int      `json:"cLevelCount"`
	DescriptionCount int      `json:"descriptionCount"`
	Files            string   `json:"files,omitempty"`
	Length           string   `json:"length,omitempty"`
	Abstract         []string `json:"abstract,omitempty"`
	Material         string   `json:"material,omitempty"`
	Language         string   `json:"language,omitempty"`
	Origin           []string `json:"origin,omitempty"`
	MetsFiles        int      `json:"metsFiles,omitempty"`
	ClevelsTotal     int      `json:"clevelsTotal"`
}

// SearchRequest holds all information for EAD search
type SearchRequest struct {
	Page             int
	Rows             int
	Query            *elastic.BoolQuery
	RawQuery         string
	Service          *elastic.SearchService
	FacetFields      []string
	Filters          []*fragments.QueryFilter
	FacetSize        int
	FacetAndBoolType bool
	SortBy           string
	NestedSortField  string
	SortAsc          bool
	NoCache          bool
	CacheRefresh     bool
	CacheReset       bool
	InventoryID      string
}

const specField = "meta.spec"

func newSearchRequest(params url.Values) (*SearchRequest, error) {
	sr := &SearchRequest{
		Page:            1,
		Rows:            10,
		NestedSortField: "@value.keyword",
		FacetSize:       50,
		Filters:         []*fragments.QueryFilter{},
	}

	rlog := c.Config.Logger.With().
		Str("application", "hub3").
		Str("search.type", "request builder").
		Logger()

	for p, v := range params {
		switch p {
		case "rows":
			size, err := strconv.Atoi(params.Get(p))
			if err != nil {
				rlog.Error().Err(err).
					Str("param", p).
					Msgf("unable to convert %v to int", v)

				return nil, err
			}

			if size > 100 {
				size = 100
			}

			sr.Rows = size
		case "facet.size":
			size, err := strconv.Atoi(params.Get(p))
			if err != nil {
				rlog.Error().Err(err).
					Str("param", p).
					Msgf("unable to convert %v to int", v)

				return nil, err
			}

			if size > 100 {
				size = 100
			}

			sr.FacetSize = size
		case "FacetBoolType":
			fbt := params.Get(p)
			if fbt != "" {
				sr.FacetAndBoolType = strings.EqualFold(fbt, "and")
			}
		case "page":
			rawPage, err := strconv.Atoi(params.Get(p))
			if err != nil {
				rlog.Error().Err(err).
					Str("param", p).
					Msgf("unable to convert %v to int", v)

				return nil, err
			}

			if rawPage == 0 {
				err := fmt.Errorf("0 pages is not allowed. Paging starts at 1")
				rlog.Error().Err(err).
					Str("param", p).
					Msg("")

				return nil, err
			}

			sr.Page = rawPage
		case "sortBy":
			sortKey := params.Get(p)
			if strings.HasPrefix(sortKey, "^") {
				sr.SortAsc = true
				sortKey = strings.TrimPrefix(sortKey, "^")
			}

			if strings.HasPrefix(sortKey, "int.") {
				sr.NestedSortField = "integer"
				sortKey = strings.TrimPrefix(sortKey, "int.")
			}

			sr.SortBy = sortKey
		case "q", "query":
			sr.RawQuery = params.Get(p)
		case "facet.field":
			sr.FacetFields = v
		case "qf", "qf[]":
			for _, filter := range v {
				qf, err := fragments.NewQueryFilter(filter)
				if err != nil {
					rlog.Error().Err(err).
						Str("param", p).
						Msg("error in filter gerenation")

					return nil, err
				}

				sr.Filters = append(sr.Filters, qf)
			}
		case "qf.dateRange", "qf.dateRange[]":
			for _, filter := range v {
				qf, err := fragments.NewDateRangeFilter(filter)
				if err != nil {
					rlog.Error().Err(err).
						Str("param", p).
						Msg("error in daterange filter gerenation")

					return sr, err
				}

				sr.Filters = append(sr.Filters, qf)
			}
		case "noCache":
			sr.NoCache = strings.EqualFold(params.Get(p), "true")
		case "cacheRefresh":
			sr.CacheRefresh = strings.EqualFold(params.Get(p), "true")
		case "cacheReset":
			sr.CacheReset = strings.EqualFold(params.Get(p), "true")
		}
	}

	return sr, nil
}

func (sr *SearchRequest) requestKey() string {
	jsonBytes, err := json.Marshal(sr)
	if err != nil {
		c.Config.Logger.Error().Err(err).
			Msg("unable to marshal request key")

		return ""
	}

	hash := xxhash.Checksum64(jsonBytes)

	return fmt.Sprintf("%016x", hash)
}

func (sr *SearchRequest) enableDescriptionSearch() bool {
	for _, f := range sr.Filters {
		if f.SearchLabel != "ead-rdf_periodDesc" {
			return false
		}
	}

	return true
}

func buildSearchRequest(r *http.Request, includeDescription bool) (*SearchRequest, error) {
	client := index.ESClient()

	s := client.Search(c.Config.ElasticSearch.GetIndexName()).
		TrackTotalHits(c.Config.ElasticSearch.TrackTotalHits)

	sr, err := newSearchRequest(r.URL.Query())
	if err != nil {
		c.Config.Logger.Error().Err(err).
			Msg("unable to create ead.SearchRequest")

		return nil, err
	}

	s = s.Size(sr.Rows)
	if sr.Page > 1 {
		s = s.From(getCursor(sr.Rows, sr.Page))
	}

	tagQuery := elastic.NewBoolQuery().Should(elastic.NewTermQuery("meta.tags", "ead"))
	if includeDescription && sr.enableDescriptionSearch() {
		tagQuery = tagQuery.Should(elastic.NewTermQuery("meta.tags", "eadDesc"))
	}

	query := elastic.NewBoolQuery()
	query = query.Must(tagQuery)

	if sr.RawQuery != "" {
		// TODO(kiivihal): replace querystring below with search.QueryTerm
		q := elastic.NewQueryStringQuery(sr.RawQuery)

		q = q.
			FieldWithBoost("tree.title", 6.0).
			FieldWithBoost("tree.inventoryID", 3.0).
			FieldWithBoost("tree.label", 2.0).
			FieldWithBoost("tree.agencyCode", 1.5).
			FieldWithBoost("tree.unitID", 1.5).
			FieldWithBoost("tree.description", 1.0).
			Field("tree.rawContent")

		if !isAdvancedSearch(sr.RawQuery) {
			q = q.MinimumShouldMatch(c.Config.ElasticSearch.MinimumShouldMatch)
		}

		query = query.Must(q)
	}

	if r.URL.Query().Get("explain") == trueParamValue {
		s = s.Explain(true)
	}

	sr.Query = query
	sr.Service = s

	return sr, nil
}

func newBigCache() {
	config := bigcache.Config{
		Shards:           1024,
		HardMaxCacheSize: c.Config.Cache.HardMaxCacheSize,
		LifeWindow:       time.Duration(c.Config.Cache.LifeWindowMinutes) * time.Minute,
		CleanWindow:      5 * time.Minute,
		MaxEntrySize:     c.Config.Cache.MaxEntrySize,
	}

	var err error

	httpCache, err = bigcache.NewBigCache(config)
	if err != nil {
		c.Config.Logger.Warn().
			Err(err).
			Msg("cannot start bigcache running without cache; %#v")
	}

	rlog := c.Config.Logger.With().Str("test", "sublogger").Logger()
	rlog.Info().Msg("starting bigCache for request caching")
}

func getCachedRequest(requestKey string, rlog *zerolog.Logger) *SearchResponse {
	entry, cacheErr := httpCache.Get(requestKey)
	if cacheErr != nil {
		rlog.Debug().
			Str("cache_key", requestKey).
			Err(cacheErr).
			Msg("cache miss")

		return nil
	}

	var eadResponse SearchResponse

	jsonErr := json.Unmarshal(entry, &eadResponse)
	if jsonErr != nil {
		rlog.Warn().Err(jsonErr).
			Msg("unable to unmarshall cached response")

		return nil
	}

	rlog.Debug().
		Str("cache_key", requestKey).
		Msg("returning response from cache")

	return &eadResponse
}

func PerformClusteredSearch(r *http.Request) (*SearchResponse, error) {
	once.Do(newBigCache)

	requestID, _ := hlog.IDFromRequest(r)
	rlog := c.Config.Logger.With().
		Str("req_id", requestID.String()).
		Str("searchType", "ead cluster search").
		Logger()

	req, requestErr := buildSearchRequest(r, true)
	if requestErr != nil {
		rlog.Error().Err(requestErr).
			Msg("performClusteredSearch error")

		return nil, requestErr
	}

	// default facets
	req.FacetFields = append(
		req.FacetFields,
		[]string{
			"tree.hasDigitalObject",
			"tree.mimeType",
			"ead-rdf_genreform",
		}...,
	)

	if req.CacheReset {
		newBigCache()
		// already cache this request
		req.CacheReset = false
	}

	requestKey := req.requestKey()
	rlog.Debug().
		Str("cache_key", requestKey).
		Msg("generating cache request key")

	if httpCache != nil && requestKey != "" && !req.NoCache && !req.CacheRefresh {
		response := getCachedRequest(requestKey, &rlog)
		if response != nil {
			return response, nil
		}
	}

	if req.CacheRefresh {
		req.CacheRefresh = false
		requestKey = req.requestKey()
	}

	s := req.Service

	b := elastic.NewCollapseBuilder(specField).
		InnerHit(
			elastic.NewInnerHit().
				Name("collapse").
				Size(1).
				Sort("tree.inventoryID", true),
		).
		MaxConcurrentGroupRequests(4)
	s = s.Collapse(b)
	s = s.FetchSource(false)

	if req.SortBy != "" {
		switch key := req.SortBy; {
		case req.SortBy == "_score":
			s = s.Sort(req.SortBy, req.SortAsc)
		case strings.Contains(key, "_"):
			path := fmt.Sprintf("resources.entries.%s", req.NestedSortField)
			fieldSort := elastic.NewFieldSort(path).
				Order(req.SortAsc).
				Nested(
					elastic.NewNestedSort("resources.entries").
						Filter(
							elastic.NewTermQuery("resources.entries.searchLabel", key),
						),
				)
			s = s.SortBy(fieldSort)
		default:
			s = s.Sort(req.SortBy, req.SortAsc)
		}
	}

	fub, err := fragments.NewFacetURIBuilder(req.RawQuery, req.Filters)
	if err != nil {
		return nil, err
	}

	for _, facetField := range req.FacetFields {
		ff, facetErr := fragments.NewFacetField(facetField)
		if facetErr != nil {
			return nil, facetErr
		}

		agg, facetErr := fragments.CreateAggregationBySearchLabel("resources.entries", ff, req.FacetAndBoolType, fub)
		if facetErr != nil {
			return nil, facetErr
		}

		s = s.Aggregation(facetField, agg)
	}

	postFilter, err := fub.CreateFacetFilterQuery("", req.FacetAndBoolType)
	if err != nil {
		rlog.Error().Err(err).
			Msg("unable to create search postfilter")

		return nil, err
	}

	s = s.PostFilter(postFilter)

	// spec count aggregation
	specCountAgg := elastic.NewCardinalityAggregation().
		Field(specField)

	eadTypeCountAgg := elastic.NewTermsAggregation().
		Field("meta.tags")

	countFilterAgg := elastic.NewFilterAggregation().
		Filter(postFilter).
		SubAggregation("specCount", specCountAgg).
		SubAggregation("typeCount", eadTypeCountAgg)

	queryStart := time.Now()

	resp, err := s.
		Query(req.Query).
		Aggregation("counts", countFilterAgg).
		Aggregation("noFiltTypeCount", eadTypeCountAgg).
		Do(r.Context())

	rlog.Info().
		Int("status", resp.Status).
		Int64("esTimeInMillis", resp.TookInMillis).
		Dur("duration", time.Since(queryStart)).
		Msg("elastic ead cluster search request")

	if err != nil {
		rlog.Error().Err(err).
			Msg("error in elasticsearch response")

		return nil, err
	}

	eadResponse := &SearchResponse{
		Archives: []Archive{},
	}

	if r.URL.Query().Get("explain") == trueParamValue {
		eadResponse.Explain = resp
	}

	if r.URL.Query().Get("service") == trueParamValue {
		ss := reflect.ValueOf(s).Elem().FieldByName("searchSource")
		src := reflect.NewAt(ss.Type(), unsafe.Pointer(ss.UnsafeAddr())).Elem().Interface().(*elastic.SearchSource)

		srcMap, sourceErr := src.Source()
		if sourceErr != nil {
			rlog.Error().Err(sourceErr).
				Msg("unable to decode elastich search request")

			return nil, sourceErr
		}

		eadResponse.Service = srcMap
	}

	unFilteredEadTypeCount, ok := resp.Aggregations.Terms("noFiltTypeCount")
	if ok {
		for _, b := range unFilteredEadTypeCount.Buckets {
			if b.Key == "eadDesc" {
				eadResponse.TotalDescriptionCount = int(b.DocCount)
			}
		}
	}

	filteredAgg, ok := resp.Aggregations.Filter("counts")
	if ok {
		specCount, ok := filteredAgg.Aggregations.Cardinality("specCount")
		if ok {
			eadResponse.ArchiveCount = int(*specCount.Value)
			eadResponse.TotalPages = getPageCount(eadResponse.ArchiveCount, req.Rows)
		}

		eadTypeCount, ok := filteredAgg.Aggregations.Terms("typeCount")
		if ok {
			for _, b := range eadTypeCount.Buckets {
				if b.Key == "ead" {
					eadResponse.TotalClevelCount = int(b.DocCount)
				}
			}
		}
	}

	eadResponse.TotalHits = eadResponse.TotalClevelCount + eadResponse.TotalDescriptionCount

	cursor := getCursor(req.Rows, req.Page)
	if cursor > eadResponse.ArchiveCount {
		pageErr := fmt.Errorf(
			"page start %d requested is greater then records returned: %d",
			cursor,
			eadResponse.ArchiveCount,
		)
		rlog.Error().Err(pageErr).
			Msg("request error")

		return nil, pageErr
	}

	eadResponse.Cursor = cursor

	for _, hit := range resp.Hits.Hits {
		fields, ok := hit.Fields[specField]
		if ok {
			spec := fields.([]interface{})[0].(string)

			ds, datasetErr := models.GetDataSet(spec)
			if datasetErr != nil {
				rlog.Error().Err(datasetErr).
					Msg("unable to retrieve dataset")

				return nil, datasetErr
			}

			archive := Archive{
				InventoryID:      spec,
				Title:            ds.Label,
				Period:           ds.Period,
				DescriptionCount: 0,
				ClevelsTotal:     ds.Clevels,
				// Files:            ds.Files,
				// Length:           ds.Length,
				// Abstract:         ds.Abstract,
				// Language:         ds.Language,
				// Material:         ds.Material,
				// Origin:           ds.ArchiveCreator,
				// MetsFiles:        ds.MetsFiles,
			}

			inner, ok := hit.InnerHits["collapse"]
			if ok {
				archive.CLevelCount = int(inner.Hits.TotalHits.Value)

				if len(inner.Hits.Hits) > 0 {
					r := new(fragments.FragmentGraph)

					if unmarshallErr := json.Unmarshal(inner.Hits.Hits[0].Source, r); unmarshallErr != nil {
						rlog.Error().Err(unmarshallErr).
							Msg("unable to unmarshal json for elasticsearch hit")

						return nil, unmarshallErr
					}

					if r.Tree.InventoryID != "" {
						archive.CLevelCount--
						archive.DescriptionCount = 1
					}
				}
			}

			if req.RawQuery != "" && ds.Description != "" && req.enableDescriptionSearch() {
				descriptionIndex := memory.NewTextIndex()

				appendErr := descriptionIndex.AppendString(ds.Description)
				if appendErr != nil {
					rlog.Error().Err(appendErr).
						Str("subquery", "description").
						Msg("error with appending description to TextIndex")

					return nil, appendErr
				}

				queryParser, parseErr := search.NewQueryParser()
				if parseErr != nil {
					rlog.Error().Err(parseErr).
						Str("subquery", "description").
						Msg("unable to create search.QueryParser.")

					return nil, parseErr
				}

				query, queryErr := queryParser.Parse(req.RawQuery)
				if queryErr != nil {
					rlog.Error().Err(queryErr).
						Str("subquery", "description").
						Msg("unable to parse query into search.QueryTerm")

					return nil, queryErr
				}

				hits, searchErr := descriptionIndex.Search(query)
				if searchErr != nil && !errors.Is(searchErr, memory.ErrSearchNoMatch) {
					rlog.Error().Err(searchErr).
						Str("subquery", "description").
						Msg("unable to search description")

					return nil, searchErr
				}

				archive.DescriptionCount = hits.Total()
			}

			eadResponse.Archives = append(eadResponse.Archives, archive)
		}
	}

	if !req.enableDescriptionSearch() {
		eadResponse.TotalDescriptionCount = 0
	}

	// build facets
	aggs, err := fragments.DecodeFacets(resp, fub)
	if err != nil {
		rlog.Error().Err(err).
			Msg("facet decode error")

		return nil, err
	}

	eadResponse.Facets = aggs

	if httpCache != nil && requestKey != "" && !req.NoCache {
		// don't cache no results
		if eadResponse.TotalHits == 0 {
			return eadResponse, nil
		}

		storeResponseInCache(requestKey, eadResponse, &rlog)
	}

	return eadResponse, nil
}

func storeResponseInCache(requestKey string, response *SearchResponse, rlog *zerolog.Logger) {
	b, err := json.Marshal(response)
	if err != nil {
		rlog.Error().Err(err).
			Msg("unable to marshal eadResponse for caching")
	} else {
		cacheErr := httpCache.Set(requestKey, b)
		if cacheErr != nil {
			rlog.Error().Err(cacheErr).
				Msg("unable to cache searchResponse")
		}
		rlog.Debug().
			Str("cache_key", requestKey).
			Msg("set cache for key")
	}
}

func PerformDetailSearch(r *http.Request) (*SearchResponse, error) {
	once.Do(newBigCache)

	inventoryID := chi.URLParam(r, "inventoryID")

	requestID, _ := hlog.IDFromRequest(r)
	rlog := c.Config.Logger.With().
		Str("req_id", requestID.String()).
		Str("searchType", "ead detail search").
		Str("inventoryID", inventoryID).
		Logger()

	req, err := buildSearchRequest(r, false)
	if err != nil {
		rlog.Error().Err(err).
			Msg("EAD detail error")

		return nil, err
	}

	req.InventoryID = inventoryID

	if req.CacheReset {
		newBigCache()
		// already cache this request
		req.CacheReset = false
	}

	requestKey := req.requestKey()
	rlog.Debug().
		Str("cache_key", requestKey).
		Msg("generating cache request key")

	if httpCache != nil && requestKey != "" && !req.NoCache && !req.CacheRefresh {
		response := getCachedRequest(requestKey, &rlog)
		if response != nil {
			return response, nil
		}
	}

	if req.CacheRefresh {
		req.CacheRefresh = false
		requestKey = req.requestKey()
	}

	s := req.Service
	query := req.Query

	query = query.Must(elastic.NewTermQuery(specField, inventoryID))

	// only return the tree part of the search response
	fsc := elastic.NewFetchSourceContext(true)
	fsc.Include("tree")
	s = s.FetchSourceContext(fsc)

	postFilter := elastic.NewBoolQuery()

	for _, qf := range req.Filters {
		switch {
		case strings.HasPrefix(qf.SearchLabel, "tree."):
			postFilter = postFilter.Must(elastic.NewTermQuery(qf.SearchLabel, qf.Value))
		default:
			f, filterErr := qf.ElasticFilter()
			if filterErr != nil {
				return nil, filterErr
			}

			if qf.Exclude {
				postFilter = postFilter.MustNot(f)
				continue
			}

			postFilter = postFilter.Must(f)
		}
	}

	req.Query = req.Query.Must(postFilter)

	resp, err := s.
		Query(query).
		Sort("tree.sortKey", true).
		Do(r.Context())

	if err != nil {
		rlog.Error().Err(err).
			Msg("error in elasticsearch response")
		return nil, err
	}

	eadResponse := &SearchResponse{
		Archives:         []Archive{},
		TotalClevelCount: int(resp.TotalHits()),
	}

	if r.URL.Query().Get("explain") == trueParamValue {
		eadResponse.Explain = resp
		ss := reflect.ValueOf(s).Elem().FieldByName("searchSource")
		src := reflect.NewAt(ss.Type(), unsafe.Pointer(ss.UnsafeAddr())).Elem().Interface().(*elastic.SearchSource)

		srcMap, err := src.Source()
		if err != nil {
			rlog.Error().Err(err).
				Msg("unable to decode elastich search request")

			return nil, err
		}

		eadResponse.Service = srcMap
	}

	eadResponse.TotalHits = eadResponse.TotalClevelCount + eadResponse.TotalDescriptionCount

	if eadResponse.TotalHits > 0 {
		eadResponse.ArchiveCount = 1
		eadResponse.TotalPages = getPageCount(eadResponse.TotalClevelCount, req.Rows)
	}

	cursor := getCursor(req.Rows, req.Page)
	if cursor > eadResponse.TotalClevelCount {
		err := fmt.Errorf(
			"page start %d requested is greater then records returned: %d",
			cursor,
			eadResponse.ArchiveCount,
		)
		rlog.Error().Err(err).
			Msg("request error")

		return nil, err
	}

	eadResponse.Cursor = cursor

	if resp == nil || resp.TotalHits() == 0 {
		return eadResponse, nil
	}

	eadResponse.CLevels = []CLevelEntry{}

	for _, hit := range resp.Hits.Hits {
		r := new(fragments.FragmentGraph)
		if err := json.Unmarshal(hit.Source, r); err != nil {
			return nil, err
		}

		tree := r.Tree

		cLevel := CLevelEntry{
			UnitID:      tree.UnitID,
			Label:       tree.Label,
			HubID:       tree.HubID,
			ResultOrder: tree.SortKey,
			Path:        tree.CLevel,
		}

		eadResponse.CLevels = append(eadResponse.CLevels, cLevel)
	}

	if httpCache != nil && requestKey != "" && !req.NoCache {
		// don't cache no results
		if eadResponse.TotalHits == 0 {
			return eadResponse, nil
		}

		storeResponseInCache(requestKey, eadResponse, &rlog)
	}

	return eadResponse, nil
}

func getCursor(rows, page int) int {
	if page == 1 || rows == 0 {
		return 0
	}

	start := ((page - 1) * rows)
	if start < 1 {
		return 0
	}

	return start
}

func getPageCount(archives, rows int) int {
	if rows == 0 || archives == 0 {
		return 0
	}

	if archives < rows {
		return 1
	}

	pages := archives / rows
	if archives%rows != 0 {
		pages++
	}

	return pages
}

// isAdvancedSearch checks if the query contains Lucene QueryString
// advanced search query syntax.
func isAdvancedSearch(query string) bool {
	parts := strings.Fields(query)
	for _, p := range parts {
		switch {
		case p == "AND":
			return true
		case p == "OR":
			return true
		case p == "NOT":
			return true
		case strings.HasPrefix(p, "-"):
			return true
		case strings.HasPrefix(p, "+"):
			return true
		case strings.HasPrefix(p, "\""):
			return true
		case strings.HasSuffix(p, "\""):
			return true
		}
	}

	return false
}