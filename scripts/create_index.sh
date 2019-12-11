curl -XDELETE 'http://localhost:9200/notifilter'
curl -H "Content-Type: application/json" -XPUT 'http://localhost:9200/notifilter' -d '
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0
  },
  "mappings": {
    "numeric_detection": true,
    "date_detection": false,
    "properties": {
      "application": {
        "type": "keyword"
      },
      "name": {
        "type": "keyword"
      },
      "received_at": {
        "type": "date"
      }
    },
    "dynamic_templates": [
      {
        "notanalyzed": {
          "mapping": {
            "type": "keyword"
          },
          "match": "*",
          "match_mapping_type": "string"
        }
      },
      {
        "onlyfloat": {
          "mapping": {
            "type": "double"
          },
          "match": "*",
          "match_mapping_type": "long"
        }
      }
    ]
  }
}
'
