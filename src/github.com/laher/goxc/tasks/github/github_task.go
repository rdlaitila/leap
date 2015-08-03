package github

import "github.com/laher/goxc/tasks"

//runs automatically
func init() {
	tasks.Register(tasks.Task{
		Name:        tasks.TASK_PUBLISH_GITHUB,
		Description: "Upload artifacts to github.com, and generate a local markdown page of links (github project details required in goxc config. See `goxc -h publish-github`)",
		Run:         RunTaskPubGH,
		DefaultSettings: map[string]interface{}{"owner": "", "apikey": "", "repository": "",
			"apihost":       "https://api.github.com",
			"prerelease":    false,
			"body":          "Built by goxc",
			"downloadshost": "https://github.com/",
			"downloadspage": "github.md",
			"fileheader":    "---\nlayout: default\ntitle: Downloads\n---\nFiles hosted at [github.com](https://github.com)\n\n",
			"include":       "*.zip,*.tar.gz,*.deb",
			"exclude":       "github.md,.goxc-temp",
			"outputFormat":  "by-file-extension", // use by-file-extension, markdown or html
			"templateText": `---
layout: default
title: Downloads
---
Files hosted at [github.com](https://github.com)

{{.AppName}} downloads (version {{.Version}})

{{range $k, $v := .Categories}}### {{$k}}

{{range $v}} * [{{.Text}}]({{.RelativeLink}})
{{end}}
{{end}}

{{.ExtraVars.footer}}`,
			"templateFile":      "", //use if populated
			"templateExtraVars": map[string]interface{}{"footer": "Generated by goxc"}}})
}

/*
type GhDownload struct {
	Text         string
	RelativeLink string
}
type GhReport struct {
	AppName    string
	Version    string
	Categories map[string]*[]GhDownload
	ExtraVars  map[string]interface{}
}
*/
