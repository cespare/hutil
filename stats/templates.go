package stats

import (
	"html/template"
)

const (
	html = `
<html>
	<head>
		<title>Server Stats</title>
		<style>
			table {
				border-collapse: collapse;
			}
			th, td {
				padding: 0 5px;
				border: 1px solid #666;
			}
		</style>
	</head>
	<body>

		<section class='stat'>
			<h1>Total requests</h1>
			<table class='counts'>
				<tr>
					{{range .TimePeriods}}
						<th>{{.Name}}</th>
					{{end}}
				</tr>
				<tr>
					{{range .Summary.DiscreteCount.Total}}
						<td>{{.}}</td>
					{{end}}
				</tr>
			</table>
		</section>

		<section class='stat'>
			<h1>Requests by response status code</h1>
			<table class='counts'>
				<tr>
					<th></th>
					{{range .TimePeriods}}
						<th>{{.Name}}</th>
					{{end}}
				</tr>
				{{range $status, $counts := .Summary.DiscreteCount.ResponseStatus}}
					<tr>
						<td>{{$status}}</td>
						{{range $counts}}
							<td>{{.}}</td>
						{{end}}
					</tr>
				{{end}}
			</table>
		</section>

	</body>
</html>
`
)

var (
	page = template.New("foo")
)

func init() {
	template.Must(page.Parse(html))
}
