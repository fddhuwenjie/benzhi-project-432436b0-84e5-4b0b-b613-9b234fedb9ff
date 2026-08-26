package audit

import "sort"

func StableEvents(in []Event) []Event {
	out := append([]Event(nil), in...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Revision < out[j].Revision })
	return out
}
