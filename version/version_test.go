package version

func ExamplePrintln() {
	tests := []struct {
		component string
		version   string
		edition   string
		buildDate string
		buildSHA  string
		want      string
	}{
		{
			component: "testing",
			version:   "1.2.3",
		},
		{
			component: "testing",
			version:   "1.2.3",
			buildSHA:  "387c20615518f1325528705e0ef09e4d30d80378",
		},
		{
			component: "testing",
			version:   "1.2.3",
			buildDate: "2019-10-09",
			buildSHA:  "387c20615518f1325528705e0ef09e4d30d80378",
		},
		{
			component: "testing",
			version:   "1.2.3",
			edition:   "commercial",
		},
	}
	for _, tt := range tests {
		Version = tt.version
		if tt.edition != "" {
			Edition = tt.edition
		}
		BuildDate = tt.buildDate
		BuildSHA = tt.buildSHA
		Println(tt.component)
	}
	// Output:
	// testing version 1.2.3, opensource edition
	// testing version 1.2.3, opensource edition, build 387c20615518f1325528705e0ef09e4d30d80378
	// testing version 1.2.3, opensource edition, build 387c20615518f1325528705e0ef09e4d30d80378, built 2019-10-09
	// testing version 1.2.3, commercial edition
}
