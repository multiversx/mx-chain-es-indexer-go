import os

from dotenv import load_dotenv


def main():
    load_dotenv()

    is_indexer_server = os.getenv('INDEXER_BINARY_SERVER')

    os.system("screen -X -S proxy quit")
    os.system("screen -X -S seednode quit")

    os.system("screen -X -S obs4294967295 quit")
    if not is_indexer_server:
        os.system("screen -X -S indexer4294967295 quit")

    num_of_shards = int(os.getenv('NUM_OF_SHARDS'))
    for shard_id in range(num_of_shards):
        os.system(f'screen -X -S obs{shard_id} quit')
        if not is_indexer_server:
            os.system(f'screen -X -S indexer{shard_id} quit')

    if is_indexer_server:
        os.system("screen -X -S indexerserver quit")

    print("done")


if __name__ == "__main__":
    main()
