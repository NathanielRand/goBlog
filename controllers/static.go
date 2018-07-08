package controllers

import "coverd/views"

func NewStatic() *Static {
	return &Static{
		LandingView: views.NewView(
			"materialize", "static/landing"),
		ContactView: views.NewView(
			"materialize", "static/contact"),
		FaqView: views.NewView(
			"materialize", "static/faq"),
		FaqQuestionView: views.NewView(
			"materialize", "static/faq-question"),
		CollectionView: views.NewView(
			"materialize", "static/collection"),
		DashboardView: views.NewView(
			"materialize", "static/dashboard"),
		NetworkView: views.NewView(
			"materialize", "static/network"),
		BillingView: views.NewView(
			"materialize", "static/billing"),
		SettingsView: views.NewView(
			"materialize", "static/settings"),
	}
}

type Static struct {
	LandingView     *views.View
	ContactView     *views.View
	FaqView         *views.View
	FaqQuestionView *views.View
	CollectionView  *views.View
	DashboardView   *views.View
	NetworkView     *views.View
	BillingView     *views.View
	SettingsView    *views.View
}
