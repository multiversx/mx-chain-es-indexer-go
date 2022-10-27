import os
import stat
import subprocess
import toml
import sys
import shutil
from pathlib import Path
from git import Repo
from dotenv import load_dotenv
from utils import get_working_dir

METACHAIN = 4294967295
WS_PORT_BASE = 22111
WS_METACHAIN_PORT = WS_PORT_BASE + 50


def update_toml_indexer(path, shard_id):
    # prefs.toml
    path_prefs = Path(path / "prefs.toml")
    prefs_data = toml.load(str(path_prefs))
    prefs_data['config']['web-socket']['server-url'] = str(shard_id)
    if shard_id != METACHAIN:
        prefs_data['config']['web-socket']['server-url'] = "localhost:" + str(WS_PORT_BASE + shard_id)
    else:
        prefs_data['config']['web-socket']['server-url'] = "localhost:" + str(WS_METACHAIN_PORT)
    f = open(path_prefs, 'w')
    toml.dump(prefs_data, f)
    f.close()


def update_toml_node(path, shard_id):
    # prefs.toml
    path_prefs = Path(path / "prefs.toml")
    prefs_data = toml.load(str(path_prefs))
    prefs_data['Preferences']['DestinationShardAsObserver'] = str(shard_id)
    f = open(path_prefs, 'w')
    toml.dump(prefs_data, f)
    f.close()

    # external.toml
    path_external = Path(path / "external.toml")
    external_data = toml.load(str(path_external))
    external_data['WebSocketConnector']['Enabled'] = True
    if shard_id != METACHAIN:
        external_data['WebSocketConnector']['URL'] = "localhost:" + str(WS_PORT_BASE + shard_id)
    else:
        external_data['WebSocketConnector']['URL'] = "localhost:" + str(WS_METACHAIN_PORT)
    f = open(path_external, 'w')
    toml.dump(external_data, f)
    f.close()


def prepare_observer(shard_id, working_dir, config_folder):
    observer_dir = str(os.getenv('OBSERVER_DIR_PREFIX'))
    current_observer = observer_dir + str(shard_id)
    working_dir_observer = Path(working_dir / current_observer)
    os.mkdir(working_dir_observer)
    os.mkdir(Path(working_dir_observer / "indexer"))
    os.mkdir(Path(working_dir_observer / "node"))

    node_config = Path(working_dir_observer / "node" / "config")
    indexer_config = Path(working_dir_observer / "indexer" / "config")

    shutil.copytree(config_folder, node_config)
    shutil.copytree("../../cmd/elasticindexer/config", indexer_config)
    shutil.copyfile("../../cmd/elasticindexer/elasticindexer", Path(working_dir_observer / "indexer/elasticindexer"))

    elastic_indexer_exec = Path(working_dir_observer / "indexer/elasticindexer")
    st = os.stat(elastic_indexer_exec)
    os.chmod(elastic_indexer_exec, st.st_mode | stat.S_IEXEC)

    shutil.copyfile(Path(working_dir / "elrond-go/cmd/node/node"), Path(working_dir_observer / "node/node"))

    node_exec_path = Path(working_dir_observer / "node/node")
    st = os.stat(node_exec_path)
    os.chmod(node_exec_path, st.st_mode | stat.S_IEXEC)

    update_toml_node(node_config, shard_id)
    update_toml_indexer(indexer_config, shard_id)


def main():
    load_dotenv()
    working_dir = get_working_dir()
    try:
        os.makedirs(working_dir)
    except FileExistsError:
        print(f"working directory {working_dir} already exists")
        print("use `python3 clean.py` command first")
        sys.exit()

    # CLONE elrond-config
    print("cloning elrond-config....")
    config_folder = Path(working_dir / "config")
    if not os.path.isdir(config_folder):
        Repo.clone_from(os.getenv('ELROND_CONFIG_URL'), config_folder)

    repo_cfg = Repo(config_folder)
    repo_cfg.git.checkout(os.getenv('ELROND_CONFIG_BRANCH'))

    # CLONE elrond-go
    print("cloning elrond-go....")
    elrond_go_folder = Path(working_dir / "elrond-go")
    if not os.path.isdir(elrond_go_folder):
        Repo.clone_from(os.getenv('ELROND_GO_URL'), elrond_go_folder)

    repo_elrond_go = Repo(elrond_go_folder)
    repo_elrond_go.git.checkout(os.getenv('ELROND_GO_BRANCH'))

    # build binary elrond-go
    print("building node...")
    subprocess.check_call(["go", "build"], cwd=Path(elrond_go_folder / "cmd/node"))

    # build binary indexer
    print("building indexer...")
    subprocess.check_call(["go", "build"], cwd="../../cmd/elasticindexer")

    # prepare observers
    print("preparing config...")
    prepare_observer(0, working_dir, config_folder)
    prepare_observer(1, working_dir, config_folder)
    prepare_observer(2, working_dir, config_folder)
    prepare_observer(METACHAIN, working_dir, config_folder)


if __name__ == "__main__":
    main()
