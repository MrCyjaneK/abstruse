package badge

import (
	"net/http"
	"strconv"
	"time"

	"github.com/bleenco/abstruse/server/api/render"
	"github.com/bleenco/abstruse/server/core"
	"github.com/go-chi/chi"
	"github.com/narqo/go-badge"
)

// HandleBadge returns an http.HandlerFunc that writes SVG status build
// icon to the http response body.
func HandleBadge(builds core.BuildStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		branch := r.URL.Query().Get("branch")

		reqtype := r.URL.Query().Get("type")

		if reqtype == "" {
			reqtype = "status"
		}
		var name = "unknown"
		var content = "unknown"
		var color = "#e74c3c"

		switch reqtype {
		case "status":
			name = "build"
			var err error
			content, err = builds.FindStatus(token, branch)

			if err != nil {
				content = core.BuildStatusUnknown
			}
			color = "#555555"
			if content == core.BuildStatusPassing {
				color = "#48bb78"
			} else if content == core.BuildStatusFailing {
				color = "#e74c3c"
			} else if content == core.BuildStatusRunning {
				color = "#ecc94b"
			}
		case "onestatus":
			_, build, err := builds.FindRepoBadge(token, branch)
			if err != nil {
				name = "error"
				content = err.Error()
				break
			}
			id, err := strconv.Atoi(r.URL.Query().Get("jobid"))
			if err != nil {
				name = "error"
				content = err.Error()
				break
			}
			if len(build.Jobs) < id {
				name = "error"
				content = "len(build.Jobs) < id"
			}

			name = ""
			content = ""
			status, err := builds.FindStatus(token, branch)
			if r.URL.Query().Get("time") == "true" {
				for bid := range build.Jobs {
					since := time.Since(*build.Jobs[bid].StartTime)
					if status == core.BuildStatusPassing {
						since = time.Now().Sub(*build.Jobs[bid].EndTime)
						color = "#48bb78"
					} else if status == core.BuildStatusRunning {
						color = "#ecc94b"
					} else if status == core.BuildStatusFailing {
						since = time.Now().Sub(*build.Jobs[bid].EndTime)
						color = "#e74c3c"
					}
					content = strconv.Itoa(int(since.Minutes())) + "m"
				}
			}
		}
		svg, err := badge.RenderBytes(name, content, badge.Color(color))
		if err != nil {
			render.InternalServerError(w, err.Error())
			return
		}

		w.Header().Set("Content-Type", "image/svg+xml")
		w.Write(svg)
	}
}
