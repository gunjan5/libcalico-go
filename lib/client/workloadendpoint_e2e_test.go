// Copyright (c) 2016 Tigera, Inc. All rights reserved.

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

// Test cases (WorkloadEndpoint object e2e):
// Test 1: Pass two fully populated WorkloadEndpointSpecs and expect the series of operations to succeed.
// Test 2: Pass one partially populated WorkloadEndpointSpec and another fully populated WorkloadEndpointSpec and expect the series of operations to succeed.
// Test 3: Pass one fully populated WorkloadEndpointSpec and another empty WorkloadEndpointSpec and expect the series of operations to succeed.

// Series of operations each test goes through:
// Update meta1 - check for failure (because it doesn't exist).
// Create meta1 with spec1.
// Apply meta2 with spec2.
// Get meta1 and meta2, compare spec1 and spec2.
// Update meta1 with spec2.
// Get meta1 compare spec2.
// List (empty Meta) ... Get meta1 and meta2.
// List (using Meta1) ... Get meta1.
// Delete meta1.
// Get meta1 ... fail.
// Delete meta2.
// List (empty Meta) ... Get no entries (should not error).

package client_test

import (
	"errors"
	"log"
	"net"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/projectcalico/libcalico-go/lib/api"
	"github.com/projectcalico/libcalico-go/lib/testutils"

	cnet "github.com/projectcalico/libcalico-go/lib/net"
)

var _ = Describe("WorkloadEndpoint tests", func() {
	_, cidr1, _ := net.ParseCIDR("10.0.0.0/24")
	_, cidr2, _ := net.ParseCIDR("20.0.0.0/24")
	_, cidr3, _ := net.ParseCIDR("192.168.0.0/24")
	_, cidr4, _ := net.ParseCIDR("172.56.0.0/24")
	mac1, _ := net.ParseMAC("01:23:45:67:89:ab")
	mac2, _ := net.ParseMAC("01:23:4b:ad:be:ef")

	DescribeTable("WorkloadEndpoint e2e tests",
		func(meta1, meta2 api.WorkloadEndpointMetadata, spec1, spec2 api.WorkloadEndpointSpec) {

			// Erase etcd clean.
			testutils.CleanEtcd()

			// Create a new client.
			c, err := testutils.NewClient("")
			if err != nil {
				log.Println("Error creating client:", err)
			}
			By("Updating the WorkloadEndpoint before it is created")
			_, outError := c.WorkloadEndpoints().Update(&api.WorkloadEndpoint{Metadata: meta1, Spec: spec1})

			// Should return an error.
			Expect(outError.Error()).To(Equal(errors.New("resource does not exist: WorkloadEndpoint(hostname=node1, orchestrator=kubernetes, workload=workload1, name=host1)").Error()))

			By("Create, Apply, Get and compare")

			// Create a WorkloadEndpoint with meta1 and spec1.
			_, outError = c.WorkloadEndpoints().Create(&api.WorkloadEndpoint{Metadata: meta1, Spec: spec1})

			// Apply a WorkloadEndpoint with meta2 and spec2.
			_, outError = c.WorkloadEndpoints().Apply(&api.WorkloadEndpoint{Metadata: meta2, Spec: spec2})

			// Get WorkloadEndpoint with meta1.
			outWorkloadEndpoint1, outError1 := c.WorkloadEndpoints().Get(meta1)
			log.Println("Out WorkloadEndpoint object: ", outWorkloadEndpoint1)

			// Get WorkloadEndpoint with meta2.
			outWorkloadEndpoint2, outError2 := c.WorkloadEndpoints().Get(meta2)
			log.Println("Out WorkloadEndpoint object: ", outWorkloadEndpoint2)

			// Should match spec1 & outWorkloadEndpoint1 and outWorkloadEndpoint2 & spec2 and errors to be nil.
			Expect(outWorkloadEndpoint1.Spec).To(Equal(spec1))
			Expect(outWorkloadEndpoint2.Spec).To(Equal(spec2))
			Expect(outError1).NotTo(HaveOccurred())
			Expect(outError2).NotTo(HaveOccurred())

			By("Update, Get and compare")

			// Update meta1 WorkloadEndpoint with spec2.
			c.WorkloadEndpoints().Update(&api.WorkloadEndpoint{Metadata: meta1, Spec: spec2})

			// Get WorkloadEndpoint with meta1.
			outWorkloadEndpoint1, outError1 = c.WorkloadEndpoints().Get(meta1)

			// Assert the Spec for WorkloadEndpoint with meta1 matches spec2 and no error.
			Expect(outWorkloadEndpoint1.Spec).To(Equal(spec2))
			Expect(outError1).NotTo(HaveOccurred())

			By("List all the WorkloadEndpoints and compare")

			// Get a list of WorkloadEndpoints.
			WorkloadEndpointList, outError := c.WorkloadEndpoints().List(api.WorkloadEndpointMetadata{})
			log.Println("Get WorkloadEndpoint list returns: ", WorkloadEndpointList.Items)
			metas := []api.WorkloadEndpointMetadata{meta1, meta2}
			expectedWorkloadEndpoints := []api.WorkloadEndpoint{}
			// Go through meta list and append them to expectedWorkloadEndpoints.
			for _, v := range metas {
				p, _ := c.WorkloadEndpoints().Get(v)
				expectedWorkloadEndpoints = append(expectedWorkloadEndpoints, *p)
			}

			// Assert the returned WorkloadEndpointList is has the meta1 and meta2 WorkloadEndpoints.
			Expect(WorkloadEndpointList.Items).To(Equal(expectedWorkloadEndpoints))

			By("List a specific WorkloadEndpoint and compare")

			// Get a WorkloadEndpoint list with meta1.
			WorkloadEndpointList, outError = c.WorkloadEndpoints().List(meta1)
			log.Println("Get WorkloadEndpoint list returns: ", WorkloadEndpointList.Items)

			// Get a WorkloadEndpoint with meta1.
			outWorkloadEndpoint1, outError1 = c.WorkloadEndpoints().Get(meta1)

			// Assert they are equal and no errors.
			Expect(WorkloadEndpointList.Items[0].Spec).To(Equal(outWorkloadEndpoint1.Spec))
			Expect(outError1).NotTo(HaveOccurred())

			By("Delete, Get and assert error")

			// Delete a WorkloadEndpoint with meta1.
			outError1 = c.WorkloadEndpoints().Delete(meta1)

			// Get a WorkloadEndpoint with meta1.
			_, outError = c.WorkloadEndpoints().Get(meta1)

			// Expect an error since the WorkloadEndpoint was deleted.
			Expect(outError.Error()).To(Equal(errors.New("resource does not exist: WorkloadEndpoint(hostname=node1, orchestrator=kubernetes, workload=workload1, name=host1)").Error()))

			// Delete the second WorkloadEndpoint with meta2.
			outError1 = c.WorkloadEndpoints().Delete(meta2)

			By("Delete all the WorkloadEndpoints, Get WorkloadEndpoint list and expect empty WorkloadEndpoint list")

			// Both WorkloadEndpoints are deleted in the calls above.
			// Get the list of all the WorkloadEndpoints.
			WorkloadEndpointList, outError = c.WorkloadEndpoints().List(api.WorkloadEndpointMetadata{})
			log.Println("Get WorkloadEndpoint list returns: ", WorkloadEndpointList.Items)

			// Create an empty WorkloadEndpoint list.
			// Note: you can't use make([]api.WorkloadEndpoint, 0) because it creates an empty underlying struct,
			// whereas new([]api.WorkloadEndpoint) just returns a pointer without creating an empty struct.
			emptyWorkloadEndpointList := new([]api.WorkloadEndpoint)

			// Expect returned WorkloadEndpointList to contain empty WorkloadEndpointList.
			Expect(WorkloadEndpointList.Items).To(Equal(*emptyWorkloadEndpointList))

		},

		// Test 1: Pass two fully populated WorkloadEndpointSpecs and expect the series of operations to succeed.
		Entry("Two fully populated WorkloadEndpointSpecs",
			api.WorkloadEndpointMetadata{
				Name:         "host1",
				Workload:     "workload1",
				Orchestrator: "kubernetes",
				Node:         "node1",
				Labels: map[string]string{
					"app":  "app-abc",
					"prod": "no",
				}},
			api.WorkloadEndpointMetadata{
				Name:         "host2",
				Workload:     "workload2",
				Orchestrator: "mesos",
				Node:         "node2",
				Labels: map[string]string{
					"app":  "app-xyz",
					"prod": "yes",
				}},
			api.WorkloadEndpointSpec{
				IPNetworks: []cnet.IPNet{cnet.IPNet{*cidr1}, cnet.IPNet{*cidr2}},
				IPNATs: []api.IPNAT{
					{
						InternalIP: cnet.IP{net.ParseIP("10.0.0.0")},
						ExternalIP: cnet.IP{net.ParseIP("20.0.0.0")},
					},
				},

				IPv4Gateway:   &cnet.IP{net.ParseIP("10.0.0.1")},
				IPv6Gateway:   &cnet.IP{net.ParseIP("fe80::33")},
				Profiles:      []string{"profile1", "profile2"},
				InterfaceName: "eth0",
				MAC:           cnet.MAC{mac1},
			},
			api.WorkloadEndpointSpec{
				IPNetworks: []cnet.IPNet{cnet.IPNet{*cidr3}, cnet.IPNet{*cidr4}},
				IPNATs: []api.IPNAT{
					{
						InternalIP: cnet.IP{net.ParseIP("192.168.0.0")},
						ExternalIP: cnet.IP{net.ParseIP("192.168.1.1")},
					},
				},

				IPv4Gateway:   &cnet.IP{net.ParseIP("192.168.0.1")},
				IPv6Gateway:   &cnet.IP{net.ParseIP("fe80::33")},
				Profiles:      []string{"profile3", "profile4"},
				InterfaceName: "eth1",
				MAC:           cnet.MAC{mac2},
			}),

		// Test 2: Pass one partially populated WorkloadEndpointSpec and another fully populated WorkloadEndpointSpec and expect the series of operations to succeed.
		Entry("One partially populated WorkloadEndpointSpec and another fully populated WorkloadEndpointSpec",
			api.WorkloadEndpointMetadata{
				Name:         "host1",
				Workload:     "workload1",
				Orchestrator: "kubernetes",
				Node:         "node1",
				Labels: map[string]string{
					"app":  "app-abc",
					"prod": "no",
				}},
			api.WorkloadEndpointMetadata{
				Name:         "host2",
				Workload:     "workload2",
				Orchestrator: "mesos",
				Node:         "node2",
				Labels: map[string]string{
					"app":  "app-xyz",
					"prod": "yes",
				}},
			api.WorkloadEndpointSpec{
				IPNetworks: []cnet.IPNet{cnet.IPNet{*cidr1}, cnet.IPNet{*cidr2}},
				IPNATs: []api.IPNAT{
					{
						InternalIP: cnet.IP{net.ParseIP("10.0.0.0")},
					},
				},
				InterfaceName: "eth1",
			},
			api.WorkloadEndpointSpec{
				IPNetworks: []cnet.IPNet{cnet.IPNet{*cidr3}, cnet.IPNet{*cidr4}},
				IPNATs: []api.IPNAT{
					{
						InternalIP: cnet.IP{net.ParseIP("192.168.0.0")},
						ExternalIP: cnet.IP{net.ParseIP("192.168.1.1")},
					},
				},

				IPv4Gateway:   &cnet.IP{net.ParseIP("192.168.0.1")},
				IPv6Gateway:   &cnet.IP{net.ParseIP("fe80::33")},
				Profiles:      []string{"profile3", "profile4"},
				InterfaceName: "eth1",
				MAC:           cnet.MAC{mac2},
			}),

		// Test 3: Pass one fully populated WorkloadEndpointSpec and another empty WorkloadEndpointSpec and expect the series of operations to succeed.
		Entry("One fully populated WorkloadEndpointSpec and another empty WorkloadEndpointSpec",
			api.WorkloadEndpointMetadata{
				Name:         "host1",
				Workload:     "workload1",
				Orchestrator: "kubernetes",
				Node:         "node1",
				Labels: map[string]string{
					"app":  "app-abc",
					"prod": "no",
				}},
			api.WorkloadEndpointMetadata{
				Name:         "host2",
				Workload:     "workload2",
				Orchestrator: "mesos",
				Node:         "node2",
				Labels: map[string]string{
					"app":  "app-xyz",
					"prod": "yes",
				}},
			api.WorkloadEndpointSpec{
				IPNetworks: []cnet.IPNet{cnet.IPNet{*cidr1}, cnet.IPNet{*cidr2}},
				IPNATs: []api.IPNAT{
					{
						InternalIP: cnet.IP{net.ParseIP("10.0.0.0")},
						ExternalIP: cnet.IP{net.ParseIP("20.0.0.0")},
					},
				},

				IPv4Gateway:   &cnet.IP{net.ParseIP("10.0.0.1")},
				IPv6Gateway:   &cnet.IP{net.ParseIP("fe80::33")},
				Profiles:      []string{"profile1", "profile2"},
				InterfaceName: "eth0",
				MAC:           cnet.MAC{mac1},
			},
			api.WorkloadEndpointSpec{}),
	)

})
