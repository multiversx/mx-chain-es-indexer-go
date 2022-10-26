import shutil
import os
from pathlib import Path
from dotenv import load_dotenv


def main():
    load_dotenv()
    working_dir = Path(Path.home() / os.getenv('WORKING_DIRECTORY'))

    try:
        shutil.rmtree(working_dir)
        print(f"removed directory: {working_dir}")
    except FileNotFoundError:
        print("nothing to clean")


if __name__ == "__main__":
    main()
