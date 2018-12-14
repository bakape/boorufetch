package boorufetch

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/olekukonko/tablewriter"
)

func logPosts(t *testing.T, err error, posts ...Post) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
	if len(posts) == 0 {
		t.Fatal("no images found")
	}

	typ := reflect.TypeOf((*Post)(nil)).Elem()
	var buf bytes.Buffer
	for i, p := range posts {
		if i == 10 {
			break
		}

		buf.Reset()
		w := tablewriter.NewWriter(&buf)
		w.SetAlignment(tablewriter.ALIGN_LEFT)
		w.SetColWidth(80)
		w.SetRowLine(true)

		val := reflect.ValueOf(p)
		for i := 0; i < typ.NumMethod(); i++ {
			name := typ.Method(i).Name
			re := val.MethodByName(name).Call(nil)
			if len(re) == 2 {
				err := re[1].Interface()
				if err != nil {
					t.Fatal(err)
				}
			}
			w.Append([]string{name, fmt.Sprint(re[0])})
		}

		w.Render()
		t.Logf("\n%s\n", buf.String())
	}
}
