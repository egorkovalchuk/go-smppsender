# go-smppsender
SMS Sender (Web service)

Use -l Name list -m "Text message"

Use -d start deamon mode(HTTP service)

Example 1 curl localhost:8080 -X GET -F src=IT -F lst=rss_1 -F text=hello 

Example 2 curl localhost:8080 -X GET -F src=IT -F dst=79XXXXXXXX -F text=hello)

Example 3 curl localhost:8080/conf -X GET -F reloadconf=1

Example 4 curl localhost:8080/list -X GET

Use -s stop deamon mode(HTTP service)

Use -t start with debug mode
