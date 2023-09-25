#!/bin/bash

python3 -m venv .docker-python-venv
source .docker-python-venv/bin/activate
pip install -r requirements.txt
.docker-python-venv/bin/streamlit run 01_🏠_Main_Page.py
