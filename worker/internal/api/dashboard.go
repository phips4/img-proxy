package api

import (
	"github.com/hashicorp/memberlist"
	"github.com/phips4/img-proxy/worker/internal"
	"html/template"
	"net/http"
)

type dashboardData struct {
	Ip         string
	NodeCount  int
	ImageCount int
}

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
	<title>Image Dashboard</title>
	<style>
		body {
			font-family: Arial, sans-serif;
			background-color: #f4f4f4;
			margin: 20px;
		}
		.container {
			max-width: 600px;
			margin: auto;
			background-color: #fff;
			padding: 20px;
			border-radius: 8px;
			box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
		}
		h1 {
			color: #333;
		}
		.info {
			margin-top: 20px;
		}
	</style>
</head>
<body>
	<div class="container">
		<h1>Image Dashboard</h1>
		<div class="info">
			<p>IP: {{.Ip}}</p>
			<p>Node Count: {{.NodeCount}}</p>
			<p>Image Count: {{.ImageCount}}</p>
		</div>
	</div>
</body>
</html>
`

// dashboardHandler is an HTTP handler function that renders the dashboard template.
func HandleDashboard(cache *internal.Cache, memberlist *memberlist.Memberlist) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := dashboardData{
			Ip:         memberlist.LocalNode().Addr.String(),
			NodeCount:  len(memberlist.Members()),
			ImageCount: cache.Count(),
		}

		tmpl, err := template.New("dashboard").Parse(htmlTemplate)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}
