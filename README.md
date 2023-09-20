# imgor
Simple image upload and sharing

Upload with a multipart form with 'image'
Example:
```shell
# Using multi-part form
curl localhost:8080/upload -H "Authorization: Bearer of-arms" -F image=@example.png

# Or directly POSTing the data (-d and --data won't work)
curl localhost:8080/upload -H "Authorization: Bearer of-arms" --data-binary @example.png

```


Running:
```shell
export $(xargs < .env) && go run .
```