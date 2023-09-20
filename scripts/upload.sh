#!/bin/bash
for file in ./*.png ./*.jpeg ./*.jpg ./*.gif; do
  [ -e "$file" ] || continue
  curl localhost:8080/upload -H "Authorization: Bearer of-arms" -F image="@$file"
done
