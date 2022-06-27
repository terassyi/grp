package bgp

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortPathes(t *testing.T) {
	tests := []struct {
		name      string
		pathes    []*Path
		reason    BestPathSelectionReason
		expIdList []int
	}{
		{
			name: "Only path",
			pathes: []*Path{
				{
					id: 1,
				},
			},
			reason:    REASON_ONLY_PATH,
			expIdList: []int{1},
		},
		{
			name: "AS path",
			pathes: []*Path{
				{
					id: 1,
					info: &peerInfo{
						neighbor: &neighbor{
							addr:     net.ParseIP("10.0.0.3"),
							port:     179,
							as:       200,
							routerId: net.ParseIP("2.2.2.2"),
						},
						as:       100,
						routerId: net.ParseIP("1.1.1.1"),
					},
					as:      200,
					asPath:  ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 300, 400}}}},
					nextHop: net.ParseIP("10.0.0.3"),
				},
				{
					id: 2,
					info: &peerInfo{
						neighbor: &neighbor{
							addr:     net.ParseIP("10.0.0.3"),
							port:     179,
							as:       200,
							routerId: net.ParseIP("2.2.2.2"),
						},
						as:       100,
						routerId: net.ParseIP("1.1.1.1"),
					},
					as:      200,
					asPath:  ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 400}}}},
					nextHop: net.ParseIP("10.0.0.3"),
				},
			},
			reason:    REASON_AS_PATH_ATTR,
			expIdList: []int{2, 1},
		},
		{
			name: "Local pref",
			pathes: []*Path{
				{
					id: 1,
					info: &peerInfo{
						neighbor: &neighbor{
							addr:     net.ParseIP("10.0.0.3"),
							port:     179,
							as:       200,
							routerId: net.ParseIP("2.2.2.2"),
						},
						as:       100,
						routerId: net.ParseIP("1.1.1.1"),
					},
					as:        200,
					asPath:    ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 300, 400}}}},
					nextHop:   net.ParseIP("10.0.0.3"),
					localPref: 100,
				},
				{
					id: 2,
					info: &peerInfo{
						neighbor: &neighbor{
							addr:     net.ParseIP("10.0.0.3"),
							port:     179,
							as:       200,
							routerId: net.ParseIP("2.2.2.2"),
						},
						as:       100,
						routerId: net.ParseIP("1.1.1.1"),
					},
					as:        200,
					asPath:    ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 400}}}},
					nextHop:   net.ParseIP("10.0.0.3"),
					localPref: 100,
				},
				{
					id: 3,
					info: &peerInfo{
						neighbor: &neighbor{
							addr:     net.ParseIP("10.0.0.3"),
							port:     179,
							as:       200,
							routerId: net.ParseIP("2.2.2.2"),
						},
						as:       100,
						routerId: net.ParseIP("1.1.1.1"),
					},
					as:        200,
					asPath:    ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 300, 400}}}},
					nextHop:   net.ParseIP("10.0.0.3"),
					localPref: 110,
				},
			},
			reason:    REASON_LOCAL_PREF_ATTR,
			expIdList: []int{3, 2, 1},
		},
		{
			name: "Local originated",
			pathes: []*Path{
				{
					id: 1,
					info: &peerInfo{
						neighbor: &neighbor{
							addr:     net.ParseIP("10.0.0.3"),
							port:     179,
							as:       200,
							routerId: net.ParseIP("2.2.2.2"),
						},
						as:       100,
						routerId: net.ParseIP("1.1.1.1"),
					},
					as:        200,
					asPath:    ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 300, 400}}}},
					nextHop:   net.ParseIP("10.0.0.3"),
					localPref: 100,
				},
				{
					id: 2,
					// info: &peerInfo{
					// 	neighbor: &neighbor{
					// 		addr:     net.ParseIP("10.0.0.3"),
					// 		port:     179,
					// 		as:       200,
					// 		routerId: net.ParseIP("2.2.2.2"),
					// 	},
					// 	as:       100,
					// 	routerId: net.ParseIP("1.1.1.1"),
					// },
					as: 100,
					// asPath:    &ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{}}}},
					nextHop:   net.ParseIP("10.0.0.3"),
					localPref: 100,
					local:     true,
				},
				{
					id: 3,
					info: &peerInfo{
						neighbor: &neighbor{
							addr:     net.ParseIP("10.0.0.3"),
							port:     179,
							as:       200,
							routerId: net.ParseIP("2.2.2.2"),
						},
						as:       100,
						routerId: net.ParseIP("1.1.1.1"),
					},
					as:        200,
					asPath:    ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 400}}}},
					nextHop:   net.ParseIP("10.0.0.3"),
					localPref: 100,
				},
			},
			reason:    REASON_LOCAL_ORIGINATED,
			expIdList: []int{2, 3, 1},
		},
		{
			name: "Origin",
			pathes: []*Path{
				{
					id: 1,
					info: &peerInfo{
						neighbor: &neighbor{
							addr:     net.ParseIP("10.0.0.3"),
							port:     179,
							as:       200,
							routerId: net.ParseIP("2.2.2.2"),
						},
						as:       100,
						routerId: net.ParseIP("1.1.1.1"),
					},
					as:        200,
					asPath:    ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 300, 400}}}},
					nextHop:   net.ParseIP("10.0.0.3"),
					localPref: 100,
				},
				{
					id: 2,
					info: &peerInfo{
						neighbor: &neighbor{
							addr:     net.ParseIP("10.0.0.3"),
							port:     179,
							as:       200,
							routerId: net.ParseIP("2.2.2.2"),
						},
						as:       100,
						routerId: net.ParseIP("1.1.1.1"),
					},
					as:        200,
					asPath:    ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 400}}}},
					nextHop:   net.ParseIP("10.0.0.3"),
					localPref: 100,
				},
				{
					id: 3,
					info: &peerInfo{
						neighbor: &neighbor{
							addr:     net.ParseIP("10.0.0.3"),
							port:     179,
							as:       200,
							routerId: net.ParseIP("2.2.2.2"),
						},
						as:       100,
						routerId: net.ParseIP("1.1.1.1"),
					},
					as:        200,
					asPath:    ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 300, 400}}}},
					nextHop:   net.ParseIP("10.0.0.3"),
					localPref: 110,
				},
				{
					id: 4,
					info: &peerInfo{
						neighbor: &neighbor{
							addr:     net.ParseIP("10.0.0.3"),
							port:     179,
							as:       200,
							routerId: net.ParseIP("2.2.2.2"),
						},
						as:       100,
						routerId: net.ParseIP("1.1.1.1"),
					},
					as:        200,
					asPath:    ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 300, 400}}}},
					origin:    Origin{value: ORIGIN_IGP},
					nextHop:   net.ParseIP("10.0.0.3"),
					localPref: 100,
				},
				{
					id: 5,
					info: &peerInfo{
						neighbor: &neighbor{
							addr:     net.ParseIP("10.0.0.3"),
							port:     179,
							as:       200,
							routerId: net.ParseIP("2.2.2.2"),
						},
						as:       100,
						routerId: net.ParseIP("1.1.1.1"),
					},
					as:        200,
					asPath:    ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 600, 400}}}},
					origin:    Origin{value: ORIGIN_EGP},
					nextHop:   net.ParseIP("10.0.0.3"),
					localPref: 100,
				},
			},
			reason:    REASON_LOCAL_PREF_ATTR,
			expIdList: []int{3, 2, 1, 4, 5},
		},
		{
			name: "Multi Exit Disc",
			pathes: []*Path{
				{
					id: 1,
					info: &peerInfo{
						neighbor: &neighbor{
							addr:     net.ParseIP("10.0.0.3"),
							port:     179,
							as:       200,
							routerId: net.ParseIP("2.2.2.2"),
						},
						as:       100,
						routerId: net.ParseIP("1.1.1.1"),
					},
					as:        200,
					asPath:    ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 300, 400}}}},
					nextHop:   net.ParseIP("10.0.0.3"),
					localPref: 100,
				},
				{
					id: 2,
					info: &peerInfo{
						neighbor: &neighbor{
							addr:     net.ParseIP("10.0.0.3"),
							port:     179,
							as:       200,
							routerId: net.ParseIP("2.2.2.2"),
						},
						as:       100,
						routerId: net.ParseIP("1.1.1.1"),
					},
					as:        200,
					asPath:    ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 400}}}},
					nextHop:   net.ParseIP("10.0.0.3"),
					localPref: 100,
				},
				{
					id: 3,
					info: &peerInfo{
						neighbor: &neighbor{
							addr:     net.ParseIP("10.0.0.3"),
							port:     179,
							as:       200,
							routerId: net.ParseIP("2.2.2.2"),
						},
						as:       100,
						routerId: net.ParseIP("1.1.1.1"),
					},
					as:        200,
					asPath:    ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 300, 400}}}},
					nextHop:   net.ParseIP("10.0.0.3"),
					localPref: 100,
				},
				{
					id: 4,
					info: &peerInfo{
						neighbor: &neighbor{
							addr:     net.ParseIP("10.0.0.3"),
							port:     179,
							as:       200,
							routerId: net.ParseIP("2.2.2.2"),
						},
						as:       100,
						routerId: net.ParseIP("1.1.1.1"),
					},
					as:             200,
					asPath:         ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 300, 400}}}},
					origin:         Origin{value: ORIGIN_EGP},
					nextHop:        net.ParseIP("10.0.0.3"),
					localPref:      100,
					pathAttributes: []PathAttr{&MultiExitDisc{discriminator: 10}},
				},
				{
					id: 5,
					info: &peerInfo{
						neighbor: &neighbor{
							addr:     net.ParseIP("10.0.0.3"),
							port:     179,
							as:       200,
							routerId: net.ParseIP("2.2.2.2"),
						},
						as:       100,
						routerId: net.ParseIP("1.1.1.1"),
					},
					as:             200,
					asPath:         ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 600, 400}}}},
					origin:         Origin{value: ORIGIN_EGP},
					nextHop:        net.ParseIP("10.0.0.3"),
					localPref:      100,
					pathAttributes: []PathAttr{&MultiExitDisc{discriminator: 0}},
				},
			},
			reason:    REASON_AS_PATH_ATTR,
			expIdList: []int{2, 1, 3, 5, 4},
		},
		{
			name: "Router ID",
			pathes: []*Path{
				{
					id: 1,
					info: &peerInfo{
						neighbor: &neighbor{
							addr:     net.ParseIP("10.0.0.3"),
							port:     179,
							as:       500,
							routerId: net.ParseIP("5.5.5.5"),
						},
						as:       100,
						routerId: net.ParseIP("1.1.1.1"),
					},
					as:        200,
					asPath:    ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 300, 400}}}},
					nextHop:   net.ParseIP("10.0.0.3"),
					localPref: 100,
				},
				{
					id: 2,
					info: &peerInfo{
						neighbor: &neighbor{
							addr:     net.ParseIP("10.0.0.3"),
							port:     179,
							as:       300,
							routerId: net.ParseIP("3.3.3.3"),
						},
						as:       100,
						routerId: net.ParseIP("1.1.1.1"),
					},
					as:        200,
					asPath:    ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{400, 700, 300}}}},
					nextHop:   net.ParseIP("10.0.0.3"),
					localPref: 100,
				},
				{
					id: 3,
					info: &peerInfo{
						neighbor: &neighbor{
							addr:     net.ParseIP("10.0.0.3"),
							port:     179,
							as:       400,
							routerId: net.ParseIP("4.4.4.4"),
						},
						as:       100,
						routerId: net.ParseIP("1.1.1.1"),
					},
					as:        200,
					asPath:    ASPath{Segments: []*ASPathSegment{{Type: SEG_TYPE_AS_SEQUENCE, AS2: []uint16{200, 300, 400}}}},
					nextHop:   net.ParseIP("10.0.0.3"),
					localPref: 100,
				},
			},
			reason:    REASON_BGP_PEER_ROUTER_ID,
			expIdList: []int{2, 3, 1},
		},
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			sortedPathes, reason := sortPathes(tt.pathes)
			assert.Equal(t, tt.reason.String(), reason.String())
			idList := []int{}
			reasonList := []BestPathSelectionReason{}
			for _, p := range sortedPathes {
				idList = append(idList, p.id)
				reasonList = append(reasonList, p.reason)
			}
			t.Logf("sorted pathlist want:%v, actual:%v", tt.expIdList, idList)
			t.Logf("sorted pathlist reason: %v", reasonList)
			assert.Equal(t, tt.expIdList, idList)
		})
	}
}
