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

package fragments

import (
	"bytes"
	"encoding/json"
	fmt "fmt"
	"io/ioutil"
	"log"
	"net/url"

	c "github.com/delving/rapid-saas/config"
	r "github.com/kiivihal/rdf2go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// testGraph creates a dummy graph for testing
func testGraph() *r.Graph {
	baseUri := "http://rapid.org/resource"

	g := r.NewGraph(baseUri)
	g.Add(r.NewTriple(r.NewResource("a"), r.NewResource("b"), r.NewResource("c")))
	//r.NewTriple(r.NewResource("a"), r.NewResource("title"), r.NewLiteral("title")),
	//r.NewTriple(r.NewResource("a"), r.NewResource("subject"), r.NewLiteralWithLanguage("subject", "nl")),
	return g
}

// URIRef is a function to create an RDFLiteal resource
func URIRef(uri string) r.Term {
	return r.NewResource(uri)
}

// NSRef creates a URIRef with the RDF Base URL
func NSRef(uri string) r.Term {
	return r.NewResource(fmt.Sprintf("%s/%s", c.Config.RDF.BaseURL, uri))
}

// Literal is a utility function to create a RDF literal
func Literal(value string, language string, dataType ObjectXSDType) r.Term {
	if language != "" {
		return r.NewLiteralWithLanguage(value, language)
	}
	if dataType != ObjectXSDType_STRING {
		t, err := dataType.GetLabel()
		if err != nil {
			log.Println("Unable to get label for this type")
		}
		return r.NewLiteralWithDatatype(value, r.NewResource(t))
	}
	return r.NewLiteral(value)
}

func testFragmentGraph(spec string, rev int32, ng string) *FragmentGraph {
	fg := NewFragmentGraph()
	fg.OrgID = "rapid"
	fg.Spec = spec
	fg.Revision = rev
	fg.NamedGraphURI = ng
	return fg
}

func testDataGraph(empty bool) (*FragmentBuilder, error) {
	spec := "test-spec"
	rev := int32(1)
	ng := "http://data.jck.nl/resource/aggregation/jhm-foto/F900893/graph"
	fg := testFragmentGraph(spec, rev, ng)
	fg.EntryURI = "http://www.openarchives.org/ore/terms/Aggregation"
	fb := NewFragmentBuilder(fg)
	dat, err := ioutil.ReadFile("test_data/test2.ttl")
	if err != nil {
		return fb, err
	}
	fg.RDF = dat
	if !empty {
		fb.ParseGraph(bytes.NewReader(fg.RDF), "text/turtle")
	}
	return fb, nil
}

var _ = Describe("Fragments", func() {

	Describe("When receiving a Triple", func() {

		spec := "test-spec"
		rev := int32(1)
		ng := "urn:1/graph"
		fg := testFragmentGraph(spec, rev, ng)
		fb := NewFragmentBuilder(fg)

		Context("with an object resource", func() {

			t := r.NewTriple(URIRef("urn:1"), URIRef("urn:subject"), URIRef("urn:target"))
			f, err := fb.CreateFragment(t)

			It("should have a spec", func() {
				Expect(t).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
				Expect(f).ToNot(BeNil())
				Expect(f.GetSpec()).To(Equal(spec))

				Expect(f.GetPredicate()).To(Equal("urn:subject"))
			})

			It("should have a revision number", func() {
				Expect(f.GetRevision()).To(Equal(rev))
			})

			It("should have a NamedGraphURI", func() {
				Expect(f.GetNamedGraphURI()).To(Equal(ng))
			})

			It("should have an n-triple", func() {
				t := f.GetTriple()
				Expect(t).ToNot(BeEmpty())
			})

			It("should have a quad with the NamedGraphURI", func() {
				q := f.Quad()
				Expect(q).ToNot(BeEmpty())
				Expect(q).To(HaveSuffix("<urn:target> <%s> .", f.GetNamedGraphURI()))
			})

			It("should have an id that is a hashed version of the Quad", func() {
				id := f.ID()
				Expect(id).ToNot(BeEmpty())
				hash := CreateHash(f.Quad())
				Expect(id).To(Equal(hash))
			})

			It("should have a subject without <>", func() {
				r := f.GetSubject()
				Expect(r).To(Equal("urn:1"))
				Expect(r).ToNot(HaveSuffix("%s", ">"))
				Expect(r).ToNot(HavePrefix("%s", "<"))
			})

			It("should have predicate without <>", func() {
				r := f.GetPredicate()
				Expect(r).To(Equal("urn:subject"))
				Expect(r).ToNot(HaveSuffix("%s", ">"))
				Expect(r).ToNot(HavePrefix("%s", "<"))
			})

			It("should have an object", func() {
				r := f.GetObject()
				Expect(r).To(Equal("urn:target"))
				Expect(r).ToNot(HaveSuffix("%s", ">"))
				Expect(r).ToNot(HavePrefix("%s", "<"))
			})

			It("should have resource as objecttype", func() {
				t := f.GetObjectType()
				Expect(t).ToNot(BeNil())
				Expect(t).To(Equal(ObjectType_RESOURCE))
			})

		})

		Context("When receiving a object resource link", func() {

			g := r.NewGraph(c.Config.RDF.BaseURL)
			t1 := r.NewTriple(NSRef("1"), NSRef("subject"), NSRef("2"))
			t2 := r.NewTriple(NSRef("1"), NSRef("subject"), NSRef("3"))
			t3 := r.NewTriple(
				NSRef("1"),
				NSRef("subject"),
				r.NewResource("https://data.cultureelerfgoed.nl/term/id/cht/99efdcca-cce0-4629-adfb-becab8381183"),
			)
			t4 := r.NewTriple(
				NSRef("1"),
				r.NewResource("http://www.w3.org/1999/02/22-rdf-syntax-ns#type"),
				r.NewResource("http://www.europeana.eu/schemas/edm/WebResource"),
			)
			g.Add(t1)
			g.Add(t2)
			g.Add(t3)
			g.Add(t4)
			g.AddTriple(NSRef("2"), NSRef("prefLabel"), r.NewLiteral("subject of 2"))
			g.AddTriple(NSRef("1"), NSRef("title"), r.NewLiteral("2"))
			fb.Graph = g

			It("should mark the fragment not graphExternal if subject present in graph", func() {
				frag, err := fb.CreateFragment(t1)
				Expect(t1).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
				Expect(frag).ToNot(BeNil())
				external := fb.IsGraphExternal(t1.Object)
				Expect(external).To(BeFalse())
				// todo reimplement with tags
				//Expect(frag.GraphExternalLink).To(BeFalse())

			})

			It("should mark the fragment not graphExternal if subject present in graph", func() {
				frag, err := fb.CreateFragment(t2)
				Expect(t2).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
				Expect(frag).ToNot(BeNil())
				external := fb.IsGraphExternal(t2.Object)
				Expect(external).To(BeTrue())
				// todo reimplement with tags
				// Expect(frag.GraphExternalLink).To(BeTrue())
			})

			It("should mark the fragment as domainExternal when the host differs from the RDF base url", func() {
				frag, err := fb.CreateFragment(t3)
				Expect(t3).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
				Expect(frag).ToNot(BeNil())
				external, err := fb.IsDomainExternal(frag.Object)
				Expect(err).ToNot(HaveOccurred())
				Expect(external).To(BeTrue())
				// todo reimplement with tags
				//Expect(frag.DomainExternalLink).To(BeTrue())
			})

			It("should mark the fragment as not domainExternal when the host equals the RDF base url", func() {
				frag, err := fb.CreateFragment(t2)
				Expect(t2).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
				Expect(frag).ToNot(BeNil())
				external, err := fb.IsDomainExternal(frag.Object)
				Expect(err).ToNot(HaveOccurred())
				Expect(external).To(BeFalse())
				// todo reimplement with tags
				//Expect(frag.DomainExternalLink).To(BeFalse())
			})

			It("should not make type links as external", func() {
				frag, err := fb.CreateFragment(t4)
				Expect(t4).ToNot(BeNil())
				Expect(err).ToNot(HaveOccurred())
				Expect(frag).ToNot(BeNil())
				// todo reimplement with tags
				//Expect(frag.IsTypeLink()).To(BeTrue())
				//Expect(frag.GetTypeLink()).To(BeTrue())
			})

		})

		Context("when getting the ObjectXSDType", func() {

			It("should return the XSD label", func() {

			})

			It("should trim <>", func() {
				t, err := GetObjectXSDType("<http://www.w3.org/2001/XMLSchema#date>")
				Expect(err).ToNot(HaveOccurred())
				Expect(t).ToNot(BeNil())
			})
		})

		Context("when receiving a triple with a literal object", func() {

			t := r.NewTriple(URIRef("urn:1"), URIRef("urn:subject"), Literal("river", "", ObjectXSDType_STRING))
			f, err := fb.CreateFragment(t)

			It("should have literal as objecttype", func() {
				Expect(err).ToNot(HaveOccurred())
				t := f.GetObjectType()
				Expect(t).ToNot(BeNil())
				Expect(t).To(Equal(ObjectType_LITERAL))
			})

			It("should have no language", func() {
				Expect(f.Language).To(Equal(""))
			})

			It("should have string as datatype", func() {
				Expect(f.DataType).To(Equal(ObjectXSDType_STRING))
			})

			It("should have http://www.w3.org/2001/XMLSchema#string as default xsdRaw", func() {
				Expect(f.GetXSDRaw()).To(Equal("xsd:string"))
			})
		})

		Context("when receiving a triple with a literal and language", func() {
			t := r.NewTriple(URIRef("urn:1"), URIRef("urn:subject"), Literal("river", "en", ObjectXSDType_STRING))
			f, err := fb.CreateFragment(t)

			It("should have a language", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(f.Language).To(Equal("en"))
			})

			It("should have string as datatype", func() {
				Expect(f.DataType).To(Equal(ObjectXSDType_STRING))
			})

			It("should have xsd:string as default xsdRaw", func() {
				Expect(f.GetXSDRaw()).To(Equal("xsd:string"))
			})
		})

		Context("when receiving a triple with literal and type", func() {

			It("should have the custom dataType", func() {
				t := r.NewTriple(URIRef("urn:1"), URIRef("urn:subject"), Literal("river", "", ObjectXSDType_DATE))
				f, err := fb.CreateFragment(t)
				Expect(err).ToNot(HaveOccurred())
				Expect(f.GetDataType()).To(Equal(ObjectXSDType_DATE))
				Expect(f.GetXSDRaw()).To(Equal("xsd:date"))
			})
		})
	})

	Describe("creating a new FragmentRequest", func() {

		Context("directly", func() {

			It("should have no triple pattern set", func() {
				fr := NewFragmentRequest()
				Expect(fr).ToNot(BeNil())
				Expect(fr.GetSubject()).To(BeEmpty())
				Expect(fr.GetPredicate()).To(BeEmpty())
				Expect(fr.GetObject()).To(BeEmpty())
				Expect(fr.GetLanguage()).To(BeEmpty())
			})

			It("should have a non-zero page start", func() {
				fr := NewFragmentRequest()
				Expect(fr.GetPage()).To(Equal(int32(1)))
			})
		})

		Context("parsing from url.Values", func() {

			It("should ignore empty values", func() {
				fr := NewFragmentRequest()
				v := url.Values{}
				v.Add("subject", "urn:1")
				v.Add("predicate", "")
				v.Add("object", "")
				v.Add("language", "")
				v.Add("page", "2")
				err := fr.ParseQueryString(v)
				Expect(err).ToNot(HaveOccurred())
				Expect(fr.GetSubject()).To(Equal("urn:1"))
				Expect(fr.GetPredicate()).To(BeEmpty())
				Expect(fr.GetObject()).To(BeEmpty())
				Expect(fr.GetLanguage()).To(BeEmpty())
			})

			It("should throw an error when the page is not an int", func() {
				fr := NewFragmentRequest()
				v := url.Values{}
				v.Add("page", "error")
				err := fr.ParseQueryString(v)
				Expect(err).To(HaveOccurred())
			})

			It("should set the page when it is an int", func() {
				fr := NewFragmentRequest()
				v := url.Values{}
				v.Add("page", "10")
				err := fr.ParseQueryString(v)
				Expect(err).ToNot(HaveOccurred())
				Expect(fr.GetPage()).To(Equal(int32(10)))
			})

			It("should set all the non-empty values", func() {
				fr := NewFragmentRequest()
				v := url.Values{}
				v.Add("subject", "urn:1")
				v.Add("predicate", "urn:subject")
				v.Add("object", "mountain")
				v.Add("language", "nl")
				v.Add("page", "3")
				err := fr.ParseQueryString(v)
				Expect(err).ToNot(HaveOccurred())
				Expect(fr.GetSubject()).To(Equal("urn:1"))
				Expect(fr.GetPredicate()).To(Equal("urn:subject"))
				Expect(fr.GetObject()).To(Equal("mountain"))
				Expect(fr.GetLanguage()).To(Equal("nl"))
				Expect(fr.GetPage()).To(Equal(int32(3)))
			})
		})
	})

	Describe("when creating a fragment", func() {

		fg := testFragmentGraph("test", int32(1), "urn:1/graph")
		fb := NewFragmentBuilder(fg)

		Context("and converting it to a BulkIndexRequest", func() {

			t := r.NewTriple(URIRef("urn:1"), URIRef("urn:subject"), Literal("river", "en", ObjectXSDType_STRING))
			f, err := fb.CreateFragment(t)

			It("the fragment should be valid", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(f).ToNot(BeNil())
			})

			It("should create the BulkIndexRequest", func() {
				bir, err := f.CreateBulkIndexRequest()
				Expect(err).ToNot(HaveOccurred())
				Expect(bir).ToNot(BeNil())
			})

			It("should have a valid header", func() {
				bir, err := f.CreateBulkIndexRequest()
				Expect(err).ToNot(HaveOccurred())
				lines, err := bir.Source()
				Expect(err).ToNot(HaveOccurred())
				header := lines[0]
				//body := lines[1]
				var h interface{}
				err = json.Unmarshal([]byte(header), &h)
				Expect(err).ToNot(HaveOccurred())
				m := h.(map[string]interface{})
				Expect(m["index"]).To(HaveKeyWithValue("_id", f.ID()))
				Expect(m["index"]).To(HaveKeyWithValue("_type", DocType))
				Expect(m["index"]).To(HaveKeyWithValue("_index", c.Config.ElasticSearch.IndexName))
			})

			It("should have a valid body", func() {
				bir, err := f.CreateBulkIndexRequest()
				Expect(err).ToNot(HaveOccurred())
				lines, err := bir.Source()
				Expect(err).ToNot(HaveOccurred())
				body := lines[1]
				Expect(body).ToNot(BeEmpty())
				var b interface{}
				err = json.Unmarshal([]byte(body), &b)
				Expect(err).ToNot(HaveOccurred())
				m := b.(map[string]interface{})
				Expect(m).To(HaveKeyWithValue("subject", "urn:1"))
				Expect(m).To(HaveKeyWithValue("XSDRaw", "xsd:string"))
				Expect(m).To(HaveKeyWithValue("language", "en"))
			})
		})
	})

	Describe("ObjectXSDType conversions", func() {

		Context("when converting to label", func() {

			It("should return the xsd label when found", func() {
				label, err := ObjectXSDType_BOOLEAN.GetLabel()
				Expect(err).ToNot(HaveOccurred())
				Expect(label).ToNot(BeEmpty())
				Expect(label).To(Equal("http://www.w3.org/2001/XMLSchema#boolean"))
			})

			It("should return an error when no label could be found", func() {
				const ObjectXSDType_ERROR ObjectXSDType = 100
				label, err := ObjectXSDType_ERROR.GetLabel()
				Expect(err).To(HaveOccurred())
				Expect(label).To(BeEmpty())
			})
		})

		Context("when requesting a prefix label", func() {

			It("should shorten the namespace to xsd", func() {
				label, err := ObjectXSDType_BOOLEAN.GetPrefixLabel()
				Expect(err).ToNot(HaveOccurred())
				Expect(label).ToNot(BeEmpty())
				Expect(label).To(Equal("xsd:boolean"))

			})
		})

		Context("when converting from a label", func() {

			It("should return the ObjectXSDType", func() {
				t, err := GetObjectXSDType("http://www.w3.org/2001/XMLSchema#boolean")
				Expect(err).ToNot(HaveOccurred())
				Expect(t).ToNot(BeNil())
				Expect(t).To(Equal(ObjectXSDType_BOOLEAN))
			})
		})

	})

	Describe("hasher", func() {

		Context("When given a string", func() {

			It("should return a short hash", func() {
				hash := CreateHash("rapid rocks.")
				Expect(hash).To(Equal("a5b3be36c0f378a1"))
			})
		})

	})

	Describe("FragmentRequest", func() {

		Context("when assiging an object", func() {

			It("should strip double quotes", func() {
				fr := NewFragmentRequest()
				fr.AssignObject(`"1982"`)
				Expect(fr.GetObject()).To(Equal("1982"))
			})

			It("should set the language when the string contains @ annotation", func() {
				fr := NewFragmentRequest()
				fr.AssignObject(`"door"@en`)
				Expect(fr.GetObject()).To(Equal("door"))
				Expect(fr.GetLanguage()).To(Equal("en"))
			})

		})
	})

	Describe("FragmentBuilder", func() {
		fb, err := testDataGraph(false)

		Context("when creating linked data", func() {

			It("should find subjects by EntryURI", func() {
				Expect(err).ToNot(HaveOccurred())
				err := fb.CreateLinkedFragments()
				Expect(err).ToNot(HaveOccurred())
				Expect(fb.Graph.Len()).To(Equal(65))
			})
		})

		Context("when a graphName is present", func() {
			spec := "test-spec"
			rev := int32(1)
			ng := "http://data.jck.nl/resource/aggregation/jhm-foto/F900893/graph"
			fg := testFragmentGraph(spec, rev, ng)

			It("should extract the about or source uri", func() {
				sourceURI := fg.GetAboutURI()
				Expect(sourceURI).ToNot(BeEmpty())
				Expect(sourceURI).ToNot(HaveSuffix("/graph"))
			})
		})

	})

})
