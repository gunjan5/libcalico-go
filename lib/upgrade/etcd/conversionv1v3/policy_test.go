// Copyright (c) 2017 Tigera, Inc. All rights reserved.

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

package conversionv1v3

import (
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1 "github.com/projectcalico/libcalico-go/lib/apis/v1"
	"github.com/projectcalico/libcalico-go/lib/apis/v1/unversioned"
	apiv3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/backend/model"
)

var order1 = 1000.00
var order2 = 999.99
var policyTable = []struct {
	description string
	v1API       unversioned.Resource
	v1KVP       *model.KVPair
	v3API       apiv3.GlobalNetworkPolicy
}{
	{
		description: "fully populated Policy",
		v1API: apiv1.Policy{
			Metadata: apiv1.PolicyMetadata{
				Name: "namyMcPolicyName",
			},
			Spec: apiv1.PolicySpec{
				Order:          &order1,
				IngressRules:   []apiv1.Rule{V1InRule1, V1InRule2},
				EgressRules:    []apiv1.Rule{V1EgressRule1, V1EgressRule2},
				Selector:       "thing == 'value'",
				DoNotTrack:     true,
				PreDNAT:        true,
				ApplyOnForward: true,
				Types:          []apiv1.PolicyType{apiv1.PolicyTypeIngress},
			},
		},
		v1KVP: &model.KVPair{
			Key: model.PolicyKey{
				Name: "namyMcPolicyName",
			},
			Value: &model.Policy{
				Order:          &order1,
				InboundRules:   []model.Rule{V1ModelInRule1, V1ModelInRule2},
				OutboundRules:  []model.Rule{V1ModelEgressRule1, V1ModelEgressRule2},
				Selector:       "thing == 'value'",
				DoNotTrack:     true,
				PreDNAT:        true,
				ApplyOnForward: true,
				Types:          []string{"ingress"},
			},
		},
		v3API: apiv3.GlobalNetworkPolicy{
			ObjectMeta: v1.ObjectMeta{
				Name: "namymcpolicyname",
			},
			Spec: apiv3.GlobalNetworkPolicySpec{
				Order:          &order1,
				Ingress:        []apiv3.Rule{V3InRule1, V3InRule2},
				Egress:         []apiv3.Rule{V3EgressRule1, V3EgressRule2},
				Selector:       "thing == 'value'",
				DoNotTrack:     true,
				PreDNAT:        true,
				ApplyOnForward: true,
				Types:          []apiv3.PolicyType{apiv3.PolicyTypeIngress},
			},
		},
	},
	{
		description: "policy name conversion",
		v1API: apiv1.Policy{
			Metadata: apiv1.PolicyMetadata{
				Name: "MaKe.-.MaKe",
			},
			Spec: apiv1.PolicySpec{
				Order:          &order1,
				IngressRules:   []apiv1.Rule{V1InRule2},
				EgressRules:    []apiv1.Rule{V1EgressRule1},
				Selector:       "thing == 'value'",
				DoNotTrack:     true,
				PreDNAT:        true,
				ApplyOnForward: true,
				Types:          []apiv1.PolicyType{apiv1.PolicyTypeIngress},
			},
		},
		v1KVP: &model.KVPair{
			Key: model.PolicyKey{
				Name: "MaKe.-.MaKe",
			},
			Value: &model.Policy{
				Order:          &order1,
				InboundRules:   []model.Rule{V1ModelInRule2},
				OutboundRules:  []model.Rule{V1ModelEgressRule1},
				Selector:       "thing == 'value'",
				DoNotTrack:     true,
				PreDNAT:        true,
				ApplyOnForward: true,
				Types:          []string{"ingress"},
			},
		},
		v3API: apiv3.GlobalNetworkPolicy{
			ObjectMeta: v1.ObjectMeta{
				Name: "make.make",
			},
			Spec: apiv3.GlobalNetworkPolicySpec{
				Order:          &order1,
				Ingress:        []apiv3.Rule{V3InRule2},
				Egress:         []apiv3.Rule{V3EgressRule1},
				Selector:       "thing == 'value'",
				DoNotTrack:     true,
				PreDNAT:        true,
				ApplyOnForward: true,
				Types:          []apiv3.PolicyType{apiv3.PolicyTypeIngress},
			},
		},
	},
	{
		description: "policy with ApplyOnForward set to it's zero value (missing), and PreDNAT and DoNotTrack set to true " +
			"should convert ApplyOnForward to true in v3 API",
		v1API: apiv1.Policy{
			Metadata: apiv1.PolicyMetadata{
				Name: "RAWR",
			},
			Spec: apiv1.PolicySpec{
				Order:          &order1,
				IngressRules:   []apiv1.Rule{V1InRule2},
				DoNotTrack:     true,
				PreDNAT:        true,
				ApplyOnForward: false,
				Types:          []apiv1.PolicyType{apiv1.PolicyTypeIngress},
			},
		},
		v1KVP: &model.KVPair{
			Key: model.PolicyKey{
				Name: "RAWR",
			},
			Value: &model.Policy{
				Order:          &order1,
				InboundRules:   []model.Rule{V1ModelInRule2},
				OutboundRules:  []model.Rule{},
				DoNotTrack:     true,
				PreDNAT:        true,
				ApplyOnForward: false,
				Types:          []string{"ingress"},
			},
		},
		v3API: apiv3.GlobalNetworkPolicy{
			ObjectMeta: v1.ObjectMeta{
				Name: "rawr",
			},
			Spec: apiv3.GlobalNetworkPolicySpec{
				Order:          &order1,
				Ingress:        []apiv3.Rule{V3InRule2},
				Egress:         []apiv3.Rule{},
				DoNotTrack:     true,
				PreDNAT:        true,
				ApplyOnForward: true, // notice this gets converted to true, because DoNotTrack and PreDNAT are true.
				Types:          []apiv3.PolicyType{apiv3.PolicyTypeIngress},
			},
		},
	},
	{
		description: "policy with ApplyOnForward set to it's zero value (missing), and PreDNAT and DoNotTrack both " +
			"set to false should NOT convert ApplyOnForward to true in v3 API",
		v1API: apiv1.Policy{
			Metadata: apiv1.PolicyMetadata{
				Name: "meow",
			},
			Spec: apiv1.PolicySpec{
				Order:          &order1,
				IngressRules:   []apiv1.Rule{V1InRule2},
				DoNotTrack:     false,
				PreDNAT:        false,
				ApplyOnForward: true,
				Types:          []apiv1.PolicyType{apiv1.PolicyTypeIngress},
			},
		},
		v1KVP: &model.KVPair{
			Key: model.PolicyKey{
				Name: "meow",
			},
			Value: &model.Policy{
				Order:          &order1,
				InboundRules:   []model.Rule{V1ModelInRule2},
				OutboundRules:  []model.Rule{},
				DoNotTrack:     false,
				PreDNAT:        false,
				ApplyOnForward: true,
				Types:          []string{"ingress"},
			},
		},
		v3API: apiv3.GlobalNetworkPolicy{
			ObjectMeta: v1.ObjectMeta{
				Name: "meow",
			},
			Spec: apiv3.GlobalNetworkPolicySpec{
				Order:          &order1,
				Ingress:        []apiv3.Rule{V3InRule2},
				Egress:         []apiv3.Rule{},
				DoNotTrack:     false,
				PreDNAT:        false,
				ApplyOnForward: true, // notice this gets converted to true, because DoNotTrack and PreDNAT are true.
				Types:          []apiv3.PolicyType{apiv3.PolicyTypeIngress},
			},
		},
	},
	{
		description: "policy with non-strictly masked CIDR should get converted to strictly masked CIDR in v3 API",
		v1API: apiv1.Policy{
			Metadata: apiv1.PolicyMetadata{
				Name: "MaKe.-.MaKe",
			},
			Spec: apiv1.PolicySpec{
				Order: &order1,
				// Source Nets selector in V1InRule1 and V1EgressRule1 are non-strictly masked CIDRs.
				IngressRules:   []apiv1.Rule{V1InRule1},
				EgressRules:    []apiv1.Rule{V1EgressRule1},
				Selector:       "thing == 'value'",
				DoNotTrack:     true,
				PreDNAT:        true,
				ApplyOnForward: true,
				Types:          []apiv1.PolicyType{apiv1.PolicyTypeIngress, apiv1.PolicyTypeEgress},
			},
		},
		v1KVP: &model.KVPair{
			Key: model.PolicyKey{
				Name: "MaKe.-.MaKe",
			},
			Value: &model.Policy{
				Order: &order1,
				// Source Nets selector in V1ModelInRule1 and V1ModelEgressRule1 are non-strictly masked CIDRs.
				InboundRules:   []model.Rule{V1ModelInRule1},
				OutboundRules:  []model.Rule{V1ModelEgressRule1},
				Selector:       "thing == 'value'",
				DoNotTrack:     true,
				PreDNAT:        true,
				ApplyOnForward: true,
				Types:          []string{"ingress", "egress"},
			},
		},
		v3API: apiv3.GlobalNetworkPolicy{
			ObjectMeta: v1.ObjectMeta{
				Name: "make.make",
			},
			Spec: apiv3.GlobalNetworkPolicySpec{
				Order: &order1,
				// Source Nets selector in V3InRule1 and V3EgressRule1 are strictly masked CIDRs.
				Ingress:        []apiv3.Rule{V3InRule1},
				Egress:         []apiv3.Rule{V3EgressRule1},
				Selector:       "thing == 'value'",
				DoNotTrack:     true,
				PreDNAT:        true,
				ApplyOnForward: true,
				Types:          []apiv3.PolicyType{apiv3.PolicyTypeIngress, apiv3.PolicyTypeEgress},
			},
		},
	},
}

func TestCanConvertV1ToV3Policy(t *testing.T) {

	for _, entry := range policyTable {
		t.Run(entry.description, func(t *testing.T) {
			RegisterTestingT(t)

			p := Policy{}

			// Test and assert v1 API to v1 backend logic.
			v1KVPResult, err := p.APIV1ToBackendV1(entry.v1API)
			Expect(err).NotTo(HaveOccurred(), entry.description)
			Expect(v1KVPResult.Key.(model.PolicyKey).Name).To(Equal(entry.v1KVP.Key.(model.PolicyKey).Name))
			Expect(v1KVPResult.Value.(*model.Policy)).To(Equal(entry.v1KVP.Value))

			// Test and assert v1 backend to v3 API logic.
			v3APIResult, err := p.BackendV1ToAPIV3(entry.v1KVP)
			Expect(err).NotTo(HaveOccurred(), entry.description)
			Expect(v3APIResult.(*apiv3.GlobalNetworkPolicy).Name).To(Equal(entry.v3API.Name), entry.description)
			Expect(v3APIResult.(*apiv3.GlobalNetworkPolicy).Spec).To(Equal(entry.v3API.Spec), entry.description)
		})
	}
}
