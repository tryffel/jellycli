package jellyfin

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

func (p *params) setLimit(n int) {
	(*p)["Limit"] = strconv.Itoa(n)
}

func (p *params) setIncludeTypes(itemType mediaItemType) {
	ptr := p.ptr()
	ptr["IncludeItemTypes"] = itemType.String()
}

func (p *params) enableRecursive() {
	(*p)["Recursive"] = "true"
}

func (p *params) setParentId(id string) {
	(*p)["ParentId"] = id
}

func (p *params) setSorting(name string, order string) {
	(*p)["SortBy"] = name
	(*p)["SortOrder"] = order
}
