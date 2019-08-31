package tableFormat

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	MaxColumnLength = 60
	ColumnPadding   = 2
)

/*
Format function formats the given data into table format.
Priority   Sum
#################
Critical   2
High       3
Low        1
It backtracks the input's type using reflect and extracts its headers and contents
To print in table format, we need following details
1) Max width of each column
2) Column headers
3) Column values
*/
func Format(in interface{}) (*string, error) {

	// Initialize
	//    1) Max width of each column
	//    2) Column headers
	//    3) Column values
	m := make(map[string]int)
	var header []string
	var contents [][]string

	err := ExtractContents(in, m, &header, &contents)

	if err != nil {
		return nil, err
	}

	// output format string
	var format string

	format += PrintHeader(m, &header)
	format += PrintContents(m, &header, &contents)

	return &format, nil
}

func PrintContents(m map[string]int, header *[]string, contents *[][]string) string {
	// print contents
	var format string
	// Printf format string
	str := strings.Repeat("%-*s ", len(*header))
	for mn := range *contents {
		row := (*contents)[mn]

		var rowFmt []interface{}
		for n := range *header {
			rowFmt = append(rowFmt, m[(*header)[n]]+ColumnPadding)
			if len(row[n]) > MaxColumnLength {
				rowFmt = append(rowFmt, row[n][:MaxColumnLength])
			} else {
				rowFmt = append(rowFmt, row[n])
			}
		}
		format += fmt.Sprintf(str+"\n", rowFmt...)
	}

	return format
}

func PrintHeader(m map[string]int, header *[]string) string {
	var headerFmt []interface{}
	var format string
	total := 0
	for n := range *header {
		// If length of column is > MaxColumnLength, replace it with MaxColumnLength
		if m[(*header)[n]] > MaxColumnLength {
			m[(*header)[n]] = MaxColumnLength
		}
		total += m[(*header)[n]] + 1 + ColumnPadding
		// append max_width and column string to interface
		headerFmt = append(headerFmt, m[(*header)[n]]+ColumnPadding)
		// if column width is more than MaxColumnLength, truncate it
		if len((*header)[n]) > MaxColumnLength {
			headerFmt = append(headerFmt, (*header)[n][:MaxColumnLength])
		} else {
			headerFmt = append(headerFmt, (*header)[n])
		}
	}
	// Printf format string
	str := strings.Repeat("%-*s ", len(*header))

	// print the heading
	format += fmt.Sprintf(str+"\n", headerFmt...)

	// print ##################
	format += fmt.Sprintf(strings.Repeat("#", total) + "\n")

	return format
}

func ExtractContents(in interface{}, m map[string]int, header *[]string, contents *[][]string) error {

	/*switch t := in.(type) {
	case types.Slice:
		fmt.Printf("incident\n")
	default:
		fmt.Printf("Type is %v\n",t)
	}
	if a, ok := in.(servicenowStore.Incident); ok {
		fmt.Printf("type is slice %v\n",a)
	}*/

	val := reflect.ValueOf(in)

	// Depending on the input type, parse the values
	// At present, it supports slice of struct values
	// this can be extended to support other formats like struct, map etc
	// TODO: create Parser interface{} with format method
	// each format type can implement its own format method
	switch inType := val.Type().Kind().String(); inType {
	case "slice":
		for j := 0; j < val.Len(); j++ {
			v := val.Index(j)
			typeOfS := v.Type()
			switch valType := v.Type().Kind().String(); valType {
			case "struct":
				if j == 0 {
					//fmt.Printf("x is %d y is %d\n",val.Len(),v.NumField() )
					*contents = make([][]string, val.Len())
				}
				for i := 0; i < v.NumField(); i++ {
					// get the value as string
					//fmt.Printf("Type is %v\n",v.Field(i).Type())
					var a string
					switch fieldType := v.Field(i).Type().Kind().String(); fieldType {
					// TODO: Other basic data types can be implemented later
					case "string":
						a = v.Field(i).Interface().(string)
					case "int":
						a = strconv.Itoa(v.Field(i).Interface().(int))
					default:
						return errors.New("unsupported format")
					}

					// add headings into array in the order
					if j == 0 {
						*header = append(*header, typeOfS.Field(i).Name)
					}

					// add contents into two dimensional slice
					(*contents)[j] = append((*contents)[j], a)
					if val, ok := m[typeOfS.Field(i).Name]; ok {
						// find the max length of a column
						if len(a) > val {
							m[typeOfS.Field(i).Name] = len(a)
						}
					} else {
						// include heading length also in column length
						if len(a) > len(typeOfS.Field(i).Name) {
							m[typeOfS.Field(i).Name] = len(a)
						} else {
							m[typeOfS.Field(i).Name] = len(typeOfS.Field(i).Name)
						}
					}
				}
			default:
				return errors.New("unsupported format")
			}
		}
	default:
		return errors.New("unsupported format")
	}
	return nil
}
