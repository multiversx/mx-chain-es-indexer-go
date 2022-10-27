import shutil
import os
import sys
from pathlib import Path
from dotenv import load_dotenv


def main():
    load_dotenv()
    working_dir_var = os.getenv('WORKING_DIRECTORY')
    if working_dir_var == "":
        print("working directory folder name cannot be empty")
        sys.exit()

    working_dir = Path(Path.home() / working_dir_var)

    try:
        shutil.rmtree(working_dir)
        print(f"removed directory: {working_dir}")
    except FileNotFoundError:
        print("nothing to clean")


if __name__ == "__main__":
    main()
