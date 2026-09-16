package bsp

import (
	"fmt"

	bpb "github.com/packetflinger/libq2/proto/v1"
)

// areaSize is the on-disk size of an area_t: numareaportals + firstareaportal.
const areaSize = 8

// areaPortalSize is the on-disk size of a dareaportal_t: portalnum + otherarea.
const areaPortalSize = 8

func parseAreas(data []byte) ([]*bpb.BSPArea, error) {
	if len(data)%areaSize != 0 {
		return nil, fmt.Errorf("areas lump size %d is not a multiple of %d", len(data), areaSize)
	}
	count := len(data) / areaSize
	areas := make([]*bpb.BSPArea, 0, count)
	r := newReader(data)
	for i := 0; i < count; i++ {
		areas = append(areas, &bpb.BSPArea{
			NumAreaPortals:  r.int32Val(),
			FirstAreaPortal: r.int32Val(),
		})
	}
	return areas, r.err
}

func marshalAreas(areas []*bpb.BSPArea) []byte {
	w := &writer{}
	for _, a := range areas {
		w.int32Val(a.GetNumAreaPortals())
		w.int32Val(a.GetFirstAreaPortal())
	}
	return w.buf.Bytes()
}

func parseAreaPortals(data []byte) ([]*bpb.BSPAreaPortal, error) {
	if len(data)%areaPortalSize != 0 {
		return nil, fmt.Errorf("area portals lump size %d is not a multiple of %d", len(data), areaPortalSize)
	}
	count := len(data) / areaPortalSize
	portals := make([]*bpb.BSPAreaPortal, 0, count)
	r := newReader(data)
	for i := 0; i < count; i++ {
		portals = append(portals, &bpb.BSPAreaPortal{
			PortalNum: r.int32Val(),
			OtherArea: r.int32Val(),
		})
	}
	return portals, r.err
}

func marshalAreaPortals(portals []*bpb.BSPAreaPortal) []byte {
	w := &writer{}
	for _, p := range portals {
		w.int32Val(p.GetPortalNum())
		w.int32Val(p.GetOtherArea())
	}
	return w.buf.Bytes()
}
