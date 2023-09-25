import streamlit as st

from dashboard.api.client import (GameProperties, MediaProperties,
                                  get_api_client)

st.set_page_config(
    page_title="Notion",
    page_icon="📝",
    layout="wide",
)


class NotionPage:
    def __init__(self) -> None:
        self.api_client = get_api_client()

    def sidebar(self):
        self.api_client.show_all_jobs_updating()

    def show(self):
        st.header("Notion")
        games_tracker_tab, media_tracker_tab = st.tabs(
            ["Games tracker", "Media tracker"]
        )

        with games_tracker_tab:
            self.games_tracker_tab()

        with media_tracker_tab:
            self.medias_tracker_tab()

        self.sidebar()

    def games_tracker_tab(self):
        st.header("Add a game")
        with st.form("add_game_to_games_tracker_database", clear_on_submit=True):
            game_url = st.text_input(
                "Game URL",
                key="add_game_to_games_tracker_database_game_url",
                placeholder="https://store.steampowered.com/app/753640/Outer_Wilds/",
            )

            game_priority_options = ["🤩 High", "😆 Medium", "🙂 Low"]
            game_priority = st.selectbox(
                "Priority",
                options=game_priority_options,
                key="add_game_to_games_tracker_database_game_priority",
            )
            game_priority = game_priority.split(" ")[1]

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
            with st.expander("📅 Started playing date"):
                game_started_date = st.date_input(
                    "invisible label",
                    label_visibility="collapsed",
                    key="add_game_to_games_tracker_database_game_started_date",
                )
                no_game_started_date = st.checkbox(
                    "I don't know the started date",
                    value=True,
                    key="add_game_to_games_tracker_database_game_no_started_date",
                )

            with st.expander("📅 Finished/Dropped date"):
                game_finished_dropped_date = st.date_input(
                    "invisible label",
                    label_visibility="collapsed",
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
                max_chars=100,
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

                st.success("Requested game page")

    def medias_tracker_tab(self):
        st.header("Add a media")
        with st.form("add_media_to_medias_tracker_database", clear_on_submit=True):
            media_url = st.text_input(
                "Media URL",
                key="add_media_to_medias_tracker_database_media_url",
                placeholder="https://www.imdb.com/title/tt0816692/",
            )

            media_type_options = {
                "📺 Series": "Series",
                "🍿 Movie": "Movie",
                "📖 Book": "Book",
                "🗯️ Comic book": "Comic book",
            }
            selected_media_type = st.selectbox(
                "Type",
                options=media_type_options.keys(),
                key="add_media_to_medias_tracker_database_media_type",
            )
            media_type = media_type_options[selected_media_type]

            media_priority_options = ["🤩 High", "😆 Medium", "🙂 Low"]
            media_priority = st.selectbox(
                "Priority",
                options=media_priority_options,
                key="add_media_to_medias_tracker_database_media_priority",
            )
            media_priority = media_priority.split(" ")[1]

            media_status_options = {
                "🗂️ Not started": "Not started",
                "📅 To be released": "To be released",
                "🍿 Watching/Reading": "Watching/Reading",
                "❌ Dropped": "Dropped",
                "✅ Finished": "Finished",
            }
            selected_media_status = st.selectbox(
                "Status",
                options=media_status_options.keys(),
                key="add_media_to_medias_tracker_database_media_status",
            )
            media_status = media_status_options[selected_media_status]

            star = "⭐"
            media_star_options = {
                "I haven't decided": None,
                star: 1,
                f"{star * 2}": 2,
                f"{star * 3}": 3,
                f"{star * 4}": 4,
                f"{star * 5}": 5,
            }
            selected_media_star = st.selectbox(
                "Stars",
                options=media_star_options.keys(),
                key="add_media_to_medias_tracker_database_media_stars",
            )
            media_stars = media_star_options[selected_media_star]

            st.divider()
            with st.expander("📅 Started playing date"):
                media_started_date = st.date_input(
                    "invisible label",
                    label_visibility="collapsed",
                    key="add_media_to_medias_tracker_database_media_started_date",
                )
                no_media_started_date = st.checkbox(
                    "I don't know the started date",
                    value=True,
                    key="add_media_to_medias_tracker_database_media_no_started_date",
                )

            with st.expander("📅 Finished/Dropped date"):
                media_finished_dropped_date = st.date_input(
                    "invisible label",
                    label_visibility="collapsed",
                    key="add_media_to_medias_tracker_database_media_finished_dropped_date",
                )
                no_media_finished_dropped_date = st.checkbox(
                    "I don't know the finished/dropped date",
                    value=True,
                    key="add_media_to_medias_tracker_database_media_no_finished_dropped_date",
                )
            st.divider()

            media_commentary = st.text_area(
                "Commentary",
                key="add_media_to_medias_tracker_database_media_commentary",
                max_chars=100,
            )

            submitted = st.form_submit_button()

            if submitted:
                if no_media_started_date:
                    media_started_date = None
                if no_media_finished_dropped_date:
                    media_finished_dropped_date = None

                media_properties = MediaProperties(
                    media_url,
                    media_type,
                    media_priority,
                    media_status,
                    media_stars,
                    media_started_date,
                    media_finished_dropped_date,
                    media_commentary,
                )

                self.api_client.add_media(media_properties)

                st.success("Requested media page")


page = NotionPage()
page.show()
