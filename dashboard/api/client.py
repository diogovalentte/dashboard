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

        api_client = APIClient(
            "http://localhost", 8080
        )  # The golang API docker service name
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


class SystemAPIClient:
    def __init__(self) -> None:
        self.base_url: str = ""

    def get_geckodriver_instances_addresses(self):
        path = "/v1/system/get_geckodrivers"
        url = urljoin(self.base_url, path)

        res = requests.get(url)
        if res.status_code != 200:
            raise APIException(
                "error while getting the geckodriver instances addresses from the API",
                url,
                "GET",
                res.status_code,
                res.text,
            )

        return res.json().get("addresses", [])

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


class TrackersAPIClient:
    def __init__(self) -> None:
        self.base_url: str = ""
        self.acceptable_status_codes: tuple = ()

    def add_media(self, media_properties: dict) -> None:
        path = "/v1/trackers/medias_tracker/add_media"
        url = urljoin(self.base_url, path)

        res = requests.post(url, json=media_properties)

        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while adding media to medias tracker database",
                url,
                "POST",
                res.status_code,
                res.text,
            )

    def add_media_manually(self, media_properties: dict) -> None:
        path = "/v1/trackers/medias_tracker/add_media_manually"
        url = urljoin(self.base_url, path)

        res = requests.post(url, json=media_properties)

        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while manually adding media to medias tracker database",
                url,
                "POST",
                res.status_code,
                res.text,
            )

    def update_media(self, game_properties: dict) -> None:
        path = "/v1/trackers/medias_tracker/update_media"
        url = urljoin(self.base_url, path)

        res = requests.post(url, json=game_properties)

        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while updating media on the medias tracker database",
                url,
                "POST",
                res.status_code,
                res.text,
            )

    def get_media(
        self, name: str
    ):
        path = "/v1/trackers/medias_tracker/get_media"
        url = urljoin(self.base_url, path)

        res = requests.post(url, json={"name": name})
        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                f"error while getting the media '{name}' from the medias tracker database",
                url,
                "POST",
                res.status_code,
                res.text,
            )

        game = res.json().get("media")

        return game

    def delete_media(self, name: str) -> None:
        path = "/v1/trackers/medias_tracker/delete_media"
        url = urljoin(self.base_url, path)

        res = requests.post(url, json={"name": name})

        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while deleting media on the medias tracker database",
                url,
                "POST",
                res.status_code,
                res.text,
            )

    def add_game(self, game_properties: dict) -> None:
        path = "/v1/trackers/games_tracker/add_game"
        url = urljoin(self.base_url, path)

        res = requests.post(url, json=game_properties)

        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while adding game to games tracker database",
                url,
                "POST",
                res.status_code,
                res.text,
            )

    def add_game_manually(self, game_properties: dict) -> None:
        path = "/v1/trackers/games_tracker/add_game_manually"
        url = urljoin(self.base_url, path)

        res = requests.post(url, json=game_properties)

        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while manually adding game to games tracker database",
                url,
                "POST",
                res.status_code,
                res.text,
            )

    def update_game(self, game_properties: dict) -> None:
        path = "/v1/trackers/games_tracker/update_game"
        url = urljoin(self.base_url, path)

        res = requests.post(url, json=game_properties)

        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while updating game on the games tracker database",
                url,
                "POST",
                res.status_code,
                res.text,
            )

    def get_game(
            self, name: str
    ):
        path = "/v1/trackers/games_tracker/get_game"
        url = urljoin(self.base_url, path)

        res = requests.post(url, json={"name": name})
        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                f"error while getting the game '{name}' from the games tracker database",
                url,
                "POST",
                res.status_code,
                res.text,
            )

        game = res.json().get("game")

        return game

    def delete_game(self, name: str) -> None:
        path = "/v1/trackers/games_tracker/delete_game"
        url = urljoin(self.base_url, path)

        res = requests.post(url, json={"name": name})

        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while deleting game on the games tracker database",
                url,
                "POST",
                res.status_code,
                res.text,
            )

    def get_all_games(
        self
    ):
        path = "/v1/trackers/games_tracker/get_all_games"
        url = urljoin(self.base_url, path)

        res = requests.get(url)
        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while getting all games from the games tracker database",
                url,
                "GET",
                res.status_code,
                res.text,
            )

        games = res.json().get("games")
        if games is not None:
            games = {game["Name"]: game for game in games}
        else:
            games = dict()

        return games

    def get_playing_games(self):
        """Return games that the user is currently playing"""
        path = "/v1/trackers/games_tracker/get_playing_games"
        url = urljoin(self.base_url, path)

        res = requests.get(url)
        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while getting playing games from the games tracker database",
                url,
                "GET",
                res.status_code,
                res.text,
            )

        games = res.json().get("games")
        if games is not None:
            games = {game["Name"]: game for game in games}
        else:
            games = dict()

        return games

    def get_to_be_released_games(self):
        """Return games that are to be released"""
        path = "/v1/trackers/games_tracker/get_to_be_released_games"
        url = urljoin(self.base_url, path)

        res = requests.get(url)
        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while getting to be released games from the games tracker database",
                url,
                "GET",
                res.status_code,
                res.text,
            )

        games = res.json().get("games")
        if games is not None:
            games = {game["Name"]: game for game in games}
        else:
            games = dict()

        return games

    def get_not_started_games(self):
        """Return games that were released but are not being played/not finished/dropped"""
        path = "/v1/trackers/games_tracker/get_not_started_games"
        url = urljoin(self.base_url, path)

        res = requests.get(url)
        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while getting not started games from the games tracker database",
                url,
                "GET",
                res.status_code,
                res.text,
            )

        games = res.json().get("games")
        if games is not None:
            games = {game["Name"]: game for game in games}
        else:
            games = dict()

        return games

    def get_finished_games(self):
        """Return games that were marked as finished"""
        path = "/v1/trackers/games_tracker/get_finished_games"
        url = urljoin(self.base_url, path)

        res = requests.get(url)
        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while getting finished games from the games tracker database",
                url,
                "GET",
                res.status_code,
                res.text,
            )

        games = res.json().get("games")
        if games is not None:
            games = {game["Name"]: game for game in games}
        else:
            games = dict()

        return games

    def get_dropped_games(self):
        """Return games that were marked as dropped"""
        path = "/v1/trackers/games_tracker/get_dropped_games"
        url = urljoin(self.base_url, path)

        res = requests.get(url)
        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while getting dropped games from the games tracker database",
                url,
                "GET",
                res.status_code,
                res.text,
            )

        games = res.json().get("games")
        if games is not None:
            games = {game["Name"]: game for game in games}
        else:
            games = dict()

        return games

    def get_all_medias(
            self
    ):
        path = "/v1/trackers/medias_tracker/get_all_medias"
        url = urljoin(self.base_url, path)

        res = requests.get(url)
        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while getting all medias from the medias tracker database",
                url,
                "GET",
                res.status_code,
                res.text,
            )

        medias = res.json().get("medias")
        if medias is not None:
            medias = {media["Name"]: media for media in medias}
        else:
            medias = dict()

        return medias

    def get_watching_reading_medias(
            self
    ):
        path = "/v1/trackers/medias_tracker/get_watching_reading_medias"
        url = urljoin(self.base_url, path)

        res = requests.get(url)
        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while getting watching/reading medias from the medias tracker database",
                url,
                "GET",
                res.status_code,
                res.text,
            )

        medias = res.json().get("medias")
        if medias is not None:
            medias = {media["Name"]: media for media in medias}
        else:
            medias = dict()

        return medias

    def get_to_be_released_medias(
            self
    ):
        path = "/v1/trackers/medias_tracker/get_to_be_released_medias"
        url = urljoin(self.base_url, path)

        res = requests.get(url)
        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while getting to be released medias from the medias tracker database",
                url,
                "GET",
                res.status_code,
                res.text,
            )

        medias = res.json().get("medias")
        if medias is not None:
            medias = {media["Name"]: media for media in medias}
        else:
            medias = dict()

        return medias

    def get_not_started_medias(
            self
    ):
        path = "/v1/trackers/medias_tracker/get_not_started_medias"
        url = urljoin(self.base_url, path)

        res = requests.get(url)
        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while getting not started medias from the medias tracker database",
                url,
                "GET",
                res.status_code,
                res.text,
            )

        medias = res.json().get("medias")
        if medias is not None:
            medias = {media["Name"]: media for media in medias}
        else:
            medias = dict()

        return medias

    def get_finished_medias(
            self
    ):
        path = "/v1/trackers/medias_tracker/get_finished_medias"
        url = urljoin(self.base_url, path)

        res = requests.get(url)
        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while getting finished medias from the medias tracker database",
                url,
                "GET",
                res.status_code,
                res.text,
            )

        medias = res.json().get("medias")
        if medias is not None:
            medias = {media["Name"]: media for media in medias}
        else:
            medias = dict()

        return medias

    def get_dropped_medias(
            self
    ):
        path = "/v1/trackers/medias_tracker/get_dropped_medias"
        url = urljoin(self.base_url, path)

        res = requests.get(url)
        if res.status_code not in self.acceptable_status_codes:
            raise APIException(
                "error while getting dropped medias from the medias tracker database",
                url,
                "GET",
                res.status_code,
                res.text,
            )

        medias = res.json().get("medias")
        if medias is not None:
            medias = {media["Name"]: media for media in medias}
        else:
            medias = dict()

        return medias

class APIClient(JobsAPIClient, TrackersAPIClient, SystemAPIClient):
    def __init__(self, base_URL: str, port: int) -> None:
        self.base_url = f"{base_URL}:{port}"
        self.acceptable_status_codes = (200, 400)
