package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
)

func main() {
	tmpl := `
<script>
	const rawData = ` + "`{{ printf \"%s\" .Data }}`" + `;
	console.log(rawData);
</script>
`
	t := template.Must(template.New("test").Parse(tmpl))
	
	data := []byte(`[{"day": "Thứ 2"}]`)
	
	var buf bytes.Buffer
	err := t.Execute(&buf, map[string]interface{}{"Data": data})
	if err != nil {
		panic(err)
	}
	fmt.Println(buf.String())
}
