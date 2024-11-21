# Works for Amazon Linux machines

# Usage: ./deployment.install.sh <IP_ADDRESS> <PATH_TO_KEY_FILE>

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
sudo curl -o /usr/local/bin/gomplate -sSL https://github.com/hairyhenderson/gomplate/releases/latest/download/gomplate_linux-amd64
sudo chmod 755 /usr/local/bin/gomplate

sudo yum update -y
sudo yum install git -y

git clone https://github.com/avoronoi/smapp
cd smapp
DEPLOY=1 gomplate -f docker-compose.yml.tmpl -o docker-compose.yml
EOF
