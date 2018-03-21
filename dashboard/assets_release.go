//+build release
//go:generate yarn install
//go:generate yarn build
//go:generate go run assets_generate.go

package dashboard

func init() {
	Assets = generatedAssets
}
