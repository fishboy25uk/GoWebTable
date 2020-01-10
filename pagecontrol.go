package gowebtable

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

var (
	limitDefault        = 10
	limitDefaultOptions = []int{10, 25, 50, 100, 250, 500, 1000}
)

//PageDetails defines the struct for pageination control information
type PageDetails struct {
	Table   string     `json:"table"`
	URL     string     `json:"url"`
	Target  string     `json:"target"`
	Fields  []Field    `json:"-"`
	Results [][]string `json:"-"`

	PageCount    int `json:"pagecount"`
	PageCurrent  int `json:"pagecurrent"`
	PagePrevious int `json:"pageprevious"`
	PageNext     int `json:"pagenext"`

	RecordsTotal int `json:"recordstotal"`
	RecordFirst  int `json:"recordfirst"`
	RecordLast   int `json:"recordlast"`
	RecordCount  int `json:"recordcount"`

	Offset int `json:"offset"`

	Limit        int   `json:"limit"`
	LimitOptions []int `json:"-"`

	OrderTerms            []OrderTerm `json:"orderterms"`
	OrderTermsString      string      `json:"-"`
	OrderElement          string      `json:"orderelement"`
	OrderDirection        string      `json:"orderdirection"`
	OrderElementDefault   string      `json:"-"`
	OrderDirectionDefault string      `json:"-"`

	GlobalFilterTerm string `json:"globalfilterterm"`

	FieldFiltersEnabled bool         `json:"-"`
	FieldFilterTerms    []FilterTerm `json:"fieldfilterterms"`
	FilterSQLString     string       `json:"-"`

	JSON string `json:"-"`
}

//FilterTerm defines the struct for filter (element and term)
type FilterTerm struct {
	Field     string `json:"field,omitempty"`
	FieldType string `json:"fieldtype,omitempty"`
	Term      string `json:"term,omitempty"`
	IsNew     bool   `json:"isnew,omitempty"`
}

//OrderTerm defines the struct for order (element and direction)
type OrderTerm struct {
	Element   string `json:"element,omitempty"`
	Direction string `json:"direction,omitempty"`
}

//PreProcess runs the FieldsProcess and FiltersProcess
func (pd *PageDetails) PreProcess(a interface{}) {
	pd.FieldsProcess(a)
	pd.FiltersProcess()
}

//FieldsProcess retrieves a list of Field objects from a passed struct
func (pd *PageDetails) FieldsProcess(a interface{}) {

	var fields []Field

	e := reflect.ValueOf(a).Elem()

	for i := 0; i < e.NumField(); i++ {
		//Get table tags
		tableTagArray := strings.Split(e.Type().Field(i).Tag.Get("gowebtable"), ",")
		var field Field
		if len(tableTagArray) == 3 {
			//Field Name
			field.Name = tableTagArray[0]
			//Field Header
			field.Header = tableTagArray[1]
			//Field Type
			switch e.Type().Field(i).Type.String() {
			case "int", "uint", "int32", "uint32", "int64", "uint64":
				field.Type = "int"
			default:
				field.Type = e.Type().Field(i).Type.String()
			}
			//Hide
			switch tableTagArray[2] {
			case "true":
				field.Hide = true
			}
		} else {
			field.Header = e.Type().Field(i).Name
		}
		fields = append(fields, field)
	}
	pd.Fields = fields
}

//FiltersProcess processes the global and field filters
func (pd *PageDetails) FiltersProcess() {

	var GlobalFilterArray []string
	var GlobalFilterString string
	var FieldFilterArray []string
	var FieldFilterString string

	//GlobalFilter
	if len(pd.GlobalFilterTerm) > 0 {
		term := strings.ToLower(pd.GlobalFilterTerm)
		var isInt bool
		if _, err := strconv.Atoi(term); err == nil {
			isInt = true
		}
		for _, f := range pd.Fields {
			var sqlString string
			if f.Type == "int" && isInt {
				sqlString = fmt.Sprintf("%s = %s", f.Name, term)
			} else if f.Type == "string" && !isInt {
				sqlString = fmt.Sprintf("LOWER(%s) LIKE '%%%s%%'", f.Name, term)
			}
			if len(sqlString) > 0 {
				GlobalFilterArray = append(GlobalFilterArray, sqlString) //Append to array
			}
		}
		GlobalFilterString = strings.Join(GlobalFilterArray, " OR ") //Create SQL string
	}

	//FieldFilter
	if len(pd.FieldFilterTerms) > 0 {
		for _, ff := range pd.FieldFilterTerms { //Iterate FieldFilterTerms
			var sqlString string
			switch ff.FieldType {
			case "int", "bool":
				sqlString = fmt.Sprintf("%s = %s", ff.Field, ff.Term)
			case "string":
				sqlString = fmt.Sprintf("%s = '%s'", ff.Field, ff.Term)
			}
			FieldFilterArray = append(FieldFilterArray, sqlString) //Append to array
		}
		FieldFilterString = strings.Join(FieldFilterArray, " AND ") //Create SQL string
	}

	//Build string
	if len(GlobalFilterString) > 0 {
		pd.FilterSQLString += "(" + GlobalFilterString + ")"
	}
	if len(GlobalFilterString) > 0 && len(FieldFilterString) > 0 {
		pd.FilterSQLString += " AND "
	}
	if len(FieldFilterString) > 0 {
		pd.FilterSQLString += "(" + FieldFilterString + ")"
	}
	if len(pd.FilterSQLString) > 0 {
		pd.FilterSQLString = " WHERE " + pd.FilterSQLString
	}

}

//PageProcess processes the values used for pageination
func (pd *PageDetails) PageProcess() error {

	//Order Defaults
	if len(pd.OrderTerms) == 0 && pd.OrderElementDefault != "" {

		switch pd.OrderDirectionDefault {
		case "desc":
			pd.OrderDirectionDefault = "desc"
		default:
			pd.OrderDirectionDefault = "asc"
		}

		pd.OrderTerms = append(pd.OrderTerms, OrderTerm{Element: pd.OrderElementDefault, Direction: pd.OrderDirectionDefault})
	}

	//Order Current
	pd.OrderElement = pd.OrderTerms[0].Element
	pd.OrderDirection = pd.OrderTerms[0].Direction

	//Order
	if len(pd.OrderTerms) > 0 {
		var OrderTermsArray []string
		for _, ot := range pd.OrderTerms {
			OrderTermsArray = append(OrderTermsArray, ot.Element+" "+ot.Direction)
		}
		pd.OrderTermsString = " ORDER BY " + strings.Join(OrderTermsArray, ",")
	}

	//Limit Default
	if pd.Limit == 0 {
		pd.Limit = limitDefault
	}

	//Limit Options Default
	if len(pd.LimitOptions) == 0 {
		pd.LimitOptions = limitDefaultOptions
	}

	//Page
	if pd.PageCurrent == 0 {
		pd.PageCurrent = 1
	}

	//Set offset
	pd.Offset = (pd.PageCurrent - 1) * pd.Limit

	//--------------------------------------------------------------------------------------------

	//Record First
	pd.RecordFirst = pd.Offset + 1

	pd.PageCount = int(math.Ceil(float64(pd.RecordsTotal) / float64(pd.Limit)))

	//For no records
	if pd.PageCount == 0 {
		pd.PageCount = 1
		pd.RecordFirst = 0
	}

	//Reset Current Page if goes over total records
	if pd.Offset+pd.Limit > pd.RecordsTotal {
		pd.PageCurrent = pd.PageCount
		pd.Offset = (pd.PageCurrent - 1) * pd.Limit
		pd.RecordFirst = pd.Offset + 1
		//pd.Offset = pd.RecordsTotal - pd.Limit

	}

	//pd.PageCurrent = int(1 + (float64(pd.Offset) / float64(pd.Limit)))

	//Record Last
	pd.RecordLast = pd.Offset + pd.Limit
	if pd.RecordLast > pd.RecordsTotal {
		pd.RecordLast = pd.RecordsTotal
	}

	//Page Previous
	if pd.PageCurrent > 1 {
		pd.PagePrevious = pd.PageCurrent - 1
	}

	//Page Next
	if pd.PageCurrent != pd.PageCount {
		pd.PageNext = pd.PageCurrent + 1
	} else {
		pd.PageNext = pd.PageCount
	}

	//Create PageDetails JSON to include
	pdJSON, err := json.Marshal(pd)
	if err != nil {
		return err
	}
	pd.JSON = string(pdJSON)

	return nil

}

//ResultsProcess converts the struct results into an array
func (pd *PageDetails) ResultsProcess(t interface{}) {

	var results [][]string

	s := reflect.ValueOf(t)

	//Data
	for i := 0; i < s.Len(); i++ {
		v := reflect.ValueOf(s.Index(i).Interface())
		var elementArray []string
		for i := 0; i < v.NumField(); i++ {
			if !pd.Fields[i].Hide {
				elementArray = append(elementArray, fmt.Sprintf("%v", v.Field(i).Interface()))
			}
		}
		results = append(results, elementArray)
	}

	pd.Results = results

}
