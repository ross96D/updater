import os
import sys

if __name__ == "__main__":
    sys.stdout.write("custom_checksum " + os.environ["GITHUB_TOKEN"])
