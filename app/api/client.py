import datetime
import logging
from urllib.parse import urljoin

import requests
import streamlit as st

from app.exceptions import APIException


@st.cache_data()
def get_api_client():
    api_client = st.session_state.get("api_client", None)
    if api_client is None:
        logger = logging.getLogger("api_client")
        logger.info("Defining the API client...")

        api_client = APIClient("http://localhost", 8080)
        st.session_state["api_client"] = api_client

        logger.info("API client defined")

    return api_client


class JSONBody:
    @property
    def json(self) -> dict:
        return self._get_json()

    def _get_json(self) -> dict:
        """Should return the class attributes in as a dict"""
        raise NotImplementedError()


class GameProperties(JSONBody):
    def __init__(
        self,
        URL: str,
        priority: str,
        status: str,
        purchased_or_gamepass: str,
        stars: str,
        started_date: datetime.date | None = None,
        finished_dropped_date: datetime.date | None = None,
        commentary: str | None = None,
    ) -> None:
        self.game_properties = {
            "url": URL,
            "priority": priority,
            "status": status,
            "purchased_or_gamepass": purchased_or_gamepass,
            "stars": stars,
            "started_date": started_date,
            "finished_dropped_date": finished_dropped_date,
            "commentary": commentary,
        }

    @property
    def json(self) -> dict:
        return self._get_json()

    def _get_json(self) -> dict:
        return self.game_properties


class APIClient:
    def __init__(self, base_URL: str, port: int) -> None:
        self.base_url = f"{base_URL}:{port}"

    def add_game(self, game_properties: GameProperties):
        path = "/v1/notion/games_tracker/add_game"
        url = urljoin(self.base_url, path)

        res = requests.post(url, json=game_properties.json)

        if res.status_code != 200:
            raise APIException(
                "error while adding game to games tracker database",
                url,
                "POST",
                res.status_code,
                res.text,
            )
