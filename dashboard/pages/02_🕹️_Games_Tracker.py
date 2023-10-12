import base64
from io import BytesIO
from datetime import date, datetime

import streamlit as st
from streamlit_calendar import calendar

from dashboard.api.client import GameProperties, get_api_client

st.set_page_config(
    page_title="Games Tracker",
    page_icon="🕹️",
    layout="wide",
)


class GamesTrackerPage:
    def __init__(self) -> None:
        self.api_client = get_api_client()
        self.game_priority_options = {
            "High": "🤩 High",
            "Medium": "😆 Medium",
            "Low": "🙂 Low"
        }

    def sidebar(self):
        # Show a game highlighted in the sidebar
        with st.sidebar.container():
            if (highlight_game := st.session_state.get("game_to_be_highlighted", None)) is not None:
                with st.expander(highlight_game["Name"], True):
                    self._show_game_img_priority_release_date_purchased_gamepass(highlight_game, {}, False)
        with st.sidebar.expander("Add a game"):
            self.add_game()
        self.api_client.show_all_jobs_updating()

    def show(self):
        st.markdown(
            "<h1 style='text-align: center; font-size: 75px'>Games Tracker</h1>",
            unsafe_allow_html=True,
        )
        to_be_released_tab, not_started_tab, finished_tab, dropped_tab = st.tabs(
            ["To be released", "Not started", "Finished", "Dropped"]
        )

        with to_be_released_tab:
            self.show_to_be_released_tab()
        with not_started_tab:
            self.show_not_started_tab()
        with finished_tab:
            self.show_finished_tab()
        with dropped_tab:
            self.show_dropped_tab()

        self.sidebar()

    def show_to_be_released_tab(self):
        games = self.api_client.get_to_be_released_games()
        games_gallery_col, calendar_col = st.columns([0.35, 0.75], gap="small")

        with games_gallery_col:
            self.show_games(st.columns(2), games, self._show_game_img_priority_release_date_purchased_gamepass)
        with calendar_col:
            calendar_events = list()
            for name, game in games.items():
                calendar_events.append({
                    "title": name,
                    "start": game["ReleaseDate"]
                })
            calendar_options = {
                "initialDate": date.today().strftime("%Y-%m-%d"),
                "initialView": "dayGridMonth",
                "fixedWeekCount": False,
                "navLinks": True,
                "headerToolbar": {
                    "left": "today prev,next",
                    "center": "title",
                    "right": "resourceTimelineDay,resourceTimelineWeek,resourceTimelineMonth",
                },
                "titleFormat": {
                    "year": "numeric",
                    "month": "long"
                }
            }
            custom_css = """
                .fc-event-past {
                    opacity: 0.8;
                }
                .fc-scrollgrid-liquid {
                    height: 59%;
                }
                .fc-event-time {
                    font-style: italic;
                }
                .fc-event-title {
                    font-weight: 700;
                }
                .fc-toolbar-title {
                    font-size: 2rem;
                }
            """
            to_be_released_calendar = calendar(
                events=calendar_events,
                options=calendar_options,
                custom_css=custom_css,
                key="to_be_released_games_calendar"
            )
            event_click = to_be_released_calendar.get("eventClick", None)
            if event_click is not None:
                game_name = event_click["event"]["title"]
                if game_name != st.session_state.get("game_name_to_be_highlighted"):
                    st.session_state["game_name_to_be_highlighted"] = game_name
                    st.session_state["game_to_be_highlighted"] = games[game_name]
                    st.rerun()

    def show_not_started_tab(self):
        games = self.api_client.get_not_started_games()
        self.show_games(st.columns(5), games, self._show_game_img_priority_purchased_gamepass)

    def show_finished_tab(self):
        games = self.api_client.get_finished_games()
        self.show_games(st.columns(5), games, self._show_game_img_priority_finished_date_stars)

    def show_dropped_tab(self):
        games = self.api_client.get_dropped_games()
        self.show_games(st.columns(5), games, self._show_game_img_priority_dropped_date_stars)

    def _show_game_img_priority_release_date_purchased_gamepass(
            self, game: dict,
            games: dict,
            show_highlight_button: bool = True
    ):
        img_bytes = base64.b64decode(game["CoverImg"])
        img_stream = BytesIO(img_bytes)
        st.image(img_stream)
        st.write(self._get_priority(game["Priority"]))
        release_date = self._get_date(game["ReleaseDate"])
        st.write(release_date if release_date is not None else "No release date")
        st.write(self._get_purchased_gamepass(game["PurchasedOrGamePass"]))
        game_name = game["Name"]
        if show_highlight_button and st.button(
                "Highlight game",
                key=f"_show_game_img_priority_release_date_purchased_gamepass_{game_name}"
        ):
            if game_name != st.session_state.get("game_name_to_be_highlighted"):
                st.session_state["game_name_to_be_highlighted"] = game_name
                st.session_state["game_to_be_highlighted"] = games[game_name]
                st.rerun()

    def _show_game_img_priority_purchased_gamepass(self, game: dict, games: dict, show_highlight_button: bool = True):
        img_bytes = base64.b64decode(game["CoverImg"])
        img_stream = BytesIO(img_bytes)
        st.image(img_stream)
        st.write(self._get_priority(game["Priority"]))
        st.write(self._get_purchased_gamepass(game["PurchasedOrGamePass"]))
        game_name = game["Name"]
        if show_highlight_button and st.button(
                "Highlight game",
                key=f"_show_game_img_priority_purchased_gamepass_{game_name}"
        ):
            if game_name != st.session_state.get("game_name_to_be_highlighted"):
                st.session_state["game_name_to_be_highlighted"] = game_name
                st.session_state["game_to_be_highlighted"] = games[game_name]
                st.rerun()

    def _show_game_img_priority_finished_date_stars(self, game: dict, games: dict, show_highlight_button: bool = True):
        img_bytes = base64.b64decode(game["CoverImg"])
        img_stream = BytesIO(img_bytes)
        st.image(img_stream)
        st.write(self._get_priority(game["Priority"]))
        finished_date = self._get_date(game["FinishedDroppedDate"])
        st.write(finished_date if finished_date is not None else "No finished date")
        st.write(self._get_stars(game["Stars"]))
        game_name = game["Name"]
        if show_highlight_button and st.button(
                "Highlight game",
                key=f"show_game_img_priority_finished_dropped_date_stars_{game_name}"
        ):
            if game_name != st.session_state.get("game_name_to_be_highlighted"):
                st.session_state["game_name_to_be_highlighted"] = game_name
                st.session_state["game_to_be_highlighted"] = games[game_name]
                st.rerun()

    def _show_game_img_priority_dropped_date_stars(self, game: dict, games: dict, show_highlight_button: bool = True):
        img_bytes = base64.b64decode(game["CoverImg"])
        img_stream = BytesIO(img_bytes)
        st.image(img_stream)
        st.write(self._get_priority(game["Priority"]))
        dropped_date = self._get_date(game["FinishedDroppedDate"])
        st.write(dropped_date if dropped_date is not None else "No dropped date")
        st.write(self._get_stars(game["Stars"]))
        game_name = game["Name"]
        if show_highlight_button and st.button(
                "Highlight game",
                key=f"show_game_img_priority_finished_dropped_date_stars_{game_name}"
        ):
            if game_name != st.session_state.get("game_name_to_be_highlighted"):
                st.session_state["game_name_to_be_highlighted"] = game_name
                st.session_state["game_to_be_highlighted"] = games[game_name]
                st.rerun()

    def _get_purchased_gamepass(self, purchased_or_gamepass: bool):
        if purchased_or_gamepass:
            purchased_or_gamepass = "✅ Purchased/In Gamepass"
        else:
            purchased_or_gamepass = "❌ NOT Purchased/In Gamepass"

        return purchased_or_gamepass

    def _get_stars(self, stars: int):
        star = "⭐"
        stars_dict = {
            0: "No stars",
            1: star,
            2: star*2,
            3: star*3,
            4: star*4,
            5: star*5,
        }

        return stars_dict[stars]

    def _get_date(self, date_str: str):
        if date_str == "0001-01-01T00:00:00Z":
            return None
        else:
            return datetime.strptime(date_str, "%Y-%m-%dT%H:%M:%SZ").strftime("%B %d, %Y")

    def _get_priority(self, priority: str):
        correct_priority = self.game_priority_options.get(priority, None)
        if correct_priority is None:
            game_priority_options = {value: key for key, value in self.game_priority_options.items()}
            correct_priority = game_priority_options[priority]

        return correct_priority

    def show_games(self, cols_list: list, games: dict, show_games_func):
        """Show games in expanders in the cols_list columns.

        Args:
            cols_list (list): A list of streamlit.columns.
            games (dict): The key is the game name, and the value is a dict with the game properties.
            show_games_func: A function that expects a game and show the game.
        """
        col_index = 0
        for name, game in games.items():
            if col_index == len(cols_list):
                col_index = 0
            with cols_list[col_index]:
                with st.expander(name, True):
                    show_games_func(game, games)
            col_index += 1

    def add_game(self):
        with st.form("add_game_to_games_tracker_database", clear_on_submit=True):
            game_url = st.text_input(
                "Game URL",
                key="add_game_to_games_tracker_database_game_url",
                placeholder="https://store.steampowered.com/app/753640/Outer_Wilds/",
            )

            game_priority = st.selectbox(
                "Priority",
                options=({value: key for key, value in self.game_priority_options.items()}).keys(),
                key="add_game_to_games_tracker_database_game_priority",
            )
            game_priority = self._get_priority(game_priority)

            game_status_options = {
                "🗂️ Not started": "Not started",
                "📅 To be released": "To be released",
                "🎮 Playing": "Playing",
                "❌ Dropped": "Dropped",
                "✅ Finished": "Finished",
            }
            selected_game_status = st.selectbox(
                "Status",
                options=game_status_options.keys(),
                key="add_game_to_games_tracker_database_game_status",
            )
            game_status = game_status_options[selected_game_status]

            star = "⭐"
            game_star_options = {
                "I haven't decided": None,
                star: 1,
                f"{star * 2}": 2,
                f"{star * 3}": 3,
                f"{star * 4}": 4,
                f"{star * 5}": 5,
            }
            selected_game_star = st.selectbox(
                "Stars",
                options=game_star_options.keys(),
                key="add_game_to_games_tracker_database_game_stars",
            )
            game_stars = game_star_options[selected_game_star]

            st.write("")
            purchased_or_gamepass = st.checkbox(
                "Already purchased/in Gamepass",
                key="add_game_to_games_tracker_database_game_purchased_gamepass",
            )

            st.divider()
            game_started_date = st.date_input(
                "📅 Started playing date",
                key="add_game_to_games_tracker_database_game_started_date",
            )
            no_game_started_date = st.checkbox(
                "I don't know the started date",
                value=True,
                key="add_game_to_games_tracker_database_game_no_started_date",
            )
            st.divider()
            game_finished_dropped_date = st.date_input(
                "📅 Finished/Dropped date",
                key="add_game_to_games_tracker_database_game_finished_dropped_date",
            )
            no_game_finished_dropped_date = st.checkbox(
                "I don't know the finished/dropped date",
                value=True,
                key="add_game_to_games_tracker_database_game_no_finished_dropped_date",
            )
            st.divider()

            game_commentary = st.text_area(
                "Commentary",
                key="add_game_to_games_tracker_database_game_commentary",
                max_chars=1000,
            )

            submitted = st.form_submit_button()

            if submitted:
                if no_game_started_date:
                    game_started_date = None
                if no_game_finished_dropped_date:
                    game_finished_dropped_date = None

                game_properties = GameProperties(
                    game_url,
                    game_priority,
                    game_status,
                    purchased_or_gamepass,
                    game_stars,
                    game_started_date,
                    game_finished_dropped_date,
                    game_commentary,
                )

                self.api_client.add_game(game_properties)

                st.success("Game page requested")
                st.session_state["update_all_games"] = True


page = GamesTrackerPage()
page.show()
