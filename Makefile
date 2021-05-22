watch:
	reflex -r \.go$\ -s -- sh -c 'clear && APP_ENV=dev go run cmd/webserver/main.go'