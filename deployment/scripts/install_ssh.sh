# Works for Amazon Linux machines
# Usage: ./deployment/install.sh <IP_ADDRESS> <PATH_TO_KEY_FILE>
ip=$1
ssh ec2-user@$ip << EOF
sudo yum update -y
sudo yum install git -y
rm -rf smapp
git clone https://github.com/avoronoi/smapp
./smapp/deployment/scripts/install_local.sh
EOF