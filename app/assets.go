package app

import "github.com/ichiban/assets"

func NewAssets() (*assets.Locator, error) {
	return assets.New()
}
