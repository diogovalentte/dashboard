import requests
import streamlit as st

from dashboard.api.client import get_api_client

st.set_page_config(
    page_title="Maintenance",
    page_icon="🛠️",
    layout="wide",
)


class MaintenancePage:
    def __init__(self) -> None:
        self.api_client = get_api_client()

    def sidebar(self):
        self.create_service_status_widget("Dashboard", "http://localhost:8501/healthz")
        self.create_service_status_widget("Backend API", "http://api:8080/v1/health")
        self.create_service_status_widget("Geckodriver", "http://api:8085/status")

        st.sidebar.divider()

        self.api_client.show_all_jobs_updating()

    def show(self):
        st.markdown(
            "<h1 style='text-align: center; font-size: 75px'>Maintenance</h1>",
            unsafe_allow_html=True,
        )
        st.title("")
        st.title("")

        col1, _ = st.columns(2)

        # Jobs
        with col1:
            st.header("Jobs")
            if st.button(
                label="Delete all jobs",
                key="delete_all_jobs_from_api",
                type="primary",
            ):
                self.api_client.delete_all_jobs()
                st.success("Jobs deleted")

        self.sidebar()

    def create_service_status_widget(
        self, service_name: str, url: str, expected_status_code: int = 200
    ):
        res = requests.get(url)
        status_code = res.status_code

        if status_code == expected_status_code:
            st.sidebar.metric(
                label="Health status",
                value=service_name,
                delta="Health",
                label_visibility="collapsed",
            )
        else:
            st.sidebar.metric(
                label="Health status",
                value=service_name,
                delta="-Unhealth",
                label_visibility="collapsed",
            )


page = MaintenancePage()
page.show()
