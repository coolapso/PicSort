package ui

import (
	_ "embed"
	"fyne.io/fyne/v2"
)

//go:embed assets/discord.svg
var discordLogoBytes []byte
var DiscordIcon = &fyne.StaticResource{StaticName: "discord.svg", StaticContent: discordLogoBytes}

//go:embed assets/coffee.svg
var coffeeLogoBytes []byte
var CoffeeIcon = &fyne.StaticResource{StaticName: "coffee.svg", StaticContent: coffeeLogoBytes}

//go:embed assets/github.svg
var githubLogoBytes []byte
var GithubIcon = &fyne.StaticResource{StaticName: "github.svg", StaticContent: githubLogoBytes}

//go:embed assets/mastodon.svg
var mastodonLogoBytes []byte
var MastodonIcon = &fyne.StaticResource{StaticName: "mastodon.svg", StaticContent: mastodonLogoBytes}

//go:embed assets/sponsor.svg
var sponsorLogoBytes []byte
var SponsorIcon = &fyne.StaticResource{StaticName: "sponsor.svg", StaticContent: sponsorLogoBytes}

//go:embed assets/x.svg
var xLogoBytes []byte
var XIcon = &fyne.StaticResource{StaticName: "x.svg", StaticContent: xLogoBytes}

var Icons = map[string]*fyne.StaticResource{
	"discord":  DiscordIcon,
	"bmc":      CoffeeIcon,
	"github":   GithubIcon,
	"mastodon": MastodonIcon,
	"sponsor":  SponsorIcon,
	"x":        XIcon,
}
