## [{{ .Current }}](https://github.com/FerretDB/FerretDB/releases/tag/{{ .Current }}) ({{ .Date }})
{{- $root := . }}
{{- range .Categories }}
{{ $prs := index $root.PRs . }}
{{- if $prs }}
### {{ . }}
{{ range $prs }}
- {{ .Title }} by @{{ .User }} in {{ .URL }}
{{- end }}
{{- end }}
{{- end }}
[All closed issues and pull requests]({{ .URL }}?closed=1).
{{- if .Previous }}
[All commits](https://github.com/FerretDB/FerretDB/compare/{{ .Previous }}...{{ .Current }}).
{{- end }}
