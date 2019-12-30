package gen

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/golang/glog"
	gogen "github.com/golang/protobuf/protoc-gen-go/generator"
	"github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/descriptor"
)

// This function is called with a param which contains the entire definition of a method.
func applyTemplate(f *descriptor.File, opts Options) (string, error) {
	w := bytes.NewBuffer(nil)

	if err := headerTemplate.Execute(w, tplHeader{
		File: f,
		AdditionalImports: opts.AdditionalImports,
	}); err != nil {
		return "", err
	}

	for _, msg := range f.Messages {
		glog.V(2).Infof("Processing %s", msg.GetName())
		if msg.Options != nil && msg.Options.GetMapEntry() {
			glog.V(2).Infof("Skipping %s, mapentry message", msg.GetName())
			continue
		}
		msgName := gogen.CamelCase(*msg.Name)
		msg.Name = &msgName
		if err := messageTemplate.Execute(w, tplMessage{
			Message: msg,
			Options: opts,

		}); err != nil {
			return "", err
		}
	}

	return w.String(), nil
}

type tplHeader struct {
	*descriptor.File
	AdditionalImports []string
}

type tplMessage struct {
	*descriptor.Message
	Options
}

// TypeName returns the name of the type for this message. This logic
// is based on the logic of Descriptor.TypeName in golang/protobuf.
func (t tplMessage) TypeName() string {
	if len(t.Outers) > 0 {
		return strings.Join(t.Outers, "_") + "_" + *t.Name
	}

	return *t.Name
}

var (
	headerTemplate = template.Must(template.New("header").Parse(`
// Code generated by protoc-gen-go-json. DO NOT EDIT.
// source: {{.GetName}}

package {{.GoPkg.Name}}

{{- if not (eq (len .Messages) 0)}}

import (
	"bytes"

	"github.com/golang/protobuf/jsonpb"
	{{range $i := .AdditionalImports}}
	"${i}"
	{{- end}}
)

{{- end -}}
`))

	messageTemplate = template.Must(template.New("message").Parse(`
// MarshalJSON implements json.Marshaler
func (msg *{{.TypeName}}) MarshalJSON() ([]byte,error) {
	var buf bytes.Buffer
	err := (&jsonpb.Marshaler{
	  EnumsAsInts: {{.EnumsAsInts}},
	  EmitDefaults: {{.EmitDefaults}},
	  OrigName: {{.OrigName}},
	}).Marshal(&buf, msg)
	return buf.Bytes(), err
}

// UnmarshalJSON implements json.Unmarshaler
func (msg *{{.TypeName}}) UnmarshalJSON(b []byte) error {
	return {{.Unmarshaler}}(bytes.NewReader(b), msg)
}
`))
)
