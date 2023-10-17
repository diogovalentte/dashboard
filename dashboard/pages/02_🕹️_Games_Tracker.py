import base64
import random
from io import BytesIO
from datetime import date, datetime

import streamlit as st
from streamlit_calendar import calendar
from streamlit_extras.tags import tagger_component

from dashboard.api.client import get_api_client


st.set_page_config(
    page_title="Games Tracker",
    page_icon="🕹️",
    layout="wide",
)


class GamesTrackerPage:
    def __init__(self) -> None:
        self.api_client = get_api_client()
        self._game_priority_options = {
            1: "🤩 High",
            2: "😆 Medium",
            3: "🙂 Low"
        }
        self._game_status_options = {
            1: "📅 To be released",
            2: "🗂️ Not started",
            3: "🎮 Playing",
            4: "✅ Finished",
            5: "❌ Dropped",
        }
        star = "⭐"
        self._game_stars_options = {
            0: "I haven't decided",
            1: star,
            2: f"{star * 2}",
            3: f"{star * 3}",
            4: f"{star * 4}",
            5: f"{star * 5}",
        }

    def sidebar(self):
        with st.sidebar.expander("Add a game"):
            self.add_game()
        with st.sidebar.expander("Playing games"):
            self.show_playing_games_tab()
        self.api_client.show_all_jobs_updating()

    def show(self):
        st.markdown(
            "<h1 style='text-align: center; font-size: 75px'>Games Tracker</h1>",
            unsafe_allow_html=True,
        )

        if (highlight_game := st.session_state.get("game_to_be_highlighted", None)) is not None:
            to_be_released_tab, not_started_tab, finished_tab, dropped_tab, update_game_tab, highlight_game_tab = st.tabs(
                ["To be released", "Not started", "Finished", "Dropped", "Update game", highlight_game["Name"]]
            )
        else:
            to_be_released_tab, not_started_tab, finished_tab, dropped_tab, update_game_tab = st.tabs(
                ["To be released", "Not started", "Finished", "Dropped", "Update game"]
            )


        with to_be_released_tab:
            self.show_to_be_released_tab()
        with not_started_tab:
            self.show_not_started_tab()
        with finished_tab:
            self.show_finished_tab()
        with dropped_tab:
            self.show_dropped_tab()
        with update_game_tab:
            self.show_update_game_tab()
        if highlight_game is not None:
            with highlight_game_tab:
                self.show_highlighted_game_tab()

        self.sidebar()

    def show_playing_games_tab(self):
        games = self.api_client.get_playing_games()
        if len(games) == 0:
            st.info("No playing games")
        else:
            for game in games.values():
                self._show_playing_game(game, games)

    def show_to_be_released_tab(self):
        games = self.api_client.get_to_be_released_games()
        games_gallery_col, calendar_col = st.columns([0.35, 0.75], gap="small")

        with games_gallery_col:
            self.show_games(st.columns(2), games, self._show_to_be_released_game)
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
        self.show_games(st.columns(5), games, self._show_not_started_game)

    def show_finished_tab(self):
        games = self.api_client.get_finished_games()
        self.show_games(st.columns(5), games, self._show_finished_game)

    def show_dropped_tab(self):
        games = self.api_client.get_dropped_games()
        self.show_games(st.columns(5), games, self._show_dropped_game)

    def show_update_game_tab(self):
        self._show_update_game()

    def show_highlighted_game_tab(self):
        highlighted_game_name = st.session_state["game_name_to_be_highlighted"]
        game = self.api_client.get_game(highlighted_game_name)
        self._show_highlighted_game(game)

    def _show_to_be_released_game(
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

    def _show_playing_game(self, game: dict, games: dict, show_highlight_button: bool = True):
        st.subheader(game["Name"])
        img_bytes = base64.b64decode(game["CoverImg"])
        img_stream = BytesIO(img_bytes)
        st.image(img_stream)
        st.write(self._get_priority(game["Priority"]))
        started_date = self._get_date(game["StartedDate"])
        if started_date is None:
            started_date = "No started date"
        st.write(started_date)
        game_name = game["Name"]
        if show_highlight_button and st.button(
                "Highlight game",
                key=f"show_playing_game_{game_name}"
        ):
            if game_name != st.session_state.get("game_name_to_be_highlighted"):
                st.session_state["game_name_to_be_highlighted"] = game_name
                st.session_state["game_to_be_highlighted"] = games[game_name]
                st.rerun()
        st.divider()

    def _show_not_started_game(self, game: dict, games: dict, show_highlight_button: bool = True):
        img_bytes = base64.b64decode(game["CoverImg"])
        img_stream = BytesIO(img_bytes)
        st.image(img_stream)
        st.write(self._get_priority(game["Priority"]))
        st.write(self._get_purchased_gamepass(game["PurchasedOrGamePass"]))
        game_name = game["Name"]
        if show_highlight_button and st.button(
                "Highlight game",
                key=f"show_not_started_game_{game_name}"
        ):
            if game_name != st.session_state.get("game_name_to_be_highlighted"):
                st.session_state["game_name_to_be_highlighted"] = game_name
                st.session_state["game_to_be_highlighted"] = games[game_name]
                st.rerun()

    def _show_finished_game(self, game: dict, games: dict, show_highlight_button: bool = True):
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
                key=f"show_finished_game_{game_name}"
        ):
            if game_name != st.session_state.get("game_name_to_be_highlighted"):
                st.session_state["game_name_to_be_highlighted"] = game_name
                st.session_state["game_to_be_highlighted"] = games[game_name]
                st.rerun()

    def _show_dropped_game(self, game: dict, games: dict, show_highlight_button: bool = True):
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
                key=f"show_dropped_game_{game_name}"
        ):
            if game_name != st.session_state.get("game_name_to_be_highlighted"):
                st.session_state["game_name_to_be_highlighted"] = game_name
                st.session_state["game_to_be_highlighted"] = games[game_name]
                st.rerun()

    def _show_highlighted_game(self, game: dict):
        game_properties_col, _, game_commentary_col = st.columns([0.25, 0.2, 0.55])
        with game_properties_col:
            # Name
            st.markdown(
                f"<h1 style='text-align: center; font-size: 44px'>{game['Name']}</h1>",
                unsafe_allow_html=True,
            )

            # Image
            img_bytes = base64.b64decode(game["CoverImg"])
            img_stream = BytesIO(img_bytes)
            st.image(img_stream, use_column_width=True)

            # Properties
            st.link_button("🛒 View on store", url=game["URL"])
            st.markdown(f'**Priority**: {self._get_priority(game["Priority"])}')
            st.markdown(f'**Status**: {self._get_status(game["Status"])}')
            st.markdown(f'**Stars**: {self._get_stars(game["Stars"])}')
            st.write(f'**Purchased/Gamepass?** {"✅" if game["PurchasedOrGamePass"] else "❌"}')

            # Dates
            release_date = self._get_date(game["ReleaseDate"])
            st.markdown(f"**Release date**: {release_date}" if release_date is not None else "**No release date.**")
            started_date = self._get_date(game["FinishedDroppedDate"])
            st.markdown(f"**Started date**: {started_date}" if started_date is not None else "**No started date.**")
            dropped_date = self._get_date(game["FinishedDroppedDate"])
            st.markdown(f"**Dropped/Finished date**: {dropped_date}" if dropped_date is not None else "**No dropped/finished date.**")

            # Developers
            base_pub_dev_html = """
                <a href="https://store.steampowered.com/search/?term={}/" target="_blank" style="text-decoration: none; color: white;">
                    <span>{}</span>
                </a>
            """
            game["Developers"] = [base_pub_dev_html.format(developer.replace(" ", "%20"), developer) for developer in game["Developers"]]
            tagger_component(
                "<strong>Developers:</strong>",
                game["Developers"],
                self._get_tag_colors(len(game["Developers"]))
            )
            # Publishers
            game["Publishers"] = [base_pub_dev_html.format(publisher.replace(" ", "%20"), publisher) for publisher in game["Publishers"]]
            tagger_component(
                "<strong>Publishers:</strong>",
                game["Publishers"],
                self._get_tag_colors(len(game["Publishers"]))
            )
            # Tags
            base_tag_html = """
                <a href="https://store.steampowered.com/tags/en/{}/" target="_blank" style="text-decoration: none; color: white;">
                    <span>{}</span>
                </a>
            """
            game["Tags"] = [base_tag_html.format(tag.replace(" ", "%20"), tag) for tag in game["Tags"]]
            tagger_component(
                "<strong>Tags</strong>: ",
                game["Tags"],
                self._get_tag_colors(len(game["Tags"]))
            )

        with game_commentary_col:
            st.markdown(game["Commentary"])

    def _get_purchased_gamepass(self, purchased_or_gamepass: bool):
        if purchased_or_gamepass:
            purchased_or_gamepass = "✅ Purchased/In Gamepass"
        else:
            purchased_or_gamepass = "❌ NOT Purchased/In Gamepass"

        return purchased_or_gamepass

    def _get_date(self, date_str: str):
        if date_str == "0001-01-01T00:00:00Z":
            return None
        else:
            return datetime.strptime(date_str, "%Y-%m-%dT%H:%M:%SZ").strftime("%B %d, %Y")

    def _get_priority(self, priority: int | str):
        correct_priority = self._game_priority_options.get(priority, None)
        if correct_priority is None:
            game_priority_options = {value: key for key, value in self._game_priority_options.items()}
            correct_priority = game_priority_options[priority]

        return correct_priority

    def _get_stars(self, stars: int | str):
        correct_stars = self._game_stars_options.get(stars, None)
        if correct_stars is None:
            game_stars_options = {value: key for key, value in self._game_stars_options.items()}
            correct_stars = game_stars_options[stars]

        return correct_stars

    def _get_status(self, status: int | str):
        correct_status = self._game_status_options.get(status, None)
        if correct_status is None:
            game_status_options = {value: key for key, value in self._game_status_options.items()}
            correct_status = game_status_options[status]

        return correct_status

    def _get_tag_colors(self, number: int):
        color_palette = [
            "lightblue",
            "orange",
            "bluegreen",
            "blue",
            "violet",
            "red",
            "green",
            "yellow",
        ]
        random.shuffle(color_palette)

        num_repeats = number // len(color_palette)

        # Create the color list by repeating the colors
        color_list = color_palette * num_repeats

        # Append any remaining colors to match the size
        color_list += color_palette[:number % len(color_palette)]

        return color_list

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

            selected_game_priority = st.selectbox(
                "Priority",
                options=self._game_priority_options.values(),
                key="add_game_to_games_tracker_database_game_priority",
            )

            selected_game_status = st.selectbox(
                "Status",
                options=self._game_status_options.values(),
                key="add_game_to_games_tracker_database_game_status",
            )

            selected_game_stars = st.selectbox(
                "Stars",
                options=self._game_stars_options.values(),
                key="add_game_to_games_tracker_database_game_stars",
            )

            st.write("")
            purchased_or_gamepass = st.checkbox(
                "Already purchased/in Gamepass?",
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
                game_priority = self._get_priority(selected_game_priority)
                game_status = self._get_status(selected_game_status)
                game_stars = self._get_stars(selected_game_stars)
                if no_game_started_date:
                    game_started_date = None
                if no_game_finished_dropped_date:
                    game_finished_dropped_date = None

                game_properties = {
                    "url": game_url,
                    "priority": game_priority,
                    "status": game_status,
                    "stars": game_stars,
                    "purchased_or_gamepass": purchased_or_gamepass,
                    "started_date": str(game_started_date) if game_started_date is not None else "",
                    "finished_dropped_date": str(game_finished_dropped_date) if game_finished_dropped_date is not None else "",
                    "commentary": game_commentary
                }

                self.api_client.add_game(game_properties)

                st.success("Game requested")
                st.session_state["update_all_games"] = True

    def _show_update_game(self):
        games = self.api_client.get_all_games()

        # Select game to update
        selected_game = st.selectbox(
            "Select a game to update",
            options=games.keys(),
            index=None,
            placeholder="Choose a game",
            key="update_game_on_games_tracker_database_select_game_to_update"
        )
        st.title("")
        if selected_game is not None:
            game = games[selected_game].copy()

            game["Status"] = self._get_status(game["Status"])
            game["Priority"] = self._get_priority(game["Priority"])
            game["Stars"] = self._get_stars(game["Stars"])

            release_date = datetime.strptime(game["ReleaseDate"], "%Y-%m-%dT%H:%M:%SZ")
            game["ReleaseDate"] = date(release_date.year, release_date.month, release_date.day)

            started_date = datetime.strptime(game["StartedDate"], "%Y-%m-%dT%H:%M:%SZ")
            game["StartedDate"] = date(started_date.year, started_date.month, started_date.day)

            finished_dropped_date = datetime.strptime(game["FinishedDroppedDate"], "%Y-%m-%dT%H:%M:%SZ")
            game["FinishedDroppedDate"] = date(
                finished_dropped_date.year,
                finished_dropped_date.month,
                finished_dropped_date.day
            )

            # Show update form
            edit_form_container = st.container()
            with st.expander("Update commentary"):
                edited_commentary = self.show_markdown_editor(game["Commentary"])

            with edit_form_container:
                with st.form("update_game_on_games_tracker_database"):
                    column_config = {
                        "Priority": st.column_config.SelectboxColumn(
                            options=self._game_priority_options.values(),
                            required=True
                        ),
                        "Status": st.column_config.SelectboxColumn(
                            options=self._game_status_options.values(),
                            required=True
                        ),
                        "Stars": st.column_config.SelectboxColumn(
                            options=self._game_stars_options.values(),
                            required=True
                        ),
                        "PurchasedOrGamePass": st.column_config.CheckboxColumn(
                            label="Purchased/Gamepass?",
                            required=True,
                            width="small"
                        ),
                        "StartedDate": st.column_config.DateColumn(
                            label="Started date",
                            required=True
                        ),
                        "FinishedDroppedDate": st.column_config.DateColumn(
                            label="Finished/Dropped date",
                            required=True
                        ),
                        "ReleaseDate": st.column_config.DateColumn(
                            label="Release date",
                            required=True
                        )
                    }
                    column_order = (
                        "Priority",
                        "Status",
                        "Stars",
                        "PurchasedOrGamePass",
                        "StartedDate",
                        "FinishedDroppedDate",
                        "ReleaseDate"
                    )
                    updated_game = st.data_editor(
                        [game],
                        use_container_width=True,
                        column_order=column_order,
                        column_config=column_config,
                        key="update_game_on_games_tracker_database_data_editor",
                    )[0]
                    st.write("")

                    if st.form_submit_button():
                        game_priority = self._get_priority(updated_game["Priority"])
                        game_status = self._get_status(updated_game["Status"])
                        game_stars = self._get_stars(updated_game["Stars"])

                        game_properties = {
                            "name": game["Name"],
                            "priority": game_priority,
                            "status": game_status,
                            "stars": game_stars,
                            "purchased_or_gamepass": updated_game["PurchasedOrGamePass"],
                            "started_date": str(updated_game["StartedDate"]) if updated_game["StartedDate"] is not None else "",
                            "finished_dropped_date": str(updated_game["FinishedDroppedDate"]) if updated_game["FinishedDroppedDate"] is not None else "",
                            "release_date": str(updated_game["ReleaseDate"]) if updated_game["ReleaseDate"] is not None else "",
                            "commentary": edited_commentary
                        }

                        self.api_client.update_game(game_properties)

                        st.success("Game update requested")
                        st.session_state["update_all_games"] = True

    def show_markdown_editor(self, original_body: str):
        markdown_viewer_col, markdown_editor_col = st.columns(2)
        with markdown_editor_col:
            edited_body = st.text_area(
                label="will be collapsed",
                value=original_body,
                label_visibility="collapsed",
                key="update_game_on_games_tracker_database_markdown_editor"
            )
        with markdown_viewer_col:
            st.markdown(edited_body)

        return edited_body


page = GamesTrackerPage()
page.show()
