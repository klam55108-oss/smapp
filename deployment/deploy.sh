# Usage: ./deployment/deploy <REGISTRY>

# Note: IAM role has to be attached to the instance for this to work.

REGISTRY=$1 docker stack deploy --with-registry-auth -c docker-compose.yml -c docker-compose.prod-env.yml smapp
