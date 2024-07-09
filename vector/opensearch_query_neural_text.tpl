{{ define "opensearch_query_neural_text" -}}
{
    "_source": { "excludes": [ "embeddings" ] },	
    "query": {
    	     "bool": {
	    "must": [
	    {{ if HasMetadata .Location.Metadata -}}
		    {{ range $k, $v := .Location.Metadata -}}
	    	    { "term": { "metadata.{{ $k }}": "{{ $v }}" } },
	      	    {{ end -}}
	    {{ end -}}	    
		{ "neural": { "name_embeddings": { "query_text": "{{ .Location.Name }}", "model_id": "{{ .ModelId }}", "max_distance": 5.0 } } },
		{ "neural": { "address_embeddings": { "query_text": "{{ .Location.Address }}", "model_id": "{{ .ModelId }}", "max_distance": 5.0 } } } 
	    ]
	    }

     }
}

{{ end -}}
