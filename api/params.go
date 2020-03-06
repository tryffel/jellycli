package api

import (
	"strconv"
	"tryffel.net/go/jellycli/interfaces"
)

type params map[string]string

// get pointer to map for convinience
func (p *params) ptr() map[string]string {
	return *p
}

func (p *params) setPaging(paging interfaces.Paging) {
	ptr := p.ptr()
	ptr["Limit"] = strconv.Itoa(paging.PageSize)
	ptr["StartIndex"] = strconv.Itoa(paging.Offset())
}

func (p *params) setIncludeTypes(itemType mediaItemType) {
	ptr := p.ptr()
	ptr["IncludeItemTypes"] = itemType.String()
}
