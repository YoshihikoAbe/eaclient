package eaclient

import (
	"bytes"
	"crypto/cipher"
	"io"
	"net/http"

	"github.com/YoshihikoAbe/avslz"
	"github.com/YoshihikoAbe/avsproperty"
)

const (
	infoHeader     = "X-Eamuse-Info"
	compressHeader = "X-Compress"
)

func EncodeRequest(prop *avsproperty.Property, svc Service) (*http.Request, error) {
	body := bytes.NewBuffer(nil)
	wr := io.Writer(body)

	request, err := http.NewRequest("POST", svc.URL, nil)
	if err != nil {
		return nil, err
	}
	request.Host = svc.Host

	if svc.Obfuscate {
		info := eamuseInfo{}
		info.fill()
		wr = cipher.StreamWriter{
			W: wr,
			S: info.makeCipher(),
		}
		request.Header.Set(infoHeader, info.String())
	}

	var lz *avslz.Writer
	if svc.Compress == CompressLZ {
		request.Header.Set(compressHeader, "lz77")
		lz = avslz.NewWriter(wr)
		wr = lz
	} else if svc.Compress == CompressNone {
		request.Header.Set(compressHeader, "none")
	}

	if svc.Format == FormatBinary {
		prop.Settings.Format = avsproperty.FormatBinary
	} else {
		prop.Settings.Format = avsproperty.FormatXML
	}
	encoding := avsproperty.EncodingByName(svc.Encoding)
	if encoding == nil {
		return nil, clientError("invalid encoding: " + svc.Encoding)
	}
	prop.Settings.Encoding = encoding
	if err := prop.Write(wr); err != nil {
		return nil, err
	}

	if lz != nil {
		if err := lz.Close(); err != nil {
			return nil, err
		}
	}

	request.Body = io.NopCloser(body)
	request.ContentLength = int64(body.Len())

	return request, nil
}

func DecodeResponse(prop *avsproperty.Property, resp *http.Response) error {
	if resp.StatusCode != 200 {
		return clientError("invalid HTTP status: " + resp.Status)
	}
	rd := io.Reader(resp.Body)

	if s := resp.Header.Get(infoHeader); s != "" {
		info := eamuseInfo{}
		if err := info.parse(s); err != nil {
			return err
		}
		rd = cipher.StreamReader{
			R: rd,
			S: info.makeCipher(),
		}
	}

	if s := resp.Header.Get(compressHeader); s == "lz77" {
		rd = avslz.NewReader(rd)
	} else if s != "" && s != "none" {
		return clientError("invalid compress type in response: " + s)
	}

	return prop.Read(rd)
}
