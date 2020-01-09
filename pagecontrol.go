package gowebtable

import (
	"math"
)

//PageDetails defines the struct for pageination control information
type PageDetails struct {
	Table  string
	URL    string
	Target string

	FilterTerms []FilterTerm `json:"filterterms,omitempty"`
	OrderTerms  []OrderTerm  `json:"orderterms,omitempty"`

	FilterDefaultElement string

	OrderDefaultElement   string
	OrderDefaultDirection string

	Page   int `json:"page,omitempty"`
	Offset int `json:"-"`
	Limit  int `json:"limit,omitempty"`

	LimitOptions []int
	LimitDefault int

	TotalFiltered int
	TotalAll      int

	RecordFirst int
	RecordLast  int

	PageCount    int
	PagePrevious int
	PageNext     int

	IsFiltered bool
}

//FilterTerm defines the struct for a filter term record
type FilterTerm struct {
	Element string `json:"element"`
	Term    string `json:"term"`
}

//OrderTerm defines the struct for an order term record
type OrderTerm struct {
	Element   string `json:"element,omitempty"`
	Direction string `json:"direction,omitempty"`
}

//PreCalculate performs a calculation of the offset prior to gathering records.
//This should be called after getting request details from the client but before starting to get totals or records from the database.
func (pd *PageDetails) PreCalculate() {

	pd.Table = "tbl"

	if pd.Page == 0 {
		pd.Page = 1
	}

	//Limit / Offset
	if pd.Limit == 0 {
		if pd.LimitDefault > 0 {
			pd.Limit = pd.LimitDefault
		} else {
			pd.Limit = 10
		}

	}

	//Set default limit options if not set
	if len(pd.LimitOptions) == 0 {
		pd.LimitOptions = []int{5, 10, 25, 50, 100, 250, 500, 1000}
	}

	pd.Offset = (pd.Page - 1) * pd.Limit

	//Add default sort
	if len(pd.OrderTerms) == 0 && pd.OrderDefaultElement != "" {
		if pd.OrderDefaultDirection == "" {
			pd.OrderDefaultDirection = "asc"
		}
		pd.OrderTerms = append(pd.OrderTerms, OrderTerm{Element: pd.OrderDefaultElement, Direction: pd.OrderDefaultDirection})
	}

	if len(pd.FilterTerms) > 0 {
		pd.IsFiltered = true
	}

}

//Calculate performs a calculation of all the page and record count details prior to display.
//This should be called after totals have been gathered but prior to getting records. Limit and offset from this record should be passed into the database query.
func (pd *PageDetails) Calculate() {

	//Record First
	pd.RecordFirst = pd.Offset + 1

	//Calculate page count
	pd.PageCount = int(math.Ceil(float64(pd.TotalFiltered) / float64(pd.Limit)))

	//For no records
	if pd.PageCount == 0 {
		pd.PageCount = 1
		pd.RecordFirst = 0
	}

	//Calculate the current page
	pd.Page = int(1 + (float64(pd.Offset) / float64(pd.Limit)))

	//Record Last
	pd.RecordLast = pd.Offset + pd.Limit
	if pd.RecordLast > pd.TotalFiltered {
		pd.RecordLast = pd.TotalFiltered
	}

	//Page Previous
	if pd.Page > 1 {
		pd.PagePrevious = pd.Page - 1
	}

	//Page Next
	if pd.Page != pd.PageCount {
		pd.PageNext = pd.Page + 1
	} else {
		pd.PageNext = pd.PageCount
	}

}
