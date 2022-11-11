from dotenv import load_dotenv
from utils import *


def start_observer(shard_id, working_dir):
    current_observer = str(os.getenv('OBSERVER_DIR_PREFIX')) + str(shard_id)
    working_dir_observer = working_dir / current_observer

    current_directory = os.getcwd()
    # start observer
    os.chdir(working_dir_observer / "node")
    command = "./node" + " --log-level *:DEBUG --no-key --log-save"
    os.system("screen -d -m -S obs" + str(shard_id) + " " + command)

    # start indexer
    os.chdir(working_dir_observer / "indexer")
    command = "./elasticindexer" + " --log-level *:DEBUG --log-save"
    os.system("screen -d -m -S indexer" + str(shard_id) + " " + command)

    os.chdir(current_directory)


def main():
    load_dotenv()
    working_dir = get_working_dir()
    if not os.path.exists(working_dir):
        print("working directory folder is missing...you should run first `python3 config.py` command")
        sys.exit()

    print("staring observers and indexers....")

    start_observer(0, working_dir)
    start_observer(1, working_dir)
    start_observer(2, working_dir)
    start_observer(METACHAIN, working_dir)

    print("done")


if __name__ == "__main__":
    main()
