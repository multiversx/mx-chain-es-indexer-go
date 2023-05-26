import stat
import subprocess
import toml
import shutil
from git import Repo
from dotenv import load_dotenv
from utils import *


def update_toml_indexer(path, shard_id):
    # prefs.toml
    is_indexer_server = os.getenv('INDEXER_BINARY_SERVER')
    path_prefs = path / "prefs.toml"
    prefs_data = toml.load(str(path_prefs))

    port = WS_PORT_BASE + shard_id
    meta_port = WS_METACHAIN_PORT
    if is_indexer_server:
        port = WS_PORT_BASE
        meta_port = WS_PORT_BASE
        prefs_data['config']['web-socket']['mode'] = "server"

    if shard_id != METACHAIN:
        prefs_data['config']['web-socket']['url'] = "localhost:" + str(port)
    else:
        prefs_data['config']['web-socket']['url'] = "localhost:" + str(meta_port)
    prefs_data['config']['web-socket']['data-marshaller-type'] = str(os.getenv('WS_MARSHALLER_TYPE'))

    f = open(path_prefs, 'w')
    toml.dump(prefs_data, f)
    f.close()


def update_toml_node(path, shard_id):
    # prefs.toml
    path_prefs = path / "prefs.toml"
    prefs_data = toml.load(str(path_prefs))
    prefs_data['Preferences']['DestinationShardAsObserver'] = str(shard_id)
    f = open(path_prefs, 'w')
    toml.dump(prefs_data, f)
    f.close()

    # config.toml
    num_of_shards = int(os.getenv('NUM_OF_SHARDS'))
    path_config = path / "config.toml"
    config_data = toml.load(path_config)
    config_data['DbLookupExtensions']['Enabled'] = True
    config_data['EpochStartConfig']['RoundsPerEpoch'] = 20
    config_data['GeneralSettings']['GenesisMaxNumberOfShards'] = num_of_shards
    f = open(path_config, 'w')
    toml.dump(config_data, f)
    f.close()

    # external.toml
    path_external = path / "external.toml"
    external_data = toml.load(str(path_external))
    external_data['HostDriverConfig']['Enabled'] = True

    port = WS_PORT_BASE + shard_id
    meta_port = WS_METACHAIN_PORT

    is_indexer_server = os.getenv('INDEXER_BINARY_SERVER')
    if is_indexer_server:
        external_data['HostDriverConfig']['IsServer'] = False
        port = WS_PORT_BASE
        meta_port = WS_PORT_BASE

    if shard_id != METACHAIN:
        external_data['HostDriverConfig']['URL'] = "localhost:" + str(port)
    else:
        external_data['HostDriverConfig']['URL'] = "localhost:" + str(meta_port)

    external_data['HostDriverConfig']['MarshallerType'] = str(os.getenv('WS_MARSHALLER_TYPE'))
    f = open(path_external, 'w')
    toml.dump(external_data, f)
    f.close()


def prepare_observer(shard_id, working_dir, config_folder):
    observer_dir = str(os.getenv('OBSERVER_DIR_PREFIX'))
    current_observer = observer_dir + str(shard_id)
    working_dir_observer = working_dir / current_observer
    os.mkdir(working_dir_observer)
    os.mkdir(working_dir_observer / "indexer")
    os.mkdir(working_dir_observer / "node")

    node_config = working_dir_observer / "node" / "config"
    indexer_config = working_dir_observer / "indexer" / "config"

    shutil.copytree(config_folder, node_config)
    shutil.copytree("../../cmd/elasticindexer/config", indexer_config)
    shutil.copyfile("../../cmd/elasticindexer/elasticindexer", working_dir_observer / "indexer/elasticindexer")

    elastic_indexer_exec = Path(working_dir_observer / "indexer/elasticindexer")
    st = os.stat(elastic_indexer_exec)
    os.chmod(elastic_indexer_exec, st.st_mode | stat.S_IEXEC)

    shutil.copyfile(working_dir / "mx-chain-go/cmd/node/node", working_dir_observer / "node/node")

    node_exec_path = working_dir_observer / "node/node"
    st = os.stat(node_exec_path)
    os.chmod(node_exec_path, st.st_mode | stat.S_IEXEC)

    update_toml_node(node_config, shard_id)
    update_toml_indexer(indexer_config, shard_id)


def prepare_indexer_server(meta_id, working_dir):
    is_indexer_server = os.getenv('INDEXER_BINARY_SERVER')
    if not is_indexer_server:
        return

    current_observer = str(os.getenv('OBSERVER_DIR_PREFIX')) + str(meta_id)
    working_dir_observer = working_dir / current_observer
    shutil.copytree(working_dir_observer / "indexer", working_dir / "indexer")


def generate_new_config(working_dir):
    mx_chain_go_folder = working_dir / "mx-chain-go" / "scripts" / "testnet"
    num_of_shards = str(os.getenv('NUM_OF_SHARDS'))

    with open(mx_chain_go_folder / "local.sh", "w") as file:
        file.write(f'export SHARDCOUNT={num_of_shards}\n')
        file.write("export SHARD_VALIDATORCOUNT=1\n")
        file.write("export SHARD_OBSERVERCOUNT=0\n")
        file.write("export SHARD_CONSENSUS_SIZE=1\n")
        file.write("export META_VALIDATORCOUNT=1\n")
        file.write("export META_OBSERVERCOUNT=0\n")
        file.write("export META_CONSENSUS_SIZE=1\n")
        file.write('export LOGLEVEL="*:DEBUG"\n')
        file.write('export OBSERVERS_ANTIFLOOD_DISABLE=0\n')
        file.write('export USETMUX=0\n')
        file.write('export USE_PROXY=0\n')


def clone_mx_chain_go(working_dir):
    print("cloning mx-chain-go....")
    mx_chain_go_folder = working_dir / "mx-chain-go"
    if not os.path.isdir(mx_chain_go_folder):
        Repo.clone_from(os.getenv('NODE_GO_URL'), mx_chain_go_folder)

    repo_mx_chain_go = Repo(mx_chain_go_folder)
    repo_mx_chain_go.git.checkout(os.getenv('NODE_GO_BRANCH'))


def clone_dependencies(working_dir):
    print("cloning dependencies")
    mx_chain_deploy_folder = working_dir / "mx-chain-deploy-go"
    if not os.path.isdir(mx_chain_deploy_folder):
        Repo.clone_from(os.getenv('MX_CHAIN_DEPLOY_GO_URL'), mx_chain_deploy_folder)

    mx_chain_proxy_folder = working_dir / "mx-chain-proxy-go"
    if not os.path.isdir(mx_chain_proxy_folder):
        Repo.clone_from(os.getenv('MX_CHAIN_PROXY_URL'), mx_chain_proxy_folder)


def prepare_seed_node(working_dir):
    print("preparing seed node")
    seed_node = Path.home() / "MultiversX/testnet/seednode"
    shutil.copytree(seed_node, working_dir / "seednode")

    mx_chain_go_folder = working_dir / "mx-chain-go"
    subprocess.check_call(["go", "build"], cwd=mx_chain_go_folder / "cmd/seednode")

    seed_node_exec = mx_chain_go_folder / "cmd/seednode/seednode"
    shutil.copyfile(seed_node_exec, working_dir / "seednode/seednode")

    st = os.stat(working_dir / "seednode/seednode")
    os.chmod(working_dir / "seednode/seednode", st.st_mode | stat.S_IEXEC)


def prepare_proxy(working_dir):
    print("preparing proxy")
    mx_chain_proxy_go_folder = working_dir / "mx-chain-proxy-go"
    subprocess.check_call(["go", "build"], cwd=mx_chain_proxy_go_folder / "cmd/proxy")

    mx_chain_proxy_go_binary_folder = mx_chain_proxy_go_folder / "cmd/proxy"
    st = os.stat(mx_chain_proxy_go_binary_folder / "proxy")
    os.chmod(mx_chain_proxy_go_binary_folder / "proxy", st.st_mode | stat.S_IEXEC)

    # config.toml
    path_config = mx_chain_proxy_go_binary_folder / "config/config.toml"
    config_data = toml.load(str(path_config))

    proxy_port = int(os.getenv('PROXY_PORT'))
    config_data['GeneralSettings']['ServerPort'] = proxy_port
    del config_data['Observers']
    del config_data['FullHistoryNodes']

    config_data['Observers'] = []

    observers_start_port = int(os.getenv('OBSERVERS_START_PORT'))
    meta_observer = {
        'ShardId': 4294967295,
        'Address': f'http://127.0.0.1:{observers_start_port}',
    }
    config_data['Observers'].append(meta_observer)

    num_of_shards = int(os.getenv('NUM_OF_SHARDS'))
    for shardID in range(num_of_shards):
        shard_observer_port = observers_start_port + shardID + 1
        meta_observer = {
            'ShardId': shardID,
            'Address': f'http://127.0.0.1:{shard_observer_port}',
        }
        config_data['Observers'].append(meta_observer)

    f = open(path_config, 'w')
    toml.dump(config_data, f)
    f.close()


def generate_config_for_local_testnet(working_dir):
    mx_chain_local_testnet_scripts = working_dir / "mx-chain-go/scripts/testnet"
    subprocess.check_call(["./clean.sh"], cwd=mx_chain_local_testnet_scripts)
    subprocess.check_call(["./config.sh"], cwd=mx_chain_local_testnet_scripts)

    config_folder = Path.home() / "MultiversX/testnet/node/config"
    os.rename(config_folder / "config_validator.toml", config_folder / "config.toml")
    shutil.copytree(config_folder, working_dir / "config")


def main():
    load_dotenv()
    working_dir = get_working_dir()
    try:
        os.makedirs(working_dir)
    except FileExistsError:
        print("something")
        print(f"working directory {working_dir} already exists")
        print("use `python3 clean.py` command first")
        sys.exit()

    num_of_shards = int(os.getenv('NUM_OF_SHARDS'))
    check_num_of_shards(num_of_shards)

    # clone mx-chain-go
    clone_mx_chain_go(working_dir)
    # clone dependencies
    clone_dependencies(working_dir)
    # generate configs
    generate_new_config(working_dir)
    generate_config_for_local_testnet(working_dir)
    # prepare seednode
    prepare_seed_node(working_dir)
    # prepare proxy
    prepare_proxy(working_dir)

    # build binary mx-chain-go
    print("building node...")
    mx_chain_go_folder = working_dir / "mx-chain-go"
    flags = '-gcflags="all=-N -l"'
    subprocess.check_call(["go", "build", flags], cwd=mx_chain_go_folder / "cmd/node")

    # build binary indexer
    print("building indexer...")
    subprocess.check_call(["go", "build", flags], cwd="../../cmd/elasticindexer")

    # prepare observers
    config_folder = working_dir / "config"
    print("preparing config...")
    prepare_observer(METACHAIN, working_dir, config_folder)
    prepare_indexer_server(METACHAIN, working_dir)

    for shardID in range(num_of_shards):
        prepare_observer(shardID, working_dir, config_folder)


if __name__ == "__main__":
    main()
