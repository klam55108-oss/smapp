# Usage: ./deployment/deploy.sh <REGISTRY>

REGISTRY=$1 DEPLOY=1 gomplate -f docker-compose.yml.tmpl -o docker-compose.yml
REGISTRY=$1 DEPLOY=1 docker stack deploy --with-registry-auth -c docker-compose.yml -c docker-compose.prod-env.yml smapp