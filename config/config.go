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

package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	// Config is the general configuration object
	Config RawConfig

	// CfgFile is the path to the config file
	CfgFile string
)

// RawConfig holds all the configuration blocks.
// These are bound from cli, Environment variables or configuration files by
// Viper.
type RawConfig struct {
	OrgID         string `json:"orgId"`
	HTTP          `json:"http"`
	ElasticSearch `json:"elasticsearch"`
	Logging       `json:"logging"`
	RDF           `json:"rdf"`
	OAIPMH        `json:"oaipmh"`
	WebResource   `json:"webresource"`
}

// ElasticSearch holds all the configuration values
// It is bound by Viper.
type ElasticSearch struct {
	Urls []string `json:"urls"`
}

// Logging holds all the logging and path configuration
type Logging struct {
	DevMode   bool   `json:"devmode"`
	SentryDSN string `json:"sentrydsn"`
}

// HTTP holds all the configuration for the http server subcommand
type HTTP struct {
	Port      int    `json:"port" mapstructure:"port"`
	StaticDir string `json:"staticDir"` // the relative path to the static directory to serve documentation.
}

// RDF holds all the configuration for SPARQL queries and RDF conversions
type RDF struct {
	SparqlHost        string   `json:"sparqlHost"`        // the base-url to the SPARQL endpoint including the scheme and the port
	SparqlPath        string   `json:"sparqlPath"`        // the relative path of the endpoint. This can should contain the database name that is injected when the sparql endpoint is build
	BaseUrl           string   `json:"baseUrl"`           // the RDF baseUrl used for minting new URIs
	RoutedEntryPoints []string `json:"RoutedEntryPoints"` // the RDF baseUrl used for minting new URIs
	// the RDF entryPoints. Lookups are made on the fully qualified URIs. It is sometimes needed to support other baseUrls as well.
	// The entry-points need to be fully qualified, i.e. with their scheme.
}

// OAIPMH holds all the configuration options for the OAI-PMH endpoint
type OAIPMH struct {
	// Make the oai-pmh endpoint available
	Enabled bool `json:"enabled"`
	// AdminEmails has a list of the admin emails of this endpoint
	AdminEmails []string `json:"adminEmails"`
	// RepositoryName is the name of the OAI-PMH repossitory
	RepositoryName string `json:"repositoryName"`
}

// WebResource holds all the configuration options for the WebResource endpoint
type WebResource struct {
	Enabled        bool   `json:"enabled"`   // Make the webresource endpoint available
	WebResourceDir string `json:"sourceDir"` // Target directory for the webresources
}

func setDefaults() {

	// setting defaults
	viper.SetDefault("HTTP.port", 3001)
	viper.SetDefault("HTTP.staticDir", "public")
	viper.SetDefault("orgId", "rapid")

	// elastic
	viper.SetDefault("ElasticSearch.urls", []string{"http://localhost:9200"})

	// logging
	viper.SetDefault("Logging.DevMode", false)

	// rdf with defaults for Blazegraph
	viper.SetDefault("RDF.SparqlHost", "http://localhost:9999")
	viper.SetDefault("RDF.SparqlPath", "/bigdata/namespace/%s/sparql")
	viper.SetDefault("RDF.BaseUrl", "http://data.rapid.org")
	viper.SetDefault("RDF.RoutedEntryPoints", []string{"http://localhost:3000", "http://localhost:3001"})

	// oai-pmh
	viper.SetDefault("OAIPMH.enabled", true)
	viper.SetDefault("OAIPMH.repostitoryName", "rapid")
	viper.SetDefault("OAIPMH.AdminEmails", "info@delving.eu")
}

func cleanConfig() {
	Config.RDF.BaseUrl = strings.TrimSuffix(Config.RDF.BaseUrl, "/")
	if !strings.HasPrefix(Config.RDF.BaseUrl, "http://") {
		log.Fatalf("RDF.BaseUrl config value '%s' should start with 'http://'", Config.RDF.BaseUrl)
	}
}

// InitConfig reads in config file and ENV variables if set.
func InitConfig() {
	if CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(CfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".rapid" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".rapid")
	}

	viper.SetEnvPrefix("RAPID")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	setDefaults()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}
	err := viper.Unmarshal(&Config)
	if err != nil {
		log.Fatal(
			fmt.Sprintf("unable to decode into struct, %v", err),
		)
	}

	cleanConfig()
}

// GetSparqlEndpoint builds the SPARQL endpoint from the RDF Config object.
// When the dbName is empty the OrgId from the configuration is used.
func (c RawConfig) GetSparqlEndpoint(dbName string) string {
	if dbName == "" {
		dbName = c.OrgID
	}
	u, err := url.Parse(c.RDF.SparqlHost)

	if err != nil {
		log.Fatal(err)
	}
	u.Path = fmt.Sprintf(c.RDF.SparqlPath, dbName)
	log.Printf("Sparql endpoint: %s", u)
	return u.String()
}
