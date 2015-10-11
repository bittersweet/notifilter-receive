curl -XDELETE 'http://localhost:9200/notifilter'
curl -XPUT 'http://localhost:9200/notifilter' -d '
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 1
  },
  "mappings": {
    "event": {
      "properties": {
        "app": {
          "type": "string",
          "index": "not_analyzed"
        },
        "name": {
          "type": "string",
          "index": "not_analyzed"
        },
        "received_at": {
          "type": "date"
        }
      },
      "dynamic_templates": [
        {
          "notanalyzed": {
            "mapping": {
              "index": "not_analyzed",
              "type": "string"
            },
            "match": "*",
            "match_mapping_type": "string"
          }
        }
      ]
    }
  }
}
'
