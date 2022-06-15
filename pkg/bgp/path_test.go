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
	}
	t.Parallel()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			sortedPathes, reason := sortPathes(tt.pathes)
			assert.Equal(t, tt.reason, reason)
			idList := []int{}
			for _, p := range sortedPathes {
				idList = append(idList, p.id)
			}
			assert.Equal(t, tt.expIdList, idList)
		})
	}
}
