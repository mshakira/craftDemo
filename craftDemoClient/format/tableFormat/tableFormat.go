package tableFormat

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	MAX_COLUMN_LENGTH = 60
	COLUMN_PADDING = 2
)

func Format(in interface{}) {

	val := reflect.ValueOf(in)

	//fmt.Printf("value is %v\n",val.Type().Kind())
	// to calculate the max length of each column
	m := make(map[string]int)

	var header []string
	var contents [][]string

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
		if m[header[n]] > MAX_COLUMN_LENGTH {
			m[header[n]] = MAX_COLUMN_LENGTH
		}
		total += m[header[n]] + 1 + COLUMN_PADDING
		//fmt.Printf("%v\n",header[n])
		inter = append(inter,m[header[n]]+COLUMN_PADDING)
		if len(header[n]) > MAX_COLUMN_LENGTH {
			inter = append(inter, header[n][:MAX_COLUMN_LENGTH])
		} else {
			inter = append(inter, header[n])
		}
	}
	str := strings.Repeat("%-*s ",len(header))

	fmt.Printf(str + "\n",inter...)
	fmt.Printf(strings.Repeat("#",total) + "\n")
	for mn:= range contents {
		row := contents[mn]

		var inter []interface{}
		for n := range header {
			inter = append(inter,m[header[n]]+COLUMN_PADDING)
			if len(row[n]) > MAX_COLUMN_LENGTH {
				inter = append(inter, row[n][:MAX_COLUMN_LENGTH])
			} else {
				inter = append(inter, row[n])
			}
		}
		fmt.Printf(str + "\n",inter...)
	}
}
