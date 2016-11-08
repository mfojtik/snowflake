package generator

import (
	"fmt"

	humanize "github.com/dustin/go-humanize"
	"github.com/mfojtik/snowflakes/pkg/sync"
)

func GenerateHTML(result []*sync.Result) string {
	max := result[0].ReferenceCount
	percentage := func(current int) string {
		if current == 0 {
			return "0"
		}
		return fmt.Sprintf("%d", int32((float32(current)/float32(max))*100))
	}
	out := `
  <html>
  <head>
  <meta http-equiv="refresh" content="30">
  <!-- Latest compiled and minified CSS -->
  <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u" crossorigin="anonymous">
  <!-- Optional theme -->
  <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap-theme.min.css" integrity="sha384-rHyoN1iRsVXV4nD0JutlnGaslCJuC7uwjduW9SVrLvRYooPp2bWYgmgJQIXwl/Sp" crossorigin="anonymous">
  <!-- Latest compiled and minified JavaScript -->
  <script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/js/bootstrap.min.js" integrity="sha384-Tc5IQib027qvyjSMfHjOMaLkfuWVxZxUPnCJA7l2mCWNIpG9mGCD8wGNIcPD7Txa" crossorigin="anonymous"></script>
	<style>
	* { font-size: small; }
	</style>
  </head>
  <body>
	<table class="table table-hover table-striped">
  <tr>
    <th>Pull Number</th>
    <th>Pull Title</th>
    <th width="100">Rate</th>
    <th>Last Occured</th>
    <th>Created</th>
  </tr>`
	for _, r := range result {
		number := fmt.Sprintf("%d", r.Number)
		out += `
		<tr>
		<td><a href="https://github.com/openshift/origin/issues/` + number + `">#` + number + `</a></td>
			<td>` + r.Title + `</td>
			<td>
			<div class="progress">
				<div class="progress-bar" role="progressbar" aria-valuenow="` + percentage(r.ReferenceCount) + `" aria-valuemin="0" aria-valuemax="100" style="width: ` + percentage(r.ReferenceCount) + `%;">
				` + fmt.Sprintf("%d", r.ReferenceCount) + `
				</div>
			</div>
			</td>
			<td>` + humanize.Time(r.LastReferencedAt) + `</td>
			<td>` + humanize.Time(r.CreatedAt) + `</td>
		</tr>
		`
	}
	out += `
	</table>
  <script src="https://ajax.googleapis.com/ajax/libs/jquery/1.12.4/jquery.min.js"></script>
  </body>
  </html>`
	return out
}
