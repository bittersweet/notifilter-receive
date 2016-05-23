curl localhost:8000/v1/preview -d '
{
  "data": { "user_role": "admin", "institute_name": "Markie" },
  "template": "{{ .user_role }} from {{ .institute_name }}"
}
'
