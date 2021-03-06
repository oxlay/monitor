/*
This file contains the display and update logic of top-level UI elements.
*/

package client

import (
	"strconv"

	ui "github.com/gizak/termui"
)

// UIPage contains the UI objects used to display the dashboard.
//
// Contrary to UIDashboard, UIPage only contains elements that are actually visible to the user.
type UIPage struct {
	Title   ui.Par // Shows the URL
	Counter ui.Par // Shows the index of the currently displayed website (e.g. 3/8)
	Left    UISide // Stats presented on the left-hand side of the dashboard
	Right   UISide // Stats presented on the right-hand side of the dashboard
	Alerts  ui.Par // Shows the latest alerts
	Footer  ui.Par // Footer with navigation information
}

// NewUIPage initializes the widgets of the dashboard with the
// appropriate UI parameters and returns a new DashboardPage.
func NewUIPage(c *Config) UIPage {
	Title := ui.NewPar("")
	Title.Height = 3

	Counter := ui.NewPar("")
	Counter.Height = 3
	Counter.Border = false

	Alerts := ui.NewPar("")
	Alerts.Height = 15
	Alerts.BorderLabel = "Alerts (aggregated over " + strconv.Itoa(c.Alerts.Timespan) + "s, "
	Alerts.BorderLabel += "refreshed every " + strconv.Itoa(c.Alerts.Frequency) + "s)"

	Footer := ui.NewPar("Use left/right arrows to navigate, or press Q to quit")
	Footer.Height = 3
	Footer.Border = false

	return UIPage{
		*Title,
		*Counter,
		NewUISide(c.Statistics.Left, ui.ColorBlue),
		NewUISide(c.Statistics.Right, ui.ColorYellow),
		*Alerts,
		*Footer,
	}
}

// Refresh updates the UIPage using the latest available data.
func (p *UIPage) Refresh(s *Store) {
	s.RLock()
	defer s.RUnlock()

	// Ensure that the current index is not out of range
	if s.CurrentIdx >= len(s.URLs) {
		return
	}

	url := s.URLs[s.CurrentIdx]

	// Update top-level widgets
	p.Title.Text = url
	p.Counter.Text = "Page " + strconv.Itoa(s.CurrentIdx+1) + "/" + strconv.Itoa(len(s.URLs))
	p.Alerts.Text = FormatAlerts(&s.Alerts, url)

	// Update stats on both sides
	p.Left.Refresh(s.Metrics[url][p.Left.Timespan])
	p.Right.Refresh(s.Metrics[url][p.Right.Timespan])
}

// FormatAlerts converts alerts to a human-readable string,
// to be displayed on the dashboard.
func FormatAlerts(a *Alerts, url string) (str string) {
	for _, alert := range (*a)[url] {
		str += "Website " + url + " is "
		if alert.BelowThreshold {
			str += "down. "
		} else {
			str += "up. "
		}
		str += "availability=" + strconv.FormatFloat(alert.Availability, 'f', 3, 64)
		str += ", time=" + alert.Timeframe.EndDate.String() + "\n"
	}
	return
}
