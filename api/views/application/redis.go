package application

import (
	"github.com/coroot/coroot/api/views/widgets"
	"github.com/coroot/coroot/model"
	"github.com/coroot/coroot/timeseries"
)

func redis(ctx timeseries.Context, app *model.Application) *widgets.Dashboard {
	dash := widgets.NewDashboard(ctx, "Redis")

	for _, i := range app.Instances {
		if i.Redis == nil {
			continue
		}
		status := widgets.NewTableCell().SetStatus(model.OK, "up")
		if !(i.Redis.Up != nil && i.Redis.Up.Last() > 0) {
			status.SetStatus(model.WARNING, "down (no metrics)")
		}
		roleCell := widgets.NewTableCell(i.Redis.Role.Value())
		switch i.Redis.Role.Value() {
		case "master":
			roleCell.SetIcon("mdi-database-edit-outline", "rgba(0,0,0,0.87)")
		case "slave":
			roleCell.SetIcon("mdi-database-import-outline", "grey")
		}

		dash.GetOrCreateTable("Instance", "Role", "Status").AddRow(
			widgets.NewTableCell(i.Name).AddTag("version: %s", i.Redis.Version.Value()),
			roleCell,
			status,
		)

		total := timeseries.Aggregate(timeseries.NanSum)
		calls := timeseries.Aggregate(timeseries.NanSum)
		for cmd, t := range i.Redis.CallsTime {
			if c, ok := i.Redis.Calls[cmd]; ok {
				total.AddInput(t)
				calls.AddInput(c)
			}
		}
		dash.
			GetOrCreateChart("Redis latency, seconds").
			AddSeries(i.Name, timeseries.Aggregate(timeseries.Div, total, calls))
		dash.
			GetOrCreateChartInGroup("Redis queries on <selector>, per seconds", i.Name).
			Stacked().
			Sorted().
			AddMany(timeseries.Top(i.Redis.Calls, timeseries.NanSum, 5))

	}
	return dash
}
