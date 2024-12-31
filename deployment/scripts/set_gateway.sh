# Sets current node to be a gateway (a node that runs Traefik instance)

# Usage: ./set_gateway.sh

docker node update --label-add role=gateway $(docker info | grep "NodeID" | awk '{print $2}')
