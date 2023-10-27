# Personal Dashboard


## How to use:
1. Install Golang, Python and SQLite3.
2. Install Firefox (from APT, not SNAP):
```sh
sudo add-apt-repository ppa:mozillateam/ppa
echo '
Package: *
Pin: release o=LP-PPA-mozillateam
Pin-Priority: 1001
' | sudo tee /etc/apt/preferences.d/mozilla-firefox

sudo apt update
sudo apt install xdg-utils firefox -y
```
3. Install Geckodriver:
```sh
curl -L -o /tmp/geckodriver.tar.gz https://github.com/mozilla/geckodriver/releases/download/v0.33.0/geckodriver-v0.33.0-linux64.tar.gz
mkdir -p /opt/geckodriver/
tar -xvf /tmp/geckodriver.tar.gz -C /opt/geckodriver/
```
4. Install [Nginx](https://www.nginx.com). Nginx will act as a **reverse proxy** for the Streamlit dashboard app.
5. Create the file **configs/configs.json** with some configs and credentials. This file should follow the structure of the **configs/configs.example.json** file.
6. The dashboard uses the [Streamlit Authenticator](https://github.com/mkhorasani/Streamlit-Authenticator/tree/main) module, check [here](https://github.com/mkhorasani/Streamlit-Authenticator/tree/main#1-hashing-passwords) how to create the file **.streamlit/credentials/credentials.yaml** (should be at this location!) with the users/passwords used to login in the dashboard.
7. Create the database:
```sh
python scripts/setup_db.py
```
8. Open the ports 80 and 443 in your **firewall** or **Security Group**.
9. Configure the Nginx reverse-proxy by changing the **server_name** value from **etc/nginx/dashboard** to your domain name.
```
server {
    listen 80;
    server_name DOMAIN_NAME; # Change to your domain name
```
9. Replace the Nginx default configuration files:
```bash
sudo rm /etc/nginx/nginx.conf && sudo ln etc/nginx/nginx.conf /etc/nginx/
sudo ln etc/nginx/dashboard /etc/nginx/sites-enabled/
```
10. Link, enable, and start the Systemd services the Dashboard, the backend API, and Nginx.
```bash
sudo systemctl link $("pwd")/etc/systemd/dashboard.service
sudo systemctl enable dashboard.service
sudo systemctl restart dashboard.service

sudo systemctl link $("pwd")/etc/systemd/dashboard-api.service
sudo systemctl enable dashboard-api.service
sudo systemctl restart dashboard-api.service

sudo systemctl enable nginx.service
sudo systemctl start nginx.service
```
11. Now any requests to port 80 will be redirected to the dashboard. You can test by accessing the following URL: http://YOUR_DOMAIN_NAME
12. You can use Certbot to automatically generate TLS/SSL certificates and configure the Nginx to use it. Use [this tutorial](https://certbot.eff.org/instructions?ws=nginx&os=ubuntufocal) to install and configure Nginx with Certbot.
