import os
import sys
from pathlib import Path
from dotenv import load_dotenv

load_dotenv()

working_dir = str(Path.home()) + str(os.getenv('WORKING_DIRECTORY'))
observer_dir = str(os.getenv('OBSERVER_DIR'))


def start_observer(shard_id):
    current_observer = observer_dir + str(shard_id)
    working_dir_observer = working_dir + current_observer

    current_directory = os.getcwd()
    # start observer
    os.chdir(working_dir_observer + "/node")
    command = "./node" + " --log-level *:DEBUG --no-key --log-save"
    # os.system("gnome-terminal 'bash -c \"" + command + ";bash\"'")
    os.system("screen -d -m -S obs" + str(shard_id) + " " + command)
    os.chdir(current_directory)
    # start indexer
    os.chdir(working_dir_observer + "/indexer")
    command = "./elasticindexer" + " --log-level *:DEBUG --log-save"
    # os.system("gnome-terminal 'bash -c \"" + command + ";bash\"'")
    os.system("screen -d -m -S indexer" + str(shard_id) + " " + command)
    os.chdir(current_directory)


if not os.path.exists(working_dir):
    print("working directory folder is missing...you should run first `python3 config.py` command")
    sys.exit()

print("staring observers and indexers....")

start_observer(0)
start_observer(1)
start_observer(2)
start_observer(4294967295)
print("done")
