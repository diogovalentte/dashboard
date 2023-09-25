import logging

import streamlit as st

from dashboard.api.client import get_api_client
from dashboard.utils import (hide_logged_pages, show_authentication_page,
                             show_logged_pages)

st.set_page_config(
    page_title="Personal dashboard",
    page_icon="🏠",
    layout="wide",
)


class MainPage:
    def sidebar(self):
        api_client = get_api_client()

        # Jobs
        api_client.show_all_jobs_updating()

    def show(self):
        st.markdown(
            "<h1 style='text-align: center; font-size: 75px'>Personal Dashboard</h1>",
            unsafe_allow_html=True,
        )

        self.sidebar()


def main():
    if st.session_state.get("authentication_status") in (None, False):
        hide_logged_pages()
        show_authentication_page()
    if st.session_state.get("authentication_status") is True:
        show_logged_pages()
        main_page = MainPage()
        main_page.show()


if __name__ == "__main__":
    logging.basicConfig(
        encoding="utf-8",
        level=logging.INFO,
        format="%(asctime)s :: %(levelname)-8s :: %(name)s :: %(message)s",
    )
    logger = logging.getLogger()

    try:
        main()
    except:
        logger.exception("An exception happened!")
        st.error("An error occurred.")
