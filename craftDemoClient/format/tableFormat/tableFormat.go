package tableFormat

import (
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
func Format(in interface{}) string {
	// output format string
	var format string

	//fmt.Printf("value is %v\n",val.Type().Kind())
	// Initialize
	//    1) Max width of each column
	//    2) Column headers
	//    3) Column values
	m := make(map[string]int)
	var header []string
	var contents [][]string

	val := reflect.ValueOf(in)

	for j := 0;j< val.Len(); j++ {
		v := val.Index(j)
		typeOfS := v.Type()
		if j == 0 {
			//fmt.Printf("x is %d y is %d\n",val.Len(),v.NumField() )
			contents = make([][]string,val.Len())
		}
		for i := 0; i < v.NumField(); i++ {
			// get the value as string
			//fmt.Printf("Type is %v\n",v.Field(i).Type())
			var a string
			switch valType := v.Field(i).Type().Kind().String(); valType {
			case "string":
				a = v.Field(i).Interface().(string)
			case "int":
				a = strconv.Itoa(v.Field(i).Interface().(int))
			}

			//a := v.Field(i).Interface().(string)
			// add headings into array in the order
			if j == 0 {
				header = append(header,typeOfS.Field(i).Name)
			}

			contents[j] = append(contents[j],a)
			if val, ok := m[typeOfS.Field(i).Name]; ok {
				//fmt.Printf("%v %v\n",val,len(a))
				if len(a) > val {
					m[typeOfS.Field(i).Name] = len(a)
				}
			} else {
				if len(a) > len(typeOfS.Field(i).Name) {
					m[typeOfS.Field(i).Name] = len(a)
				} else {
					m[typeOfS.Field(i).Name] = len(typeOfS.Field(i).Name)
				}
			}
		}
	}

	// foreach
	var inter []interface{}
	total := 0
	for n := range header {
		// If length of column is > MAX_COLUMN_LENGTH, replace it with MAX_COLUMN_LENGTH
		if m[header[n]] > MaxColumnLength {
			m[header[n]] = MaxColumnLength
		}
		total += m[header[n]] + 1 + ColumnPadding
		//fmt.Printf("%v\n",header[n])
		inter = append(inter,m[header[n]]+ColumnPadding)
		if len(header[n]) > MaxColumnLength {
			inter = append(inter, header[n][:MaxColumnLength])
		} else {
			inter = append(inter, header[n])
		}
	}
	str := strings.Repeat("%-*s ",len(header))

	format += fmt.Sprintf(str + "\n",inter...)
	format += fmt.Sprintf(strings.Repeat("#",total) + "\n")
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

	return format
}
