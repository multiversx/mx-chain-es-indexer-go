import os


def main():
    os.system("screen -X -S proxy quit")
    os.system("screen -X -S seednode quit")
    os.system("screen -X -S obs0 quit")
    os.system("screen -X -S obs1 quit")
    os.system("screen -X -S obs2 quit")
    os.system("screen -X -S obs4294967295 quit")
    os.system("screen -X -S indexer0 quit")
    os.system("screen -X -S indexer1 quit")
    os.system("screen -X -S indexer2 quit")
    os.system("screen -X -S indexer4294967295 quit")
    print("done")


if __name__ == "__main__":
    main()
