import datetime
import logging
import time
from urllib.parse import urljoin

import requests
import streamlit as st

from dashboard.exceptions import APIException


@st.cache_data()
def get_api_client():
    api_client = st.session_state.get("api_client", None)
    if api_client is None:
        logger = logging.getLogger("api_client")
        logger.info("Defining the API client...")

        api_client = APIClient("http://api", 8080)  # The golang API docker service name
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
            "started_date": str(started_date) if started_date is not None else "",
            "finished_dropped_date": str(finished_dropped_date)
            if finished_dropped_date is not None
            else "",
            "commentary": commentary,
        }

    @property
    def json(self) -> dict:
        return self._get_json()

    def _get_json(self) -> dict:
        return self.game_properties


class MediaProperties(JSONBody):
    def __init__(
        self,
        URL: str,
        media_type: str,
        priority: str,
        status: str,
        stars: str,
        started_date: datetime.date | None = None,
        finished_dropped_date: datetime.date | None = None,
        commentary: str | None = None,
    ) -> None:
        self.media_properties = {
            "url": URL,
            "type": media_type,
            "priority": priority,
            "status": status,
            "stars": stars,
            "started_date": str(started_date) if started_date is not None else "",
            "finished_dropped_date": str(finished_dropped_date)
            if finished_dropped_date is not None
            else "",
            "commentary": commentary,
        }

    @property
    def json(self) -> dict:
        return self._get_json()

    def _get_json(self) -> dict:
        return self.media_properties


class JobsAPIClient:
    def __init__(self) -> None:
        self.base_url: str = ""

    def get_all_jobs(self):
        path = "/v1/jobs/get_all"
        url = urljoin(self.base_url, path)

        res = requests.get(url)
        if res.status_code != 200:
            raise APIException(
                "error while getting all jobs from the API",
                url,
                "GET",
                res.status_code,
                res.text,
            )

        return res.json().get("jobs", [])

    def delete_all_jobs(self):
        path = "/v1/jobs/delete_all"
        url = urljoin(self.base_url, path)

        res = requests.delete(url)
        if res.status_code != 200:
            raise APIException(
                "error while deleting all jobs of the API",
                url,
                "DELETE",
                res.status_code,
                res.text,
            )

    def show_all_jobs(self, jobs_placeholder: st.delta_generator.DeltaGenerator):
        jobs = self.get_all_jobs()

        states = {
            "Completed": "complete",
            "Failed": "error",
            "Starting": "running",
            "Executing": "running",
        }
        if len(jobs) > 0:
            with jobs_placeholder.container():
                jobs.reverse()
                for job in jobs:
                    state = job["State"]
                    state_description = job["StateDescription"]
                    value = job["Value"]
                    created_at = job["CreatedAt"]
                    completed_failed_at = job["Completed_Failed_At"]

                    expanded = True if state in ("Starting", "Executing") else False

                    status = st.status(
                        job["Task"], state=states[state], expanded=expanded
                    )

                    if state in ("Starting", "Executing"):
                        status.info(state_description)
                    elif state == "Completed":
                        status.success(state_description)
                    elif state == "Failed":
                        status.error(state_description)

                    if value != "":
                        status.info(value)

                    status.markdown(
                        f"<strong>Created at</strong>: {created_at}",
                        unsafe_allow_html=True,
                    )
                    if completed_failed_at != "":
                        status.markdown(
                            f"<strong>{state} at</strong>: {completed_failed_at}",
                            unsafe_allow_html=True,
                        )

    def show_all_jobs_updating(self, seconds: int = 1):
        st.sidebar.title("Jobs")
        jobs_placeholder = st.sidebar.empty()
        while True:
            self.show_all_jobs(jobs_placeholder)
            time.sleep(seconds)


class NotionAPIClient:
    def __init__(self) -> None:
        self.base_url: str = ""
        self.acceptable_status_codes: tuple = ()

    def add_game(self, game_properties: GameProperties) -> None:
        path = "/v1/notion/games_tracker/add_game"
        url = urljoin(self.base_url, path)

        res = requests.post(url, json=game_properties.json)

        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while adding game to games tracker database",
                url,
                "POST",
                res.status_code,
                res.text,
            )

    def add_media(self, media_properties: MediaProperties) -> None:
        path = "/v1/notion/medias_tracker/add_media"
        url = urljoin(self.base_url, path)

        res = requests.post(url, json=media_properties.json)

        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while adding media to medias tracker database",
                url,
                "POST",
                res.status_code,
                res.text,
            )


class APIClient(JobsAPIClient, NotionAPIClient):
    def __init__(self, base_URL: str, port: int) -> None:
        self.base_url = f"{base_URL}:{port}"
        self.acceptable_status_codes = (200, 400)
