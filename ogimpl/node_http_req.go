package ogimpl

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"text/template"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

var HttpReqNodeFactory = func() ogcore.Node {
	return &HttpReqNode{}
}

type HttpReqNode struct {
	ograph.BaseNode
	*slog.Logger

	Method      string
	Url         string
	ContentType string
	Body        *string
	BodyTpl     *string
}

func (node *HttpReqNode) Run(ctx context.Context, state ogcore.State) error {
	if node.Logger == nil {
		node.Logger = slog.Default()
	}

	switch node.Method {
	case "GET":
		resp, err := http.Get(node.Url)
		if err != nil {
			return err
		} else {
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				node.Error("http request get non 200 response", "NodeName", node.Name(), "Code", resp.StatusCode, "Body", string(body))

				return errors.New("http request get non 200 response")
			}
		}
	case "POST":
		body, err := node.getReqBody(state)
		if err != nil {
			return fmt.Errorf("get request body failed, error: %w", err)
		}

		resp, err := http.Post(node.Url, node.ContentType, body)
		if err != nil {
			return err
		} else {
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				node.Error("http request get non 200 response", "NodeName", node.Name(), "Code", resp.StatusCode, "Body", string(body))

				return errors.New("http request get non 200 response")
			}
		}
	default:
		return errors.New("unsupported request method")
	}

	return nil
}

func (node *HttpReqNode) getReqBody(state ogcore.State) (io.Reader, error) {
	if node.Body != nil {
		return strings.NewReader(*node.Body), nil
	}

	if node.BodyTpl != nil {
		bodyBuf := new(bytes.Buffer)

		funcMap := template.FuncMap{
			"GetState": func(key string) any {
				val, _ := state.Get(key)
				return val
			},
		}

		tpl, err := template.New(node.Name()).Funcs(funcMap).Parse(*node.BodyTpl)
		if err != nil {
			return nil, err
		}

		err = tpl.Execute(bodyBuf, "")
		if err != nil {
			return nil, err
		} else {
			return bodyBuf, nil
		}
	}

	return strings.NewReader(""), nil
}
