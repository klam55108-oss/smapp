# Works for Amazon Linux machines

# Usage: ./deployment/install.sh <IP_ADDRESS> <PATH_TO_KEY_FILE>

ip=$1
key_file=$2

ssh -i $key_file ec2-user@$ip << EOF
sudo yum update -y
sudo yum install -y docker
sudo service docker start
sudo usermod -a -G docker ec2-user
EOF

ssh -i $key_file ec2-user@$ip << EOF
# Install gomplate
sudo curl -o /usr/bin/gomplate -sSL https://github.com/hairyhenderson/gomplate/releases/latest/download/gomplate_linux-amd64
sudo chmod 755 /usr/bin/gomplate

sudo yum update -y
sudo yum install git -y

git clone https://github.com/avoronoi/smapp
cd smapp
DEPLOY=1 gomplate -f docker-compose.yml.tmpl -o docker-compose.yml
chmod +x deployment/init_swarm.sh
chmod +x deployment/deploy.sh

sudo curl -Lo /usr/bin/docker-credential-ecr-login https://amazon-ecr-credential-helper-releases.s3.us-east-2.amazonaws.com/0.9.0/linux-amd64/docker-credential-ecr-login
sudo chmod +x /usr/bin/docker-credential-ecr-login
mkdir ~/.docker && echo '{
  "credsStore": "ecr-login"
}' > ~/.docker/config.json
EOF
