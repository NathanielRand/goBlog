package controllers

import "coverd/views"

func NewStatic() *Static {
	return &Static{
		AboutView: views.NewView(
			"materialize", "static/about"),
		AdvertiseView: views.NewView(
			"materialize", "static/advertise"),
		AffiliateView: views.NewView(
			"materialize", "static/affiliate"),
		BillingView: views.NewView(
			"materialize", "static/billing"),
		CollectionView: views.NewView(
			"materialize", "static/collection"),
		ContactView: views.NewView(
			"materialize", "static/contact"),
		DashboardView: views.NewView(
			"materialize", "static/dashboard"),
		FaqView: views.NewView(
			"materialize", "static/faq"),
		FaqQuestionView: views.NewView(
			"materialize", "static/faq-question"),
		HelpView: views.NewView(
			"materialize", "static/help"),
		InvestorsView: views.NewView(
			"materialize", "static/investors"),
		LandingView: views.NewView(
			"materialize", "static/landing"),
		NetworkView: views.NewView(
			"materialize", "static/network"),
		PoolView: views.NewView(
			"materialize", "static/pool"),
		PrivacyPolicyView: views.NewView(
			"materialize", "static/privacy-policy"),
		SettingsView: views.NewView(
			"materialize", "static/settings"),
	}
}

type Static struct {
	AboutView         *views.View
	AdvertiseView     *views.View
	AffiliateView     *views.View
	BillingView       *views.View
	CollectionView    *views.View
	ContactView       *views.View
	DashboardView     *views.View
	FaqView           *views.View
	FaqQuestionView   *views.View
	HelpView          *views.View
	InvestorsView     *views.View
	LandingView       *views.View
	NetworkView       *views.View
	PoolView          *views.View
	PrivacyPolicyView *views.View
	SettingsView      *views.View
}
