import os
import sys
from pathlib import Path


def get_working_dir():
    working_dir_var = os.getenv('WORKING_DIRECTORY')
    if working_dir_var == "":
        print("working directory folder name cannot be empty")
        sys.exit()

    return Path(Path.home() / working_dir_var)
