import streamlit as st

from dashboard.api.client import GameProperties, get_api_client

st.set_page_config(
    page_title="Games Tracker",
    page_icon="🕹️",
    layout="wide",
)


class GamesTrackerPage:
    def __init__(self) -> None:
        self.api_client = get_api_client()

    def sidebar(self):
        self.api_client.show_all_jobs_updating()

    def show(self):
        st.markdown(
            "<h1 style='text-align: center; font-size: 75px'>Games Tracker</h1>",
            unsafe_allow_html=True,
        )
        games_tracker_tab, add_game_tab = st.tabs(["Games tracker", "Add a game"])

        with add_game_tab:
            self.add_game()

    def add_game(self):
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
            col1, col2 = st.columns(2)
            with col1:
                game_started_date = st.date_input(
                    "📅 Started playing date",
                    key="add_game_to_games_tracker_database_game_started_date",
                )
                no_game_started_date = st.checkbox(
                    "I don't know the started date",
                    value=True,
                    key="add_game_to_games_tracker_database_game_no_started_date",
                )

            with col2:
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


page = GamesTrackerPage()
page.show()
