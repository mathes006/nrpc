package main

const tFile = `// This code was autogenerated from {{.GetName}}, do not edit.

{{- $pkgName := GoPackageName .}}
package {{$pkgName}}

import (
	"context"
	"log"
	"time"

	"github.com/golang/protobuf/proto"
	nats "github.com/nats-io/go-nats"
	"github.com/rapidloop/nrpc"
)

{{range .Service -}}
// {{.GetName}}Server is the interface that providers of the service
// {{.GetName}} should implement.
type {{.GetName}}Server interface {
	{{- range .Method}}
	{{.GetName}}(ctx context.Context, req {{GetPkg $pkgName .GetInputType}}) (resp {{GetPkg $pkgName .GetOutputType}}, err error)
	{{- end}}
}

// {{.GetName}}Handler provides a NATS subscription handler that can serve a
// subscription using a given {{.GetName}}Server implementation.
type {{.GetName}}Handler struct {
	ctx    context.Context
	nc     *nats.Conn
	server {{.GetName}}Server
}

func New{{.GetName}}Handler(ctx context.Context, nc *nats.Conn, s {{.GetName}}Server) *{{.GetName}}Handler {
	return &{{.GetName}}Handler{
		ctx:    ctx,
		nc:     nc,
		server: s,
	}
}

func (h *{{.GetName}}Handler) Subject() string {
	return "{{.GetName}}"
}

func (h *{{.GetName}}Handler) Handler(msg *nats.Msg) {
	// decode the request
	name, inner, err := nrpc.Decode(msg.Data)
	if err != nil {
		return
	}

	// call handler and form response
	var resp proto.Message
	var errstr string
	switch name {
	{{- $serviceName := .GetName}}{{- range .Method}}
	case "{{.GetName}}":
		var innerReq {{GetPkg $pkgName .GetInputType}}
		if err := proto.Unmarshal(inner, &innerReq); err != nil {
			log.Printf("{{$serviceName}}Handler: {{.GetName}} request unmarshal failed: %v", err)
			errstr = "bad request received: " + err.Error()
		} else if innerResp, err := h.server.{{.GetName}}(h.ctx, innerReq); err != nil {
			log.Printf("{{$serviceName}}Handler: {{.GetName}} handler failed: %v", err)
			errstr = "handler error: " + err.Error()
		} else {
			resp = &innerResp
		}
	{{end -}}
	default:
		log.Printf("{{$serviceName}}Handler: unknown name %q", name)
		errstr = "unknown name: " + name
	}

	// encode and send response
	nrpc.Publish(resp, errstr, h.nc, msg.Reply) // error is logged
}

type {{.GetName}}Client struct {
	nc      *nats.Conn
	Subject string
	Timeout time.Duration
}

func New{{.GetName}}Client(nc *nats.Conn) *{{.GetName}}Client {
	return &{{.GetName}}Client{
		nc:      nc,
		Subject: "{{.GetName}}",
		Timeout: 5 * time.Second,
	}
}
{{$serviceName := .GetName}}
{{- range .Method}}
func (c *{{$serviceName}}Client) {{.GetName}}(req {{GetPkg $pkgName .GetInputType}}) (resp {{GetPkg $pkgName .GetOutputType}}, err error) {
	// call
	respBytes, err := nrpc.Call("{{.GetName}}", &req, c.nc, c.Subject, c.Timeout)
	if err != nil {
		return // already logged
	}

	// decode inner reponse
	if err = proto.Unmarshal(respBytes, &resp); err != nil {
		log.Printf("{{.GetName}}: response unmarshal failed: %v", err)
		return
	}

	return
}
{{end -}}
{{- end -}}
`
