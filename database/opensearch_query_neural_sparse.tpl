{{ define "opensearch_query_neural_sparse" -}}
{
    "query": {
	    "neural_sparse": { "content_embedding": { "query_text": "{{ .Location.String }}", "model_id": "{{ .ModelId }}" } }
    {{ if HasMetadata .Location.Metadata -}},
    	    "filter": {
		"must": [
		    {{ range $k, $v := .Location.Metadata -}}
	    	    { "term": { "metadata.{{ $k }}": "{{ $k }}" },
	      	    {{ end -}}
		]
	    }    
    {{ end -}}
     }
}    
{{ end -}}
