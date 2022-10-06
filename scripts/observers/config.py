import os
import toml
import sys
import shutil
from pathlib import Path
from git import Repo
from dotenv import load_dotenv

load_dotenv()

observer_dir = str(os.getenv('OBSERVER_DIR'))
METACHAIN = 4294967295


def update_toml_indexer(path, shard_id):
    # prefs.toml
    data1 = toml.load(path+"/prefs.toml")
    data1['config']['web-socket']['server-url'] = str(shard_id)
    if shard_id != METACHAIN:
        data1['config']['web-socket']['server-url'] = "localhost:" + str(22111+shard_id)
    else:
        data1['config']['web-socket']['server-url'] = "localhost:" + str(22111+10)
    f = open(path+"/prefs.toml", 'w')
    toml.dump(data1, f)
    f.close()


def update_toml_node(path, shard_id):
    # prefs.toml
    data1 = toml.load(path+"/prefs.toml")
    data1['Preferences']['DestinationShardAsObserver'] = str(shard_id)
    f = open(path+"/prefs.toml", 'w')
    toml.dump(data1, f)
    f.close()

    # external.toml
    data2 = toml.load(path+"/external.toml")
    data2['WebSocketConnector']['Enabled'] = True
    if shard_id != METACHAIN:
        data2['WebSocketConnector']['URL'] = "localhost:" + str(22111+shard_id)
    else:
        data2['WebSocketConnector']['URL'] = "localhost:" + str(22111+10)
    f = open(path+"/external.toml", 'w')
    toml.dump(data2, f)
    f.close()


def prepare_observer(shard_id):
    current_observer = observer_dir + str(shard_id)
    working_dir_observer = working_dir + current_observer
    os.mkdir(working_dir_observer)
    os.mkdir(working_dir_observer + "/indexer")
    os.mkdir(working_dir_observer + "/node")

    node_config = working_dir_observer + "/node/config"
    indexer_config = working_dir_observer + "/indexer/config"

    shutil.copytree(config_folder, node_config)
    shutil.copytree("../../cmd/elasticindexer/config", indexer_config)
    shutil.copyfile("../../cmd/elasticindexer/elasticindexer", working_dir_observer+"/indexer/elasticindexer")
    os.system("chmod +x " + working_dir_observer+"/indexer/elasticindexer")
    shutil.copyfile(working_dir + "/elrond-go/cmd/node/node",  working_dir_observer+"/node/node")
    os.system("chmod +x " + working_dir_observer+"/node/node")
    update_toml_node(node_config, shard_id)
    update_toml_indexer(indexer_config, shard_id)


working_dir = str(Path.home()) + str(os.getenv('WORKING_DIRECTORY'))
try:
    os.makedirs(working_dir)
except FileExistsError:
    print("working directory already exits")
    print("use `python3 clean.py` command first")
    sys.exit()

# CLONE elrond-config
print("cloning elrond-config....")
config_folder = working_dir + "/config"
if not os.path.isdir(config_folder):
    Repo.clone_from(os.getenv('ELROND_CONFIG_URL'), config_folder)

repo_cfg = Repo(config_folder)
repo_cfg.git.checkout(os.getenv('ELROND_CONFIG_BRANCH'))

# CLONE elrond-go
print("cloning elrond-go....")
elrond_go_folder = working_dir + "/elrond-go"
if not os.path.isdir(elrond_go_folder):
    Repo.clone_from(os.getenv('ELROND_GO_URL'), elrond_go_folder)

repo_elrond_go = Repo(elrond_go_folder)
repo_elrond_go.git.checkout(os.getenv('ELROND_GO_BRANCH'))

# build binary elrond-go
print("building node...")
current_dir = os.getcwd()
os.chdir(elrond_go_folder + "/cmd/node")
os.system("go build")
os.chdir(current_dir)

# build binary indexer
print("building indexer...")
os.chdir("../../cmd/elasticindexer")
os.system("go build")

os.chdir(current_dir)

# prepare observers
print("preparing config...")
prepare_observer(0)
prepare_observer(1)
prepare_observer(2)
prepare_observer(4294967295)
