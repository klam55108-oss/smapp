sudo yum update -y
sudo yum install -y docker
sudo service docker start
sudo usermod -a -G docker ec2-user
# Start a subshell with docker group
sg docker << EOF
# Install gomplate
sudo curl -o /usr/bin/gomplate -sSL https://github.com/hairyhenderson/gomplate/releases/latest/download/gomplate_linux-amd64
sudo chmod 755 /usr/bin/gomplate
cd smapp
DEPLOY=1 gomplate -f docker-compose.yml.tmpl -o docker-compose.yml
sudo curl -Lo /usr/bin/docker-credential-ecr-login https://amazon-ecr-credential-helper-releases.s3.us-east-2.amazonaws.com/0.9.0/linux-amd64/docker-credential-ecr-login
sudo chmod +x /usr/bin/docker-credential-ecr-login
rm -rf ~/.docker
mkdir ~/.docker && echo '{
 "credsStore": "ecr-login"
}' > ~/.docker/config.json
EOF