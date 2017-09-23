// Copyright © 2017 Delving B.V. <info@delving.eu>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hub3

import (
	"bytes"
	"fmt"
	"net/url"
	"time"

	. "bitbucket.org/delving/rapid/config"
	"github.com/knakk/rdf"
	"github.com/knakk/sparql"
	"github.com/sirupsen/logrus"
)

const queries = `
# SPARQL queries that are loaded as a QueryBank

# ask returns a boolean
# tag: ask_subject
ASK { <{{ .Uri }}> ?p ?o }

# tag: ask_predicate
ASK { ?s <{{ .Uri }}> ?o }

# tag: ask_object
ASK { ?s <{{ .Uri }}> ?o }

# tag: ask_query
ASK { {{ .Query }} }

# The DESCRIBE form returns a single result RDF graph containing RDF data about resources.
# tag: describe
DESCRIBE <{{ .Uri }}>
`

var queryBank sparql.Bank

// SparqlQueryURL is the fully qualified URI to the SPARQL endpoint
var SparqlQueryURL string

// SparqlRepo is the repository used for querying
var SparqlRepo *sparql.Repo

func init() {
	SparqlQueryURL = getSparqlEndpoint("")
	f := bytes.NewBufferString(queries)
	queryBank = sparql.LoadBank(f)
	SparqlRepo = buildRepo(SparqlQueryURL)
}

// buildRepo builds the query repository
func buildRepo(endPoint string) *sparql.Repo {
	if endPoint == "" {
		endPoint = getSparqlEndpoint("")
	}
	repo, err := sparql.NewRepo(endPoint,
		sparql.Timeout(time.Millisecond*1500),
	)
	if err != nil {
		logger.Fatal(err)
	}
	return repo
}

// getSparqlEndpoint builds the SPARQL endpoint from the Config object.
// When the dbName is empty the OrgId from the configuration is used.
func getSparqlEndpoint(dbName string) string {
	if dbName == "" {
		dbName = Config.OrgID
	}
	u, err := url.Parse(Config.RDF.SparqlHost)

	if err != nil {
		logger.Fatal(err)
	}
	u.Path = fmt.Sprintf(Config.RDF.SparqlPath, dbName)
	//logger.WithField("endpoint", u.String()).Debug("Building SPARQL endpoint")
	return u.String()
}

// PrepareAsk takes an a string and returns a valid SPARQL ASK query
func PrepareAsk(uri string) (string, error) {
	q, err := queryBank.Prepare("ask_subject", struct{ Uri string }{uri})

	if err != nil {
		logger.WithFields(logrus.Fields{"err": err, "uri": uri}).Error("Unable to build ask query")
		return "", err
	}
	return q, err
}

// AskSPARQL performs a SPARQL ASK query
func AskSPARQL(query string) bool {
	res, err := SparqlRepo.Query(query)
	if err != nil {
		logger.WithField("sparql", "ask").Fatal(err)
	}
	bindings := res.Bindings()
	logger.Debug(bindings)
	return false
}

func DescribeSPARQL(uri string) map[string][]rdf.Term {
	query, err := queryBank.Prepare("describe", struct{ Uri string }{uri})
	if err != nil {
		logger.WithField("uri", uri).Errorf("Unable to build describe query.")
	}
	res, err := SparqlRepo.Query(query)
	if err != nil {
		logger.WithField("query", query).Errorf("Unable query endpoint")
	}
	return res.Bindings()
}
