package models

import (
	"time"

	"bitbucket.org/delving/rapid/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dataset", func() {

	Context("When creating a dataset URI", func() {

		config.InitConfig()
		uri := createDatasetURI("test")
		It("should end with the spec", func() {
			Expect(uri).To(HaveSuffix("test"))
		})

		It("should contain the resource path and type", func() {
			Expect(uri).To(ContainSubstring("/resource/dataset/"))
		})

		It("should start with the RDF baseUrl from the configuration.", func() {
			baseUrl := config.Config.RDF.BaseUrl
			Expect(uri).To(HavePrefix(baseUrl))
		})

	})

	Context("When creating a new Dataset", func() {
		spec := "test_spec"
		dataset := NewDataset(spec)
		It("should set the spec", func() {
			Expect(dataset).ToNot(BeNil())
			Expect(dataset.Spec).To(Equal(spec))
		})

		It("should set a datasetUri", func() {
			uri := dataset.URI
			Expect(uri).ToNot(BeEmpty())
			Expect(uri).To(Equal(createDatasetURI(spec)))
		})

		It("should set the creation time", func() {
			created := dataset.Created
			Expect(created).ToNot(BeNil())
			Expect(created.Day()).To(Equal(time.Now().Day()))
			Expect(created.Month()).To(Equal(time.Now().Month()))
			Expect(created.Year()).To(Equal(time.Now().Year()))
		})

		It("the creationd and modification time should be the same", func() {
			Expect(dataset.Created).To(Equal(dataset.Modified))
		})

		It("should set the revision to zero", func() {
			Expect(dataset.Revision).To(Equal(0))
		})

		It("should set deleted to be false", func() {
			Expect(dataset.Deleted).To(BeFalse())
		})

		It("should have access set to true", func() {
			Expect(dataset.Access.OAIPMH).To(BeTrue())
			Expect(dataset.Access.Search).To(BeTrue())
			Expect(dataset.Access.LOD).To(BeTrue())
		})

	})

	Context("When saving a DataSet", func() {
		spec := "test_spec"
		dataset := NewDataset(spec)

		It("should have nothing saved before save", func() {
			var ds []DataSet
			err := orm.All(&ds)
			Expect(err).To(BeNil())
			Expect(len(ds)).To(Equal(0))
		})

		It("should save a dataset without errors", func() {
			Expect(dataset.Save()).To(BeNil())
			var ds []DataSet
			err := orm.All(&ds)
			Expect(err).To(BeNil())
			Expect(len(ds)).To(Equal(1))
		})

		It("should be able to find it in the database", func() {
			var ds DataSet
			err := orm.One("Spec", spec, &ds)
			Expect(err).To(BeNil())
			Expect(ds.Created.Unix()).To(Equal(dataset.Created.Unix()))
			Expect(ds.Modified.UnixNano()).ToNot(Equal(dataset.Modified.UnixNano()))
			Expect(ds.Access.LOD).To(BeTrue())
		})

	})

	Context("When calling GetOrCreateDataSet", func() {

		It("should create the datasets when no dataset is available", func() {
			ds, err := GetOrCreateDataSet("test2")
			Expect(err).ToNot(HaveOccurred())
			Expect(ds.Spec).To(Equal("test2"))
		})

		It("should not store the dataset again on Get", func() {
			datasetCount, err := ListDataSets()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(datasetCount) > 0).To(BeTrue())
			ds, err := GetOrCreateDataSet("test2")
			Expect(err).ToNot(HaveOccurred())
			Expect(ds.Spec).To(Equal("test2"))
			newCount, err := ListDataSets()
			Expect(len(datasetCount)).To(Equal(len(newCount)))
		})
	})

	Context("When calling ListDatasets", func() {

		It("should return an array of all stored datasets", func() {
			datasets, err := ListDataSets()
			Expect(err).ToNot(HaveOccurred())
			Expect(datasets).ToNot(BeEmpty())
		})

	})

})