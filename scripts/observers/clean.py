import shutil
from dotenv import load_dotenv
from utils import *


def main():
    load_dotenv()
    working_dir = get_working_dir()
    try:
        shutil.rmtree(working_dir)
        print(f"removed directory: {working_dir}")
    except FileNotFoundError:
        print("nothing to clean")


if __name__ == "__main__":
    main()
