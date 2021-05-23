watch:
	reflex -r \.go$\ -s -- sh -c 'clear && APP_ENV=dev go run -race cmd/webserver/main.go'