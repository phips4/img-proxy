package api

import (
	"github.com/hashicorp/memberlist"
	"github.com/phips4/img-proxy/worker/internal"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type dashboardData struct {
	Ip           string
	GatewayCount int
	WorkerCount  int
	NodeCount    int
	ImageCount   int
	Meta         string
	Name         string
}

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
	<title>Worker Dashboard</title>
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
		<h1>Worker Dashboard</h1>
		<div class="info">
			<p>Name: {{.Name}}</p>
			<p>Meta: {{.Meta}}</p>
			<p>IP: {{.Ip}}</p>
			<p>Worker Node Count: {{.WorkerCount}}</p>
			<p>Gateway Node Count: {{.GatewayCount}}</p>
			<p>total: {{.NodeCount}}</p>
		</div>
	</div>
</body>
</html>
`

// HandleDashboard is an HTTP handler function that renders the dashboard template.
func HandleDashboard(cache *internal.Cache, ml *memberlist.Memberlist) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := dashboardData{
			Ip:         ml.LocalNode().Addr.String(),
			NodeCount:  len(ml.Members()),
			ImageCount: cache.Count(),
			Name:       ml.LocalNode().Name,
			Meta:       string(ml.LocalNode().Meta),
		}

		var gateways []*memberlist.Node
		for _, m := range ml.Members() {
			meta := string(m.Meta)
			if strings.Contains(meta, "gateway") {
				gateways = append(gateways, m)
			}
		}

		data.GatewayCount = len(gateways)
		var workers []*memberlist.Node

		for _, m := range ml.Members() {
			meta := string(m.Meta)
			if strings.Contains(meta, "worker") {
				workers = append(workers, m)
			}
		}
		data.WorkerCount = len(workers)

		tmpl, err := template.New("dashboard").Parse(htmlTemplate)
		if err != nil {
			log.Println("DashboardHandler (worker) error parsing template:", err)
			http.Error(w, internalErrorStr, http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			log.Println("DashboardHandler (worker) error executing template:", err)
			http.Error(w, internalErrorStr, http.StatusInternalServerError)
			return
		}
	}
}
