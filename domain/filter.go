package domain

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Filter struct {
	Archived *bool   `json:"archived"`
	State    *string `json:"state"`
}

func NewFilter() *Filter {
	return new(Filter)
}

func (f *Filter) FromReq(r *http.Request) *Filter {
	archivedStr := r.URL.Query().Get("archived")
	if len(archivedStr) > 0 {
		if archived, err := strconv.ParseBool(archivedStr); err == nil {
			f.Archived = &archived
		}
	}
	if state := r.URL.Query().Get("active"); len(state) > 0 {
		f.State = &state
	}
	return f
}

// Since not all filters are supported by all tables,
// a filter-user can mask the filter provided by the caller
func (f *Filter) Mask(mask Filter) *Filter {
	if f == nil {
		return f
	}
	if mask.Archived == nil {
		f.Archived = nil
	}
	if mask.State == nil {
		f.State = nil
	}
	return f
}

func (f *Filter) ToWhere() string {
	if f == nil {
		return ""
	}
	conditions := []string{}
	if f.Archived != nil {
		comp := "IS"
		if *f.Archived {
			comp = "IS NOT"
		}
		conditions = append(conditions, fmt.Sprintf("archived_at %s NULL", comp))
	}
	if f.State != nil {
		conditions = append(conditions, fmt.Sprintf("state = %v", f.State))
	}
	if len(conditions) == 0 {
		return ""
	}
	return fmt.Sprintf("(%s)", strings.Join(conditions, " AND "))
}
