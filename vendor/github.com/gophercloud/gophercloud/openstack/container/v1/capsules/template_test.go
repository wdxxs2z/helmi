package capsules

import (
	"testing"

	th "github.com/gophercloud/gophercloud/testhelper"
)

func TestTemplateParsing(t *testing.T) {
	templateJSON := new(Template)
	templateJSON.Bin = []byte(ValidJSONTemplate)
	err := templateJSON.Parse()
	th.AssertNoErr(t, err)
	th.AssertDeepEquals(t, ValidJSONTemplateParsed, templateJSON.Parsed)

	templateYAML := new(Template)
	templateYAML.Bin = []byte(ValidYAMLTemplate)
	err = templateYAML.Parse()
	th.AssertNoErr(t, err)
	th.AssertDeepEquals(t, ValidYAMLTemplateParsed, templateYAML.Parsed)

	templateInvalid := new(Template)
	templateInvalid.Bin = []byte("Keep Austin Weird")
	err = templateInvalid.Parse()
	if err == nil {
		t.Error("Template parsing did not catch invalid template")
	}
}
