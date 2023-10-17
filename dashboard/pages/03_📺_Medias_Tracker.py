import base64
import random
from io import BytesIO
from datetime import date, datetime

import streamlit as st
from streamlit_calendar import calendar
from streamlit_extras.tags import tagger_component

from dashboard.api.client import get_api_client

st.set_page_config(
    page_title="Medias Tracker",
    page_icon="📺",
    layout="wide",
)


class MediasTrackerPage:
    def __init__(self) -> None:
        self.api_client = get_api_client()
        self._media_priority_options = {
            1: "🤩 High",
            2: "😆 Medium",
            3: "🙂 Low"
        }
        self._media_status_options = {
            1: "🗂️ Not started",
            2: "📅 To be released",
            3: "🍿 Watching/Reading",
            4: "✅ Finished",
            5: "❌ Dropped",
        }
        self._media_type_options = {
            1: "📺 Series",
            2: "🍿 Movie",
            3: "📖 Book",
            4: "🗯️ Comic book",
        }
        star = "⭐"
        self._media_stars_options = {
            0: "I haven't decided",
            1: star,
            2: f"{star * 2}",
            3: f"{star * 3}",
            4: f"{star * 4}",
            5: f"{star * 5}",
        }

    def sidebar(self):
        with st.sidebar.expander("Add a media"):
            self.add_media()
        with st.sidebar.expander("Watching/Reading medias"):
            self.show_watching_reading_medias_tab()
        # self.api_client.show_all_jobs_updating()

    def show(self):
        st.markdown(
            "<h1 style='text-align: center; font-size: 75px'>Medias Tracker</h1>",
            unsafe_allow_html=True,
        )

        if (media_name_to_be_highlighted := st.session_state.get("media_name_to_be_highlighted", None)) is not None:
            to_be_released_tab, not_started_tab, finished_tab, dropped_tab, update_media_tab, highlight_media_tab = st.tabs(
                ["To be released", "Not started", "Finished", "Dropped", "Update media", media_name_to_be_highlighted]
            )
        else:
            to_be_released_tab, not_started_tab, finished_tab, dropped_tab, update_media_tab = st.tabs(
                ["To be released", "Not started", "Finished", "Dropped", "Update media"]
            )

        with to_be_released_tab:
            self.show_to_be_released_tab()
        with not_started_tab:
            self.show_not_started_tab()
        with finished_tab:
            self.show_finished_tab()
        with dropped_tab:
            self.show_dropped_tab()
        with update_media_tab:
            self.show_update_media_tab()
        if media_name_to_be_highlighted is not None:
            with highlight_media_tab:
                self.show_highlighted_media_tab()

        self.sidebar()

    def show_watching_reading_medias_tab(self):
        medias = self.api_client.get_watching_reading_medias()
        if len(medias) == 0:
            st.info("No watching/reading medias")
        else:
            for media in medias.values():
                self._show_watching_reading_media(media, medias)

    def show_to_be_released_tab(self):
        medias = self.api_client.get_to_be_released_medias()
        medias_gallery_col, calendar_col = st.columns([0.35, 0.75], gap="small")

        with medias_gallery_col:
            self.show_medias(st.columns(3), medias, self._show_to_be_released_media)
        with calendar_col:
            calendar_events = list()
            for name, media in medias.items():
                calendar_events.append({
                    "title": name,
                    "start": media["ReleaseDate"]
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
                key="to_be_released_medias_calendar"
            )
            event_click = to_be_released_calendar.get("eventClick", None)
            if event_click is not None:
                media_name = event_click["event"]["title"]
                if media_name != st.session_state.get("media_name_to_be_highlighted"):
                    st.session_state["media_name_to_be_highlighted"] = media_name
                    st.session_state["media_to_be_highlighted"] = medias[media_name]
                    st.rerun()

    def show_not_started_tab(self):
        medias = self.api_client.get_not_started_medias()
        self.show_medias(st.columns(10), medias, self._show_not_started_media)

    def show_finished_tab(self):
        medias = self.api_client.get_finished_medias()
        self.show_medias(st.columns(10), medias, self._show_finished_media)

    def show_dropped_tab(self):
        medias = self.api_client.get_dropped_medias()
        self.show_medias(st.columns(10), medias, self._show_dropped_media)

    def show_update_media_tab(self):
        self._show_update_media()

    def show_highlighted_media_tab(self):
        highlighted_media_name = st.session_state["media_name_to_be_highlighted"]
        media = self.api_client.get_media(highlighted_media_name)
        self._show_highlighted_media(media)

    def _show_watching_reading_media(self, media: dict, medias: dict, show_highlight_button: bool = True):
        st.subheader(media["Name"])
        img_bytes = base64.b64decode(media["CoverImg"])
        img_stream = BytesIO(img_bytes)
        st.image(img_stream, use_column_width=True)
        st.write(self._get_priority(media["Priority"]))
        started_date = self._get_date(media["StartedDate"])
        if started_date is None:
            started_date = "No started date"
        st.write(started_date)
        media_name = media["Name"]
        if show_highlight_button and st.button(
                "Highlight media",
                key=f"show_watching_reading_media_{media_name}"
        ):
            if media_name != st.session_state.get("media_name_to_be_highlighted"):
                st.session_state["media_name_to_be_highlighted"] = media_name
                st.session_state["media_to_be_highlighted"] = medias[media_name]
                st.rerun()
        st.divider()

    def _show_to_be_released_media(
            self, media: dict,
            medias: dict,
            show_highlight_button: bool = True
    ):
        img_bytes = base64.b64decode(media["CoverImg"])
        img_stream = BytesIO(img_bytes)
        st.image(img_stream, use_column_width=True)
        st.write(self._get_media_type(media["MediaType"]))
        st.write(self._get_priority(media["Priority"]))
        release_date = self._get_date(media["ReleaseDate"])
        st.write(release_date if release_date is not None else "No release date")
        media_name = media["Name"]
        if show_highlight_button and st.button(
                "Highlight media",
                key=f"show_to_be_released_media_{media_name}"
        ):
            if media_name != st.session_state.get("media_name_to_be_highlighted"):
                st.session_state["media_name_to_be_highlighted"] = media_name
                st.session_state["media_to_be_highlighted"] = medias[media_name]
                st.rerun()

    def _show_not_started_media(self, media: dict, medias: dict, show_highlight_button: bool = True):
        img_bytes = base64.b64decode(media["CoverImg"])
        img_stream = BytesIO(img_bytes)
        st.image(img_stream)
        st.write(self._get_media_type(media["MediaType"]))
        st.write(self._get_priority(media["Priority"]))
        media_name = media["Name"]
        if show_highlight_button and st.button(
                "Highlight media",
                key=f"show_not_started_media_{media_name}"
        ):
            if media_name != st.session_state.get("media_name_to_be_highlighted"):
                st.session_state["media_name_to_be_highlighted"] = media_name
                st.session_state["media_to_be_highlighted"] = medias[media_name]
                st.rerun()

    def _show_finished_media(self, media: dict, medias: dict, show_highlight_button: bool = True):
        img_bytes = base64.b64decode(media["CoverImg"])
        img_stream = BytesIO(img_bytes)
        st.image(img_stream)
        st.write(self._get_media_type(media["MediaType"]))
        st.write(self._get_priority(media["Priority"]))
        finished_date = self._get_date(media["FinishedDroppedDate"])
        st.write(finished_date if finished_date is not None else "No finished date")
        st.write(self._get_stars(media["Stars"]))
        media_name = media["Name"]
        if show_highlight_button and st.button(
                "Highlight media",
                key=f"show_finished_media_{media_name}"
        ):
            if media_name != st.session_state.get("media_name_to_be_highlighted"):
                st.session_state["media_name_to_be_highlighted"] = media_name
                st.session_state["media_to_be_highlighted"] = medias[media_name]
                st.rerun()

    def _show_dropped_media(self, media: dict, medias: dict, show_highlight_button: bool = True):
        img_bytes = base64.b64decode(media["CoverImg"])
        img_stream = BytesIO(img_bytes)
        st.image(img_stream)
        st.write(self._get_media_type(media["MediaType"]))
        st.write(self._get_priority(media["Priority"]))
        dropped_date = self._get_date(media["FinishedDroppedDate"])
        st.write(dropped_date if dropped_date is not None else "No dropped date")
        st.write(self._get_stars(media["Stars"]))
        media_name = media["Name"]
        if show_highlight_button and st.button(
                "Highlight media",
                key=f"show_dropped_media_{media_name}"
        ):
            if media_name != st.session_state.get("media_name_to_be_highlighted"):
                st.session_state["media_name_to_be_highlighted"] = media_name
                st.session_state["media_to_be_highlighted"] = medias[media_name]
                st.rerun()

    def _show_highlighted_media(self, media: dict):
        media_properties_col, media_name_col, media_commentary_col = st.columns([0.25, 0.2, 0.55])
        with media_properties_col:
            # Name
            st.markdown(
                f"<h1 style='text-align: center; font-size: 44px'>{media['Name']}</h1>",
                unsafe_allow_html=True,
            )

            # Image
            img_bytes = base64.b64decode(media["CoverImg"])
            img_stream = BytesIO(img_bytes)
            st.image(img_stream, width=250)

            # Properties
            st.link_button("📺 View on site", url=media["URL"])
            st.markdown(f'**Type**: {self._get_media_type(media["MediaType"])}')
            st.markdown(f'**Priority**: {self._get_priority(media["Priority"])}')
            st.markdown(f'**Status**: {self._get_status(media["Status"])}')
            st.markdown(f'**Stars**: {self._get_stars(media["Stars"])}')

            # Dates
            release_date = self._get_date(media["ReleaseDate"])
            st.markdown(f"**Release date**: {release_date}" if release_date is not None else "**No release date.**")
            started_date = self._get_date(media["FinishedDroppedDate"])
            st.markdown(f"**Started date**: {started_date}" if started_date is not None else "**No started date.**")
            dropped_date = self._get_date(media["FinishedDroppedDate"])
            st.markdown(f"**Dropped/Finished date**: {dropped_date}" if dropped_date is not None else "**No dropped/finished date.**")

            # Genres
            base_genres_html = """
                <a href="https://www.imdb.com/search/title/?genres={}" target="_blank" style="text-decoration: none; color: white;">
                    <span>{}</span>
                </a>
            """
            media["Genres"] = [base_genres_html.format(genre.replace(" ", "%20"), genre) for genre in media["Genres"]]
            tagger_component(
                "<strong>Genres:</strong>",
                media["Genres"],
                self._get_tag_colors(len(media["Genres"]))
            )
            # Staff
            base_staff_html = """
                <a href="https://www.imdb.com/find/?q={}" target="_blank" style="text-decoration: none; color: white;">
                    <span>{}</span>
                </a>
            """
            media["Staff"] = [base_staff_html.format(staff.replace(" ", "%20"), staff) for staff in media["Staff"]]
            tagger_component(
                "<strong>Staff:</strong>",
                media["Staff"],
                self._get_tag_colors(len(media["Staff"]))
            )

        with media_commentary_col:
            st.markdown(media["Commentary"])

    def _get_date(self, date_str: str):
        if date_str == "0001-01-01T00:00:00Z":
            return None
        else:
            return datetime.strptime(date_str, "%Y-%m-%dT%H:%M:%SZ").strftime("%B %d, %Y")

    def _get_priority(self, priority: int | str):
        correct_priority = self._media_priority_options.get(priority, None)
        if correct_priority is None:
            media_priority_options = {value: key for key, value in self._media_priority_options.items()}
            correct_priority = media_priority_options[priority]

        return correct_priority

    def _get_status(self, status: int | str):
        correct_status = self._media_status_options.get(status, None)
        if correct_status is None:
            media_status_options = {value: key for key, value in self._media_status_options.items()}
            correct_status = media_status_options[status]

        return correct_status

    def _get_media_type(self, media_type: int | str):
        correct_media_type = self._media_type_options.get(media_type, None)
        if correct_media_type is None:
            media_type_options = {value: key for key, value in self._media_type_options.items()}
            correct_media_type = media_type_options[media_type]

        return correct_media_type

    def _get_stars(self, stars: int | str):
        correct_stars = self._media_stars_options.get(stars, None)
        if correct_stars is None:
            media_stars_options = {value: key for key, value in self._media_stars_options.items()}
            correct_stars = media_stars_options[stars]

        return correct_stars

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

    def show_medias(self, cols_list: list, medias: dict, show_media_func):
        """Show medias in expanders in the cols_list columns.

        Args:
            cols_list (list): A list of streamlit.columns.
            medias (dict): The key is the media name, and the value is a dict with the media properties.
            show_media_func: A function that expects a media and medias dict and show the media.
        """
        col_index = 0
        for name, media in medias.items():
            if col_index == len(cols_list):
                col_index = 0
            with cols_list[col_index]:
                with st.expander(name, True):
                    show_media_func(media, medias)
            col_index += 1

    def add_media(self):
        st.header("Add a media")
        with st.form("add_media_to_medias_tracker_database", clear_on_submit=True):
            media_url = st.text_input(
                "Media URL",
                key="add_media_to_medias_tracker_database_media_url",
                placeholder="https://www.imdb.com/title/tt0816692/",
            )

            selected_media_type = st.selectbox(
                "Type",
                options=self._media_type_options.values(),
                key="add_media_to_medias_tracker_database_media_type",
            )

            selected_media_priority = st.selectbox(
                "Priority",
                options=self._media_priority_options.values(),
                key="add_media_to_medias_tracker_database_media_priority",
            )

            selected_media_status = st.selectbox(
                "Status",
                options=self._media_status_options.values(),
                key="add_media_to_medias_tracker_database_media_status",
            )

            selected_media_stars = st.selectbox(
                "Stars",
                options=self._media_stars_options.values(),
                key="add_media_to_medias_tracker_database_media_stars",
            )

            st.divider()
            media_started_date = st.date_input(
                "📅 Started playing date",
                key="add_media_to_medias_tracker_database_media_started_date",
            )
            no_media_started_date = st.checkbox(
                "I don't know the started date",
                value=True,
                key="add_media_to_medias_tracker_database_media_no_started_date",
            )
            st.divider()
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
                media_type = self._get_media_type(selected_media_type)
                media_priority = self._get_priority(selected_media_priority)
                media_status = self._get_status(selected_media_status)
                media_stars = self._get_stars(selected_media_stars)
                if no_media_started_date:
                    media_started_date = None
                if no_media_finished_dropped_date:
                    media_finished_dropped_date = None

                media_properties = {
                    "url": media_url,
                    "media_type": media_type,
                    "priority": media_priority,
                    "status": media_status,
                    "stars": media_stars,
                    "started_date": media_started_date,
                    "finished_dropped_date": media_finished_dropped_date,
                    "commentary": media_commentary
                }

                self.api_client.add_media(media_properties)

                st.success("Requested media page")


    def _show_update_media(self):
        medias = self.api_client.get_all_medias()

        # Select media to update
        selected_media = st.selectbox(
            "Select a media to update",
            options=medias.keys(),
            index=None,
            placeholder="Choose a media",
            key="update_media_on_medias_tracker_database_select_media_to_update"
        )
        st.title("")
        if selected_media is not None:
            media = medias[selected_media].copy()

            media["MediaType"] = self._get_media_type(media["MediaType"])
            media["Status"] = self._get_status(media["Status"])
            media["Priority"] = self._get_priority(media["Priority"])
            media["Stars"] = self._get_stars(media["Stars"])

            release_date = datetime.strptime(media["ReleaseDate"], "%Y-%m-%dT%H:%M:%SZ")
            media["ReleaseDate"] = date(release_date.year, release_date.month, release_date.day)

            started_date = datetime.strptime(media["StartedDate"], "%Y-%m-%dT%H:%M:%SZ")
            media["StartedDate"] = date(started_date.year, started_date.month, started_date.day)

            finished_dropped_date = datetime.strptime(media["FinishedDroppedDate"], "%Y-%m-%dT%H:%M:%SZ")
            media["FinishedDroppedDate"] = date(
                finished_dropped_date.year,
                finished_dropped_date.month,
                finished_dropped_date.day
            )

            # Show update form
            edit_form_container = st.container()
            with st.expander("Update commentary"):
                edited_commentary = self.show_markdown_editor(media["Commentary"])

            with edit_form_container:
                with st.form("update_media_on_medias_tracker_database"):
                    column_config = {
                        "MediaType": st.column_config.SelectboxColumn(
                            options=self._media_type_options.values(),
                            required=True
                        ),
                        "Priority": st.column_config.SelectboxColumn(
                            options=self._media_priority_options.values(),
                            required=True
                        ),
                        "Status": st.column_config.SelectboxColumn(
                            options=self._media_status_options.values(),
                            required=True
                        ),
                        "Stars": st.column_config.SelectboxColumn(
                            options=self._media_stars_options.values(),
                            required=True
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
                        "MediaType",
                        "Priority",
                        "Status",
                        "Stars",
                        "StartedDate",
                        "FinishedDroppedDate",
                        "ReleaseDate"
                    )
                    updated_media = st.data_editor(
                        [media],
                        use_container_width=True,
                        column_order=column_order,
                        column_config=column_config,
                        key="update_media_on_medias_tracker_database_data_editor",
                    )[0]
                    st.write("")

                    if st.form_submit_button():
                        media_type = self._get_media_type(updated_media["MediaType"])
                        media_priority = self._get_priority(updated_media["Priority"])
                        media_status = self._get_status(updated_media["Status"])
                        media_stars = self._get_stars(updated_media["Stars"])

                        media_properties = {
                            "name": media["Name"],
                            "media_type": media_type,
                            "priority": media_priority,
                            "status": media_status,
                            "stars": media_stars,
                            "started_date": str(updated_media["StartedDate"]) if updated_media["StartedDate"] is not None else "",
                            "finished_dropped_date": str(updated_media["FinishedDroppedDate"]) if updated_media["FinishedDroppedDate"] is not None else "",
                            "release_date": str(updated_media["ReleaseDate"]) if updated_media["ReleaseDate"] is not None else "",
                            "commentary": edited_commentary
                        }

                        self.api_client.update_media(media_properties)

                        st.success("Media update requested")
                        st.session_state["update_all_medias"] = True

    def show_markdown_editor(self, original_body: str):
        markdown_viewer_col, markdown_editor_col = st.columns(2)
        with markdown_editor_col:
            edited_body = st.text_area(
                label="will be collapsed",
                value=original_body,
                label_visibility="collapsed",
                key="update_media_on_medias_tracker_database_markdown_editor"
            )
        with markdown_viewer_col:
            st.markdown(edited_body)

        return edited_body


page = MediasTrackerPage()
page.show()
