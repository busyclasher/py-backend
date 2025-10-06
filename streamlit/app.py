import os
from typing import Any, Dict, List

import requests
import streamlit as st

API_BASE = os.getenv("API_BASE", "http://localhost:8080")

st.set_page_config(page_title="StoryForge Reader", layout="wide")
st.title("StoryForge Narrative Viewer")


def list_stories() -> List[Dict[str, Any]]:
    response = requests.get(f"{API_BASE}/api/stories", timeout=10)
    response.raise_for_status()
    return response.json()


def get_story(story_id: str) -> Dict[str, Any]:
    response = requests.get(f"{API_BASE}/api/stories/{story_id}", timeout=10)
    response.raise_for_status()
    return response.json()


stories_data = list_stories()
story_options = {"-- Select --": None}
for item in stories_data:
    story_options[item["title"]] = item["id"]

selection = st.sidebar.selectbox("Choose a story", options=list(story_options.keys()))
selected_id = story_options.get(selection)

if selected_id:
    story = get_story(selected_id)
    st.subheader(story["title"])
    st.caption(story["description"])

    for block in story["blocks"]:
        st.markdown(f"### Block #{block['position'] + 1} — {block['type'].title()}")
        if block["type"] == "markdown":
            st.markdown(block["source"])
        else:
            st.code(block["source"], language=block.get("language") or "text")
        if block["outputs"]:
            with st.expander("Outputs", expanded=False):
                for output in block["outputs"]:
                    st.code(output["data"], language="text")

    st.divider()
    st.subheader("Comments")
    for comment in story["comments"]:
        st.markdown(f"**{comment['author']}** — {comment['createdAt']}")
        st.info(comment["body"])
else:
    st.info("Create a story from the Next.js workspace to see it here.")
