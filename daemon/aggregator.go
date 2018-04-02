package agent

import (
	"monitor/payload"
)

func (w *Websites) Alerts(timespan int, threshold float64) payload.Alerts {
	alerts := make(payload.Alerts)
	for i, website := range *w {
		avail := website.Availability(timespan)
		if (avail < threshold) && !website.DownAlertSent {
			// if the website is considered down but no alert for this event was sent yet
			// send a "website is down" alert
			alerts[website.URL] = payload.NewAlert(avail, true)
			(*w)[i].DownAlertSent = true
		} else if (avail >= threshold) && website.DownAlertSent {
			// if the website is considered up but website was last reported down
			// send a "website has recovered" alert
			alerts[website.URL] = payload.NewAlert(avail, false)
			(*w)[i].DownAlertSent = false
		}
	}
	return alerts
}

func (w *Website) Availability(timespan int) float64 {
	// TODO: remove duplicated code with aggregateResults /!\

	// Copy trace results to ensure that they are not modified by
	// concurrent functions while results are being aggregated
	tr := w.TraceResults
	startIdx := tr.StartIndexFor(timespan)
	return tr.Availability(startIdx)
}

func (w *Websites) aggregateResults(timespan int) payload.Stats {
	p := payload.NewStats(timespan)
	for _, website := range *w {
		p.Metrics[website.URL] = website.aggregateResults(timespan)
	}
	return p
}

func (w *Website) aggregateResults(timespan int) payload.Metric {
	// Copy trace results to ensure that they are not modified by
	// concurrent functions while results are being aggregated
	tr := w.TraceResults
	startIdx := tr.StartIndexFor(timespan)
	// TTFBs := tr.TTFBs(startIdx)
	return payload.Metric{
		Availability:     tr.Availability(startIdx),
		Average:          tr.Average(startIdx),
		Max:              tr.Max(startIdx),
		StatusCodeCounts: tr.CountCodes(startIdx),
		ErrorCounts:      tr.CountErrors(startIdx),
	}
}
