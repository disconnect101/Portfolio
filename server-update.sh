cd ~/portfolio/Portfolio
# pull/upload updated code
git pull
cd backend && mvn clean package -DskipTests && sudo systemctl restart portfolio-api
cd ../frontend && npm ci && npm run build && sudo rsync -a --delete dist/ /var/www/portfolio/
sudo systemctl reload caddy
