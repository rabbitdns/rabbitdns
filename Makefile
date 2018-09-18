cmd/rabbitdns-server/rabbitdns-server:
	cd cmd/rabbitdns-server && dep ensure && go build
cmd/rabbitdns-server/rabbitdns-client:
	cd cmd/rabbitdns-client && dep ensure && go build