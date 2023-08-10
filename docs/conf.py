"""
This file is the configuration file for the Sphinx documentation builder.
See the documentation: http://www.sphinx-doc.org/en/master/config
"""

import json
import os
import pathlib
import sys
import time

# Doc Path
# If extensions (or modules to document with autodoc) are in another directory,
# add these directories to sys.path here. If the directory is relative to the
# documentation root, use os.path.abspath to make it absolute, like shown here.
sys.path.append(os.path.abspath(os.path.dirname(__file__)))

project = "Determined"
html_title = "Determined AI Documentation"
copyright = time.strftime("%Y, Determined AI")
author = "hello@determined.ai"
version = pathlib.Path(__file__).parents[1].joinpath("VERSION").read_text().strip()
release = version
language = "en"

source_suffix = {".rst": "restructuredtext"}
templates_path = ["_templates"]
html_static_path = ["assets", "_static"]
html_css_files = [
    "https://cdn.jsdelivr.net/npm/@docsearch/css@3",
    "styles/determined.css",
]

html_js_files = [
    ("https://cdn.jsdelivr.net/npm/@docsearch/js@3", {"defer": "defer"}),
    ("scripts/docsearch.sbt.js", {"defer": "defer"}),
]


def env_get_outdated(app, env, added, changed, removed):
    return ["index"]


def setup(app):
    app.connect("env-get-outdated", env_get_outdated)


exclude_patterns = [
    "_build",
    "Thumbs.db",
    ".DS_Store",
    "examples",
    "requirements.txt",
    "site",
    "README.md",
    "release-notes/README.md",
]
html_baseurl = "https://docs.determined.ai"  # Base URL for sitemap.
highlight_language = "none"
todo_include_todos = True

# HTML theme settings
html_show_sourcelink = False
html_show_sphinx = False
html_theme = "sphinx_book_theme"
html_favicon = "assets/images/favicon.ico"
html_last_updated_fmt = None
# See https://pradyunsg.me/furo/

# `navbar-logo.html` and `sbt-sidebar-nav.html` come from `sphinx-book-theme`
html_sidebars = {
    "**": [
        "navbar-logo.html",
        "sidebar-version.html",
        "search-field.html",
        "sbt-sidebar-nav.html",
    ],
# to suppress sidebar on home page uncomment this line:
#    "index": [],
}

pygments_style = "sphinx"
pygments_dark_style = "monokai"
html_theme_options = {
    "logo": {
        "image_light": "assets/images/logo-determined-ai.svg",
        "image_dark": "assets/images/logo-determined-ai-white.svg",
    },
    "switcher": {
        "json_url": "https://docs.determined.ai/latest/_static/version-switcher/versions.json",
        "version_match": version,
    },
    "repository_url": "https://github.com/determined-ai/determined",
    "use_repository_button": True,
    "use_download_button": False,
    "use_fullscreen_button": False,
}
html_use_index = True
html_domain_indices = True

extensions = [
    "sphinx_ext_downloads",
    "sphinx.ext.autodoc",
    "sphinx.ext.extlinks",
    "sphinx.ext.intersphinx",
    "sphinx.ext.mathjax",
    "sphinx.ext.napoleon",
    "sphinx_copybutton",
    "sphinx_sitemap",
    "sphinx_reredirects",
    "sphinx_tabs.tabs",
    "myst_parser",
]

myst_extensions = [
    "colon_fence",
]

# Our custom sphinx extension uses this value to decide where to look for
# downloadable files.
dai_downloads_root = os.path.join("site", "downloads")

# sphinx.ext.autodoc configurations.
# See https://www.sphinx-doc.org/en/master/usage/extensions/autodoc.html
autosummary_generate = True
autoclass_content = "class"
autodoc_mock_imports = [
    "mmcv",
    "mmdet",
    "transformers",
    "deepspeed",
    "datasets",
    "analytics",
]

# sphinx-sitemap configurations.
# See https://github.com/jdillard/sphinx-sitemap
# The URLs generated by sphinx-sitemap include both the version number and the
# language by default. We don't use language in the published URL, and we also
# want to encourage the latest version of the docs to be indexed, so only
# include that variant in the sitemap.
sitemap_url_scheme = "latest/{link}"

with open(".redirects/redirects.json") as f:
    redirects = json.load(f)
