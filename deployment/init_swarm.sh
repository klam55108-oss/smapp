# Works for Amazon Linux machines

# Executed on remote machine

# Usage: cd smapp && ./deployment/init_swarm.sh

docker swarm init

openssl rand -base64 32  | tr -d '\n' | docker secret create mysql_password -

(
    jwt_private_key=$(openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048)
    echo "$jwt_private_key" | docker secret create jwt_private_key -
    echo "$jwt_private_key" | openssl rsa -pubout | docker config create jwt_public_key -
)
