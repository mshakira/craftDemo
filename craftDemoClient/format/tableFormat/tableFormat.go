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
It backtracks the input's type using reflect and extracts its headers and contents
To print in table format, we need following details
1) Max width of each column
2) Column headers
3) Column values
 */
func Format(in interface{}) (*string,error) {
	// output format string
	var format string

	// Initialize
	//    1) Max width of each column
	//    2) Column headers
	//    3) Column values
	m := make(map[string]int)
	var header []string
	var contents [][]string

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
					contents = make([][]string, val.Len())
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
						return nil, errors.New("unsupported format")
					}

					// add headings into array in the order
					if j == 0 {
						header = append(header, typeOfS.Field(i).Name)
					}

					// add contents into two dimensional slice
					contents[j] = append(contents[j], a)
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
				return nil, errors.New("unsupported format")
			}
		}
	default:
		return nil, errors.New("unsupported format")
	}

	// interface to hold multiple data types.
	// We need to hold integer and string for Printf format
	// Printf("%*-s",15,str)
	var inter []interface{}
	total := 0
	for n := range header {
		// If length of column is > MaxColumnLength, replace it with MaxColumnLength
		if m[header[n]] > MaxColumnLength {
			m[header[n]] = MaxColumnLength
		}
		total += m[header[n]] + 1 + ColumnPadding
		// append max_width and column string to interface
		inter = append(inter,m[header[n]]+ColumnPadding)
		// if column width is more than MaxColumnLength, truncate it
		if len(header[n]) > MaxColumnLength {
			inter = append(inter, header[n][:MaxColumnLength])
		} else {
			inter = append(inter, header[n])
		}
	}
	// Printf format string
	str := strings.Repeat("%-*s ",len(header))

	// print the heading
	format += fmt.Sprintf(str + "\n",inter...)
	// print ##################
	format += fmt.Sprintf(strings.Repeat("#",total) + "\n")
	// print contents
	for mn:= range contents {
		row := contents[mn]

		var inter []interface{}
		for n := range header {
			inter = append(inter,m[header[n]]+ColumnPadding)
			if len(row[n]) > MaxColumnLength {
				inter = append(inter, row[n][:MaxColumnLength])
			} else {
				inter = append(inter, row[n])
			}
		}
		format += fmt.Sprintf(str + "\n",inter...)
	}

	return &format, nil
}
