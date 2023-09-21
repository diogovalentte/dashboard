# Personal Dashboard


## How to use:
1. Download firefox into **/opt/firefox/**:
```
curl -L -o /tmp/firefox-117.0.1.tar.bz2 "https://download.mozilla.org/?product=firefox-117.0.1&os=linux64&lang=en-US"
tar -xvf /tmp/firefox-117.0.1.tar.bz2 -C /opt
```
2. Download geckodriver into **/opt/geckodriver/**:
```
curl -L -o /tmp/geckodriver.tar.gz https://github.com/mozilla/geckodriver/releases/download/v0.33.0/geckodriver-v0.33.0-linux64.tar.gz
mkdir -p /opt/geckodriver/
tar -xvf /tmp/geckodriver.tar.gz -C /opt/geckodriver/
```
3. Create the file **configs/configs.json** with some configs and credentials. This file should follow the structure of the **configs/configs.example.json** file.
4. Build and start the API in a shell:
```
go build -o dashboard-api
./dashboard-api
```
