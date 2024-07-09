{{ define "opensearch_query_simple" -}}{
    "query": { "simple_query_string": { "query": "{{ .Location.String }}", "fields": [ "content" ], "default_operator": "AND" } }
}{{ end -}}
