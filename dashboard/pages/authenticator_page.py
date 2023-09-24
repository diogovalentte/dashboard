import logging

import streamlit as st
import yaml
from streamlit_authenticator import Authenticate


@st.cache_resource()
def get_credentials():
    logger = logging.getLogger("authenticator_page")
    logger.info("Reading credentials from YAML file...")

    with open(".streamlit/credentials/credentials.yaml") as file:
        configs = yaml.load(file, Loader=yaml.SafeLoader)

    logger.info("Credentials read")

    return configs


class AuthenticatorPage:
    def show(self):
        """Shows a authentication page with a form for the user authenticate"""
        configs = get_credentials()
        authenticator = Authenticate(
            configs["credentials"],
            configs["cookie"]["name"],
            configs["cookie"]["key"],
            configs["cookie"]["expiry_days"],
        )
        name, authentication_status, username = authenticator.login("Login", "main")

        st.session_state["name"] = name
        st.session_state["authentication_status"] = authentication_status
        st.session_state["username"] = username

        if st.session_state["authentication_status"] is False:
            st.error("Username/password is incorrect")
        elif st.session_state["authentication_status"] is None:
            st.warning("Please enter your username and password")
