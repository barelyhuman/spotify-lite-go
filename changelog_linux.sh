curl -s https://api.github.com/repos/barelyhuman/commitlog/releases/latest \
| grep browser_download_url \
| grep linux-amd64 \
| cut -d '"' -f 4 \
| wget -qi -
tar -xvzf commitlog-linux-amd64.tar.gz
chmod +x commitlog 
./commitlog . > CHANGELOG.txt