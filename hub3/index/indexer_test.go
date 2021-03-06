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

package index

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Indexer", func() {

	Describe("when initialised", func() {

		It("should have a search client", func() {
			c := ESClient()
			Expect(c).ToNot(BeNil())
			Expect(client).ToNot(BeNil())
		})

		It("should have a bulk-indexer", func() {
			bps := IndexingProcessor()
			Expect(bps).ToNot(BeNil())
			Expect(processor).ToNot(BeNil())
		})

	})
})

//var testTurtleRecord = string(ioutil.ReadFile("test_data/test1.ttl"))
