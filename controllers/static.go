package controllers

import "muto/views"

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
		PulseView: views.NewView(
			"materialize", "static/pulse"),
		CollectionView: views.NewView(
			"materialize", "static/collection"),
	}
}

type Static struct {
	LandingView     *views.View
	ContactView     *views.View
	FaqView         *views.View
	FaqQuestionView *views.View
	PulseView       *views.View
	CollectionView  *views.View
}
