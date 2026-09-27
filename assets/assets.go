// Package assets embeds the app icon and the example build.
package assets

import _ "embed"

//go:generate go run ../cmd/genicon .
//go:generate go tool rsrc -ico icon.ico -arch amd64 -o ../rsrc_windows_amd64.syso

// IconICO is the tray icon (Windows .ico).
//
//go:embed icon.ico
var IconICO []byte

// IconPNG is the 256 px icon.
//
//go:embed icon.png
var IconPNG []byte

// ExampleBuild is written to the builds folder on first run.
//
//go:embed example.yml
var ExampleBuild []byte

// ExampleBuildName is the file name of ExampleBuild.
const ExampleBuildName = "terran-bio-cyclone-medivac.yml"
