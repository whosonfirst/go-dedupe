{{ define "opensearch_query" -}}
{
    "query": { "bool": {
	"must": [
	    {{ range $k, $v := .Location.Metadata -}}
	    { "term": { "metadata.{{ $k }}": "{{ $k }}" },
	      {{ end -}}
	    { "neural_sparse": { "name_embedding": { "query_text": "{{ .Location.Name }}", "model_id": "{{ .ModelId }}", "max_token_score": 2.0 } } },
	    { "neural_sparse": { "address_embedding": { "query_text": "{{ .Location.Address }}", "model_id": "{{ .ModelId }}", "max_token_score": 2.0 } } }	      
	]
    }}
}    
{{ end -}}
