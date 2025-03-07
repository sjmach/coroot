package application

import (
	"github.com/coroot/coroot/api/views/widgets"
	"github.com/coroot/coroot/model"
	"github.com/coroot/coroot/timeseries"
	"math"
)

type netSummary struct {
	status   model.Status
	rttMin   *timeseries.AggregatedTimeseries
	rttMax   *timeseries.AggregatedTimeseries
	rttSum   *timeseries.AggregatedTimeseries
	rttCount *timeseries.AggregatedTimeseries
}

func newNetSummary() *netSummary {
	return &netSummary{
		rttMin: timeseries.Aggregate(timeseries.Min),
		rttMax: timeseries.Aggregate(timeseries.Max),
		rttSum: timeseries.Aggregate(timeseries.NanSum),
		rttCount: timeseries.Aggregate(
			func(sum, v float64) float64 {
				if math.IsNaN(sum) {
					sum = 0
				}
				return sum + timeseries.Defined(v)
			},
		),
	}
}

func (s *netSummary) addRtt(rtt timeseries.TimeSeries) {
	s.rttMax.AddInput(rtt)
	s.rttMin.AddInput(rtt)
	s.rttSum.AddInput(rtt)
	s.rttCount.AddInput(rtt)
}

func network(ctx timeseries.Context, app *model.Application, world *model.World) *widgets.Dashboard {
	dash := widgets.NewDashboard(ctx, "Network")
	upstreams := map[model.ApplicationId]*netSummary{}

	for _, instance := range app.Instances {
		for _, u := range instance.Upstreams {
			if u.RemoteInstance == nil {
				continue
			}
			upstreamApp := world.GetApplication(u.RemoteInstance.OwnerId)
			if upstreamApp == nil {
				continue
			}
			summary := upstreams[upstreamApp.Id]
			if summary == nil {
				summary = newNetSummary()
				upstreams[upstreamApp.Id] = summary
			}
			linkStatus := u.Status()
			if linkStatus > summary.status {
				summary.status = linkStatus
			}
			if u.Rtt != nil {
				summary.addRtt(u.Rtt)
			}
			instanceObsolete := instance.Pod != nil && instance.Pod.IsObsolete()
			if instanceObsolete || u.Obsolete() {
				linkStatus = model.UNKNOWN
			}
			if instance.Node != nil && u.RemoteInstance.Node != nil {
				sn := instance.Node
				dn := u.RemoteInstance.Node
				dash.GetOrCreateDependencyMap().UpdateLink(
					widgets.DependencyMapInstance{Name: instance.Name, Obsolete: instanceObsolete},
					widgets.DependencyMapNode{Name: sn.Name.Value(), Provider: sn.CloudProvider.Value(), Region: sn.Region.Value(), AZ: sn.AvailabilityZone.Value()},
					widgets.DependencyMapInstance{Name: u.RemoteInstance.Name, Obsolete: u.Obsolete()},
					widgets.DependencyMapNode{Name: dn.Name.Value(), Provider: dn.CloudProvider.Value(), Region: dn.Region.Value(), AZ: dn.AvailabilityZone.Value()},
					linkStatus,
				)
			}
		}
	}
	for appId, summary := range upstreams {
		dash.GetOrCreateChartInGroup("Network round-trip time to <selector>, seconds", appId.Name).
			AddSeries("min", summary.rttMin).
			AddSeries("avg", timeseries.Aggregate(timeseries.Div, summary.rttSum, summary.rttCount)).
			AddSeries("max", summary.rttMax)
	}
	return dash
}
