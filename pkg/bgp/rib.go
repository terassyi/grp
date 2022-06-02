package bgp

// Routing Information Base
type Rib interface{}

// Adj-RIB-In: The Adj-RIB-In store routing information that has been learned from inbound UPDATE messages.
// 	           Their contents represent routes that are available as an input to the Decision Process.
type AdjRibIn struct{}

// Loc-RIB: The Loc-RIB contains the local routing information that the BGP speaker has selected by applying its local policies
//          to the routing information contained in its Adj-RIB-In.
type LocRib struct{}

// Adj-RIB-Out: The Adj-RIB-Out store the information that the local routing information that the BGP speaker has selected
//              for advertisement to its peers.
//              The routing information stored in the Adj-RIB-Out will be carried in the local BGP speaker's UPDATE messages and advertised to its peers.
type AdjRibOut struct{}

func (in *AdjRibIn) DecisionProcess() (any, error) {
	return nil, nil
}
