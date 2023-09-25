# Personal Dashboard


## How to use:
1. Install [Docker](https://www.docker.com).
2. Create the file **configs/configs.json** with some configs and credentials. This file should follow the structure of the **configs/configs.example.json** file.
3. The dashboard uses the [Streamlit Authenticator](https://github.com/mkhorasani/Streamlit-Authenticator/tree/main) module, check [here](https://github.com/mkhorasani/Streamlit-Authenticator/tree/main#1-hashing-passwords) how to create the file **.streamlit/credentials/credentials.yaml** (should be at this location!) with the users/passwords used to login in the dashboard.
4. Start the API and the dashboard in the background:
```
docker compose up -d
```
5. Access the dashboard at [http://localhost:8501](http://localhost:8501).
