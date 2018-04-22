package httpRouter

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"
	"context"
	"errors"
)

func CreateHttpRouter() *HttpRouter {
	return &HttpRouter{
		actions: map[string]map[string]HttpAction{},
	}
}

type HttpRouter struct {
	writer  http.ResponseWriter
	actions map[string]map[string]HttpAction
}

func (r *HttpRouter) GET(route string, action ActionFunc, ctx context.Context) {
	r.BindAction(http.MethodGet, route, action, ctx)
}

func (r *HttpRouter) POST(route string, action ActionFunc, ctx context.Context) {
	r.BindAction(http.MethodPost, route, action, ctx)
}

func (r *HttpRouter) ANY(route string, action ActionFunc, ctx context.Context) {
	r.BindAction("*", route, action, ctx)
}

func (r *HttpRouter) BindAction(method, route string, action ActionFunc, ctx context.Context) {
	if _, ok := r.actions[route]; !ok {
		r.actions[route] = map[string]HttpAction{}
	}
	r.actions[route][method] = HttpAction{
		action: action,
		ctx:    ctx,
	}
}

func (r *HttpRouter) Redirect(url string) {
	r.writer.Header().Add("Location", url)
	r.writer.WriteHeader(302)
}

func (r *HttpRouter) GetFormValues(req *http.Request) (res map[string]string) {
	req.ParseForm()
	res = map[string]string{}

	for k := range req.Form {
		res[k] = req.FormValue(k)
	}

	return res
}

func (r *HttpRouter) bindToHttp(route string) {
	http.HandleFunc(route, func(wr http.ResponseWriter, req *http.Request) {
		var action HttpAction
		var ok bool
		var err error
		var cnt *bytes.Buffer

		defer func() {
			if err != nil {
				wr.WriteHeader(500)
				wr.Write([]byte("Error: " + err.Error()))
				return
			}

			if cnt != nil && cnt.Len() > 0 {
				cnt.WriteTo(wr)
			}
		}()

		r.writer = wr

		if action, ok = r.actions[route][req.Method]; !ok {
			if action, ok = r.actions[route]["*"]; !ok {
				err = errors.New("Route not matched: " + route)
				return
			}
		}

		ctx := context.WithValue(action.ctx, "params", r.GetFormValues(req))

		r, err := action.action(ctx)
		if err != nil {
			return
		}

		cnt = bytes.NewBuffer(nil)
		if view, ok := ctx.Value("view").(string); ok {
			view = "views/" + strings.Trim(view, "/")
			if tpl := AppTemplates.GetTemplates().Lookup(view); tpl != nil {
				err = tpl.Execute(cnt, r)
			} else {
				err = errors.New("View not found: " + view)
			}
		}

		if err != nil {
			return
		}

		if layout, ok := ctx.Value("layout").(string); ok {
			layout = "layout/" + layout
			if tpl := AppTemplates.GetTemplates().Lookup(layout); tpl != nil {
				lv, ok := ctx.Value("layoutVars").(map[string]interface{})
				if !ok {
					lv = map[string]interface{}{}
				}
				lv["content"] = template.HTML(cnt.String())

				cnt.Reset()

				err = tpl.Execute(cnt, lv)
			} else {
				err = errors.New("Layout not found: " + layout)
			}
		}
	})
}

func (r *HttpRouter) Bind() {
	for route := range r.actions {
		r.bindToHttp(route)
	}
}
