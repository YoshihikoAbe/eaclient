package eaclient

import (
	"bytes"
	"fmt"
	"net/http"
	"unicode"

	"github.com/YoshihikoAbe/avsproperty"
)

type clientError string

func (err clientError) Error() string {
	return "eaclient: " + string(err)
}

type CompressType int

const (
	CompressDisable CompressType = iota
	CompressNone
	CompressLZ
)

func (c *CompressType) UnmarshalText(b []byte) error {
	switch s := string(bytes.ToLower(b)); s {
	case "":
		fallthrough
	case "disable":
		*c = CompressDisable

	case "none":
		*c = CompressNone

	case "lz":
		fallthrough
	case "lz77":
		*c = CompressLZ

	default:
		return clientError("invalid compress type: " + s)
	}
	return nil
}

type FormatType int

const (
	FormatBinary FormatType = iota
	FormatXML
)

func (f *FormatType) UnmarshalText(b []byte) error {
	switch s := string(bytes.ToLower(b)); s {
	case "":
		fallthrough
	case "binary":
		*f = FormatBinary

	case "xml":
		*f = FormatXML

	default:
		return clientError("invalid format type: " + s)
	}
	return nil
}

type Service struct {
	URL       string       `yaml:"url"`
	Host      string       `yaml:"host"`
	Obfuscate bool         `yaml:"obfuscate"`
	Compress  CompressType `yaml:"compress"`
	Format    FormatType   `yaml:"format"`
	Encoding  string       `yaml:"encoding"`
}

type Client struct {
	Model        string      `yaml:"model"`
	Srcid        string      `yaml:"srcid"`
	UserAgent    string      `yaml:"useragent"`
	DisableQuery bool        `yaml:"disable_query"`
	HTTP         http.Client `yaml:"-"`
}

func (client *Client) Send(svc Service, call *avsproperty.Node) (*avsproperty.Property, error) {
	if !validModel(client.Model) {
		return nil, clientError("invalid character in model")
	}

	if call == nil {
		return nil, clientError("call node is nil")
	}
	if call.Name().String() != "call" {
		return nil, clientError("root node's name is not \"call\"")
	}
	if len(call.Children()) != 1 {
		return nil, clientError("call node has an invalid number of children")
	}

	module := call.Children()[0]
	method := module.AttributeValue("method")
	if method == "" {
		return nil, clientError("module node does not contain a method attribute")
	}

	if !client.DisableQuery {
		// this isn't used for routing on real e-amusement,
		// but it makes the logs look more authentic
		svc.URL += fmt.Sprintf("?model=%s&f=%s.%s", client.Model, module.Name(), method)
	}

	call.SetAttribute("srcid", client.Srcid)
	call.SetAttribute("model", client.Model)

	prop := &avsproperty.Property{
		Root: call,
	}
	err := client.do(prop, svc)
	if err != nil {
		return nil, err
	}
	return prop, nil
}

func (client *Client) do(prop *avsproperty.Property, svc Service) error {
	req, err := EncodeRequest(prop, svc)
	if err != nil {
		return err
	}
	if s := client.UserAgent; s != "" {
		req.Header.Set("User-Agent", s)
	} else {
		req.Header.Set("User-Agent", "EAMUSE.XRPC/1.0")
	}

	resp, err := client.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := DecodeResponse(prop, resp); err != nil {
		return err
	}

	if prop.Root == nil {
		return clientError("empty response property")
	}
	if prop.Root.Name().String() != "response" {
		return clientError("name of root node in response property is not \"response\"")
	}

	return nil
}

func validModel(s string) bool {
	for _, r := range s {
		if !unicode.In(r, unicode.Number, unicode.Letter) && r != ':' {
			return false
		}
	}
	return true
}
