# Personal Dashboard


## How to use:
1. Install [Docker](https://www.docker.com).
2. Create the file **configs/configs.json** with some configs and credentials. This file should follow the structure of the **configs/configs.example.json** file.
3. The dashboard uses the [Streamlit Authenticator](https://github.com/mkhorasani/Streamlit-Authenticator/tree/main) module, check [here](https://github.com/mkhorasani/Streamlit-Authenticator/tree/main#1-hashing-passwords) how to create the file **.streamlit/credentials/credentials.yaml** (should be at this location!) with the users/passwords used to login in the dashboard.
4. Install [Nginx](https://www.nginx.com). Nginx will act as a **reverse proxy** for the Streamlit dashboard app.
5. Open the ports 80 and 443 in your **firewall** or **Security Group**.
6. Configure the Nginx reverse-proxy by changing the **server_name** value from **etc/nginx/dashboard** to your domain name.
```
server {
    listen 80;
    server_name DOMAIN_NAME; # Change to your domain name
```
7. Replace the Nginx default configuration files:
```bash
sudo rm /etc/nginx/nginx.conf && sudo ln etc/nginx/nginx.conf /etc/nginx/
sudo ln etc/nginx/dashboard /etc/nginx/sites-enabled/
```
8. Link, enable, and start the Systemd service of the Streamlit dashboard and Nginx.
```bash
sudo systemctl link $("pwd")/etc/systemd/dashboard.service
sudo systemctl enable dashboard.service
sudo systemctl restart dashboard.service

sudo systemctl enable nginx.service
sudo systemctl start nginx.service
```
9. Now any requests to port 80 will be redirected to the dashboard. You can test by accessing the following URL: http://YOUR_DOMAIN_NAME
10. You can use Certbot to automatically generate TLS/SSL certificates and configure the Nginx to use it. Use [this tutorial](https://certbot.eff.org/instructions?ws=nginx&os=ubuntufocal) to install and configure Nginx with Certbot.
