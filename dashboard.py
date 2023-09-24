import logging

import streamlit as st

from dashboard.api.client import get_api_client
from dashboard.pages.authenticator_page import AuthenticatorPage
from dashboard.pages.main_page import MainPage
from dashboard.pages.notion_page import NotionPage

st.set_page_config(
    page_title="Personal dashboard",
    # page_icon="",
    layout="wide",
)

logging.basicConfig(
    encoding="utf-8",
    level=logging.INFO,
    format="%(asctime)s :: %(levelname)-8s :: %(name)s :: %(message)s",
)
logger = logging.getLogger()


def select_page():
    pages = {"Main page": MainPage, "Notion": NotionPage}

    st.sidebar.header("Select a page:")
    selected_page = st.sidebar.selectbox(
        "invisible label",
        label_visibility="collapsed",
        options=list(pages.keys()),
        key="select_page",
    )

    pages[selected_page]().show()


def sidebar():
    api_client = get_api_client()

    select_page()
    st.sidebar.divider()

    # Jobs
    api_client.show_all_jobs_updating()


def main():
    if st.session_state.get("authentication_status") in (None, False):
        authenticator_page = AuthenticatorPage()
        authenticator_page.show()
    if st.session_state.get("authentication_status") is True:
        sidebar()


if __name__ == "__main__":
    try:
        main()
    except:
        logger.exception("An exception happened!")
        st.error("An error occurred.")
