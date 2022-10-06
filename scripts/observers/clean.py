import shutil
import os
from pathlib import Path
from dotenv import load_dotenv

load_dotenv()

working_dir = str(Path.home()) + str(os.getenv('WORKING_DIRECTORY'))

try:
    shutil.rmtree(working_dir)
    print("done")
except:
    print('did nothing')
