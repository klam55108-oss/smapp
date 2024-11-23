# Works for Amazon Linux machines

# Executed on remote machine

# Usage: cd smapp && source ./deployment/init_swarm.sh

docker swarm init

openssl rand -base64 32  | tr -d '\n' | docker secret create mysql_password -

temp_key_file=$(mktemp)
chmod 600 $temp_key_file
openssl genpkey -algorithm RSA -out $temp_key_file -pkeyopt rsa_keygen_bits:2048
docker secret create jwt_private_key $temp_key_file
export JWT_PUBLIC_KEY=$(echo $jwt_private_key | openssl rsa -pubout -in $temp_key_file)
shred -u $temp_key_file || rm $temp_key_file
