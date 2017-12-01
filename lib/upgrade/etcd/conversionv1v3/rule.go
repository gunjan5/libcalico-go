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
	"fmt"
	"sync"
	"strings"

	log "github.com/sirupsen/logrus"

	apiv1 "github.com/projectcalico/libcalico-go/lib/apis/v1"
	apiv3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/backend/model"
	"github.com/projectcalico/libcalico-go/lib/net"
	"github.com/projectcalico/libcalico-go/lib/numorstring"
)

// rulesAPIToBackend converts an API Rule structure slice to a Backend Rule structure slice.
func rulesAPIToBackend(ars []apiv1.Rule) []model.Rule {
	if ars == nil {
		return []model.Rule{}
	}

	brs := make([]model.Rule, len(ars))
	for idx, ar := range ars {
		brs[idx] = ruleAPIToBackend(ar)
	}
	return brs
}

// rulesV1BackendToV3API converts a Backend Rule structure slice to an API Rule structure slice.
func rulesV1BackendToV3API(brs []model.Rule) []apiv3.Rule {
	if brs == nil {
		return nil
	}

	ars := make([]apiv3.Rule, len(brs))
	for idx, br := range brs {
		ars[idx] = ruleBackendToAPI(br)
	}
	return ars
}

var logDeprecationOnce sync.Once

// ruleAPIToBackend converts an API Rule structure to a Backend Rule structure.
func ruleAPIToBackend(ar apiv1.Rule) model.Rule {
	var icmpCode, icmpType, notICMPCode, notICMPType *int
	if ar.ICMP != nil {
		icmpCode = ar.ICMP.Code
		icmpType = ar.ICMP.Type
	}

	if ar.NotICMP != nil {
		notICMPCode = ar.NotICMP.Code
		notICMPType = ar.NotICMP.Type
	}

	if ar.Source.Net != nil || ar.Source.NotNet != nil ||
		ar.Destination.Net != nil || ar.Destination.NotNet != nil {
		logDeprecationOnce.Do(func() {
			log.Warning("The Net and NotNet fields in Source/Destination " +
				"EntityRules are deprecated.  Please use Nets or NotNets.")
		})
	}

	return model.Rule{
		Action:      ruleActionAPIToBackend(ar.Action),
		IPVersion:   ar.IPVersion,
		Protocol:    ar.Protocol,
		ICMPCode:    icmpCode,
		ICMPType:    icmpType,
		NotProtocol: ar.NotProtocol,
		NotICMPCode: notICMPCode,
		NotICMPType: notICMPType,

		SrcTag:      ar.Source.Tag,
		SrcNet:      ar.Source.Net,
		SrcNets:     ar.Source.Nets,
		SrcSelector: ar.Source.Selector,
		SrcPorts:    ar.Source.Ports,
		DstTag:      ar.Destination.Tag,
		DstNet:      normalizeIPNet(ar.Destination.Net),
		DstNets:     normalizeIPNets(ar.Destination.Nets),
		DstSelector: ar.Destination.Selector,
		DstPorts:    ar.Destination.Ports,

		NotSrcTag:      ar.Source.NotTag,
		NotSrcNet:      ar.Source.NotNet,
		NotSrcNets:     ar.Source.NotNets,
		NotSrcSelector: ar.Source.NotSelector,
		NotSrcPorts:    ar.Source.NotPorts,
		NotDstTag:      ar.Destination.NotTag,
		NotDstNet:      normalizeIPNet(ar.Destination.NotNet),
		NotDstNets:     normalizeIPNets(ar.Destination.NotNets),
		NotDstSelector: ar.Destination.NotSelector,
		NotDstPorts:    ar.Destination.NotPorts,
	}
}

// normalizeIPNet converts an IPNet to a network by ensuring the IP address is correctly masked.
func normalizeIPNet(n *net.IPNet) *net.IPNet {
	if n == nil {
		return nil
	}
	return n.Network()
}

// normalizeIPNets converts an []*IPNet to a slice of networks by ensuring the IP addresses
// are correctly masked.
func normalizeIPNets(nets []*net.IPNet) []*net.IPNet {
	if nets == nil {
		return nil
	}
	out := make([]*net.IPNet, len(nets))
	for i, n := range nets {
		out[i] = normalizeIPNet(n)
	}
	return out
}

// ruleBackendToAPI convert a Backend Rule structure to an API Rule structure.
func ruleBackendToAPI(br model.Rule) apiv3.Rule {
	var icmp, notICMP *apiv3.ICMPFields
	if br.ICMPCode != nil || br.ICMPType != nil {
		icmp = &apiv3.ICMPFields{
			Code: br.ICMPCode,
			Type: br.ICMPType,
		}
	}
	if br.NotICMPCode != nil || br.NotICMPType != nil {
		notICMP = &apiv3.ICMPFields{
			Code: br.NotICMPCode,
			Type: br.NotICMPType,
		}
	}

	srcNets := br.AllSrcNets()
	var srcNetsStr []string
	for _, net := range srcNets {
		srcNetsStr = append(srcNetsStr, net.String())
	}

	dstNets := br.AllDstNets()
	var dstNetsStr []string
	for _, net := range dstNets {
		dstNetsStr = append(dstNetsStr, net.String())
	}

	notSrcNets := br.AllNotSrcNets()
	var notSrcNetsStr []string
	for _, net := range notSrcNets {
		notSrcNetsStr = append(notSrcNetsStr, net.String())
	}

	notDstNets := br.AllNotDstNets()
	var notDstNetsStr []string
	for _, net := range notDstNets {
		notDstNetsStr = append(notDstNetsStr, net.String())
	}

	// Tags are deprecated in v3.0+, so we convert Tags to selectors.
	// For example:
	// A rule that looks like this with Calico v1 API:
	// apiv1.Rule{
	//	Action: "allow",
	//	Source: apiv1.EntityRule{
	//		Tag:      "tag1",
	//		Selector: "label1 == 'value1' || make == 'cake'",
	//	},
	//}
	// That rule will be converted to the following for Calico v3 API:
	// apiv3.Rule{
	//	Action: "allow",
	//	Source: apiv3.EntityRule{
	//		Selector: "(label1 == 'value1' || make == 'cake') && tag1 == ''",
	//	},
	//}
	srcSelector := br.SrcSelector
	if br.SrcTag != "" {
		srcSelector = fmt.Sprintf("(%s) && %s == ''", br.SrcSelector, br.SrcTag)
	}

	dstSelector := br.DstSelector
	if br.DstTag != "" {
		dstSelector = fmt.Sprintf("(%s) && %s == ''", br.DstSelector, br.DstTag)
	}

	notSrcSelector := br.NotSrcSelector
	if br.NotSrcTag != "" {
		notSrcSelector = fmt.Sprintf("(%s) && %s == ''", br.NotSrcSelector, br.NotSrcTag)
	}

	notDstSelector := br.NotDstSelector
	if br.NotDstTag != "" {
		notDstSelector = fmt.Sprintf("(%s) && %s == ''", br.NotDstSelector, br.NotDstTag)
	}

	v3Protocol := numorstring.ProtocolV3FromProtocolV1(*br.Protocol)

	return apiv3.Rule{
		Action:      ruleActionToV3API(br.Action),
		IPVersion:   br.IPVersion,
		Protocol:    &v3Protocol,
		ICMP:        icmp,
		NotProtocol: br.NotProtocol,
		NotICMP:     notICMP,
		Source: apiv3.EntityRule{
			Nets:        srcNetsStr,
			Selector:    srcSelector,
			Ports:       br.SrcPorts,
			NotNets:     notSrcNetsStr,
			NotSelector: notSrcSelector,
			NotPorts:    br.NotSrcPorts,
		},

		Destination: apiv3.EntityRule{
			Nets:        dstNetsStr,
			Selector:    dstSelector,
			Ports:       br.DstPorts,
			NotNets:     notDstNetsStr,
			NotSelector: notDstSelector,
			NotPorts:    br.NotDstPorts,
		},
	}
}

// ruleActionAPIToBackend converts the rule action field value from the API
// value to the equivalent backend value.
func ruleActionAPIToBackend(action string) string {
	if action == "Pass" {
		return "next-tier"
	}
	return action
}

// ruleActionBackendToAPI converts the rule action field value from the backend
// value to the equivalent API value.
func ruleActionToV3API(inAction string) apiv3.Action {
	if inAction == "" {
		return apiv3.Allow
	} else if inAction == "next-tier" {
		return apiv3.Pass
	} else {
		for _, action := range []apiv3.Action{apiv3.Allow, apiv3.Deny, apiv3.Log, apiv3.Pass} {
			if strings.ToLower(inAction) == strings.ToLower(string(action)) {
				return action
			}
		}
	}

	return apiv3.Action(inAction)
}
