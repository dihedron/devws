package dto

import (
	"math"

	"github.com/dihedron/devws/openstack"
)

type TableData struct {
	Records      []openstack.Workstation
	Page         int
	TotalPages   int
	TotalRecords int
	PrevPage     int
	NextPage     int
	Pages        []int
}

func NewTableData(vms []openstack.Workstation, page int) *TableData {

	td := &TableData{}
	return td.Paginate(vms, page)

}

const pageSize = 10

func (t *TableData) Paginate(vms []openstack.Workstation, page int) *TableData {
	total := len(vms)
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}

	// Build page number slice
	pages := make([]int, totalPages)
	for i := range pages {
		pages[i] = i + 1
	}

	prev := page - 1
	if prev < 1 {
		prev = 1
	}
	next := page + 1
	if next > totalPages {
		next = totalPages
	}

	if total > 0 {
		t.Records = vms[start:end]
	} else {
		t.Records = vms
	}
	t.Page = page
	t.TotalPages = totalPages
	t.TotalRecords = total
	t.PrevPage = prev
	t.NextPage = next
	t.Pages = pages

	return t
}
