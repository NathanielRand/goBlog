package controllers

import "GoBlog/views"

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
	}
}

type Static struct {
	LandingView     *views.View
	ContactView     *views.View
	FaqView         *views.View
	FaqQuestionView *views.View
	CollectionView  *views.View
	DashboardView   *views.View
}
