cd ~/portfolio/Portfolio
# pull/upload updated code
git pull
cd go-backend && go build -o portfolio-api && sudo systemctl restart portfolio-api
cd ../frontend && npm ci && npm run build && sudo rsync -a --delete dist/ /var/www/portfolio/
sudo systemctl reload caddy


#GOMEMLIMIT=250MiB GOGC=20 go build -p 1 -gcflags="all=-N -l" -ldflags="-s -w" -o portfolio-api ./...