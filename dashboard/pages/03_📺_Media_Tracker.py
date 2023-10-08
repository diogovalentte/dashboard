import streamlit as st

from dashboard.api.client import MediaProperties, get_api_client

st.set_page_config(
    page_title="Media Tracker",
    page_icon="📺",
    layout="wide",
)


class MediaTrackerPage:
    def __init__(self) -> None:
        self.api_client = get_api_client()

    def sidebar(self):
        self.api_client.show_all_jobs_updating()

    def show(self):
        st.markdown(
            "<h1 style='text-align: center; font-size: 75px'>Media Tracker</h1>",
            unsafe_allow_html=True,
        )
        media_tracker_tab, add_media_tab = st.tabs(["Media tracker", "Add a media"])

        with add_media_tab:
            self.add_media()

    def add_media(self):
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
            col1, col2 = st.columns(2)
            with col1:
                media_started_date = st.date_input(
                    "📅 Started playing date",
                    key="add_media_to_medias_tracker_database_media_started_date",
                )
                no_media_started_date = st.checkbox(
                    "I don't know the started date",
                    value=True,
                    key="add_media_to_medias_tracker_database_media_no_started_date",
                )

            with col2:
                media_finished_dropped_date = st.date_input(
                    "📅 Finished/Dropped date",
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


page = MediaTrackerPage()
page.show()
