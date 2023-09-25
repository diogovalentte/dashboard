import logging
import os
from pathlib import Path

import streamlit as st
import yaml
from streamlit.source_util import (_on_pages_changed, calc_md5, get_pages,
                                   page_icon_and_name)
from streamlit_authenticator import Authenticate

MAIN_SCRIPT_PATH = "01_🏠_Main_Page.py"
LOGGED_PAGES = ("02_📝_Notion_Page.py",)


def remove_page(main_script_path: str, page_name: str):
    """Hide/Remove a page from a multipage Streamlit application by not loading it.

    Args:
        main_script_path (str): The name of the file used to run the app, like "streamlit run main_script.py".
        page_name (str): The name of the page to hide/remove.
    """
    page_name = page_name.replace(".py", "")
    current_pages = get_pages(main_script_path)

    for key, value in current_pages.items():
        if value["page_name"] == page_name:
            del current_pages[key]
            break
        else:
            pass
    _on_pages_changed.send()


def hide_logged_pages():
    for page in LOGGED_PAGES:
        remove_page(MAIN_SCRIPT_PATH, page)


def add_page(main_script_path: str, page_file_name: str):
    """Add a page to a multipage Streamlit application to it load the page.

    Args:
        main_script_path (str): The name of the file used to run the app, like "streamlit run main_script.py".
        page_file_name (str): The name of the file to add.
    """
    absolute_path = os.path.abspath(os.path.dirname(__file__))
    page_path_str = os.path.join(absolute_path, "pages", page_file_name)
    if not os.path.exists(page_path_str):
        raise ValueError(f"Page not exists: {page_path_str}")

    page_path = Path(page_path_str)
    pages = get_pages(main_script_path)
    pi, pn = page_icon_and_name(page_path)

    psh = calc_md5(page_path_str)
    pages[psh] = {
        "page_script_hash": psh,
        "page_name": pn,
        "icon": pi,
        "script_path": page_path_str,
    }
    _on_pages_changed.send()


def show_logged_pages():
    for page in LOGGED_PAGES:
        add_page(MAIN_SCRIPT_PATH, page)


@st.cache_resource()
def get_credentials():
    logger = logging.getLogger("authenticator_page")
    logger.info("Reading credentials from YAML file...")

    with open(".streamlit/credentials/credentials.yaml") as file:
        configs = yaml.load(file, Loader=yaml.SafeLoader)

    logger.info("Credentials read")

    return configs


def show_authentication_page():
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
