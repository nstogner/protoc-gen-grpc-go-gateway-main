package main

import (
	"bytes"
	"go/format"
	"io"
	"io/ioutil"
	"log"
	"os"
	"text/template"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

func main() {
	encodeResponse(
		generateResponse(
			parseRequest(
				decodeRequest(os.Stdin),
			),
		),
		os.Stdout,
	)
}

// decodeRequest unmarshals the protobuf request.
func decodeRequest(r io.Reader) *plugin.CodeGeneratorRequest {
	var req plugin.CodeGeneratorRequest
	input, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal("unable to read stdin: " + err.Error())
	}
	if err := proto.Unmarshal(input, &req); err != nil {
		log.Fatal("unable to marshal stdin as protobuf: " + err.Error())
	}
	return &req
}

// parseRequest wrangles the request to fit needs of the template.
func parseRequest(req *plugin.CodeGeneratorRequest) *params {
	for _, pf := range req.GetProtoFile() {
		for _, svc := range pf.GetService() {
			return &params{
				ServiceDescriptorProto: *svc,
				PackageName:            pf.GetPackage(),
				ProtoName:              pf.GetName(),
			}
		}

	}
	return nil
}

// generateResponse executes the template.
func generateResponse(p *params) *plugin.CodeGeneratorResponse {
	if p == nil {
		return nil
	}

	var resp plugin.CodeGeneratorResponse

	w := &bytes.Buffer{}
	if err := tmpl.Execute(w, p); err != nil {
		log.Fatal("unable to execute template: " + err.Error())
	}

	fmted, err := format.Source([]byte(w.String()))
	if err != nil {
		log.Fatal("unable to go-fmt output: " + err.Error())
	}

	fileName := "main.go"
	fileContent := string(fmted)
	resp.File = append(resp.File, &plugin.CodeGeneratorResponse_File{
		Name:    &fileName,
		Content: &fileContent,
	})

	return &resp
}

// encodeResponse marshals the protobuf response.
func encodeResponse(resp *plugin.CodeGeneratorResponse, w io.Writer) {
	if resp == nil {
		return
	}

	outBytes, err := proto.Marshal(resp)
	if err != nil {
		log.Fatal("unable to marshal response to protobuf: " + err.Error())
	}

	if _, err := w.Write(outBytes); err != nil {
		log.Fatal("unable to write protobuf to stdout: " + err.Error())
	}
}

// params is the data provided to the template.
type params struct {
	descriptor.ServiceDescriptorProto
	ProtoName   string
	PackageName string
	fileName    string
}

var tmpl = template.Must(template.New("server").Parse(`
// Code initially generated by protoc-gen-grpc-go-http-main
// source: {{.ProtoName}}

package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

func main() {
	port := "8080"
	if p, ok := os.LookupEnv("PORT"); ok {
		port = p
	}

	tgtAddr := "localhost:50051"
	if ta, ok := os.LookupEnv("TARGET_ADDR"); ok {
		tgtAddr = ta
	}

	log.Fatal(listen(port, tgtAddr))
}

func listen(port, tgtAddr string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := {{.PackageName}}.Register{{.ServiceDescriptorProto.Name}}HandlerFromEndpoint(ctx, mux, tgtAddr, opts)
	if err != nil {
		return err
	}

	return http.ListenAndServe(":"+port, mux)
}
`))
