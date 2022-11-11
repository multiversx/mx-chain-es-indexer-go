import os
import sys
from pathlib import Path


METACHAIN = 4294967295
WS_PORT_BASE = 22111
WS_METACHAIN_PORT = WS_PORT_BASE + 50


def get_working_dir():
    working_dir_var = os.getenv('WORKING_DIRECTORY')
    if working_dir_var == "":
        print("working directory folder name cannot be empty")
        sys.exit()

    return Path.home() / working_dir_var
